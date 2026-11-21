"""Simulador aproximado de torneo a partir de manos-AAAAMMDD.jsonl de Lince.
Uso: python torneo_sim.py manos.jsonl [manos2.jsonl ...]
"""
import json, sys, random, statistics as st
from collections import defaultdict, Counter

AG = {'bet', 'raise', 'allin'}
INI = 10000


def cargar(rutas):
    ms, vistos = [], set()
    for r in rutas:
        for l in open(r, encoding='utf-8'):
            try:
                m = json.loads(l)
            except Exception:
                continue
            if m['id_mano'] in vistos:
                continue
            vistos.add(m['id_mano']); ms.append(m)
    return ms


def pct(a, b):
    return f'{100 * a / b:5.1f}' if b else '    -'


def pctl(v, q):
    v = sorted(v)
    if not v: return float('nan')
    k = (len(v) - 1) * q; i = int(k); f = k - i
    return v[i] + (v[min(i + 1, len(v) - 1)] - v[i]) * f


def ganancias(m):
    """Neto de cada jugador en la mano (reparto - invertido), deduciendo all-ins de los saldos."""
    J = [j['id'] for j in m['jugadores']]; n = len(J); sal = {j['id']: j['saldo'] for j in m['jugadores']}
    cg = m['ciega_grande']; b = m['boton']
    inv = Counter(); calle = Counter()
    sb, bb = J[(b + 1) % n], J[(b + 2) % n]
    if n == 2:
        sb, bb = J[b % 2], J[(b + 1) % 2]
    inv[sb] += cg // 2; inv[bb] += cg; calle[sb] = cg // 2; calle[bb] = cg
    et = 'preflop'; cur = cg; pote = cg // 2 + cg
    for a in (m.get('acciones') or []):
        p = a['id_jugador']; t = a['accion']['tipo']; mo = a['accion'].get('monto', 0)
        if a['etapa'] != et:
            et = a['etapa']; calle = Counter(); cur = 0
        if t == 'allin':
            mo = calle[p] + sal[p] - inv[p]
        if t == 'call': add = cur - calle[p]
        elif t in AG: add = mo - calle[p]
        else: add = 0
        add = min(max(add, 0), sal[p] - inv[p])
        if t in AG and mo > cur: cur = mo
        inv[p] += add; calle[p] += add; pote += add
    rep = Counter({r['id_jugador']: r['monto'] for r in (m.get('repartos') or [])})
    return {p: rep[p] - inv[p] for p in J}, sal


def main():
    ms = cargar(sys.argv[1:])
    if not ms:
        print('sin manos'); return
    YO = ms[0]['yo']
    mesas = defaultdict(list)
    for m in ms:
        mesas[m['id_mano'].rsplit('-mano', 1)[0]].append(m)

    def clave(k):
        # practica-r208-m02 -> (208, 2)
        p = k.split('-')
        try: return (int(p[1][1:]), int(p[2][1:]))
        except Exception: return (0, 0)
    claves = sorted(mesas, key=clave)

    # ---- 1. por mesa
    filas = []
    print(f'Manos {len(ms)} | mesas {len(mesas)}')
    print('\n1) POR MESA (fichas finales = saldo previo de la ultima mano + neto de esa mano; iniciales = saldo en la 1a mano de Lince)')
    print(f'{"ronda-mesa":11}{"manos":>6}{"jug":>4}{"ini":>7}{"final":>7}{"neto":>7}{"puesto":>7}  quiebra  nota')
    for k in claves:
        L = mesas[k]
        L.sort(key=lambda m: int(m['id_mano'].rsplit('-', 1)[1]))
        fin = {}
        for m in L:
            g, sal = ganancias(m)
            for p in sal: fin[p] = sal[p] + g[p]
        # jugadores que ya no aparecen (eliminados) quedan con su ultimo valor calculado (0 si quebraron)
        ini = next(j['saldo'] for j in L[0]['jugadores'] if j['id'] == YO)
        f = fin[YO]
        puesto = 1 + sum(1 for p, v in fin.items() if p != YO and v > f)
        quebro = f <= 0
        nota = []
        if int(L[0]['id_mano'].rsplit('-', 1)[1]) != 1: nota.append('empieza a mitad')
        if len(L) < 50 and not quebro: nota.append('incompleta?')
        filas.append(dict(k=k, n=len(L), jug=len(fin), ini=ini, f=f, neto=f - ini, puesto=puesto, quebro=quebro, parcial=bool(nota)))
        print(f'{k[9:] if k.startswith("practica-") else k:11}{len(L):6}{len(fin):4}{ini:7}{f:7}{f - ini:7}{puesto:7}  {"SI" if quebro else "no":7}  {",".join(nota)}')

    # ---- 2. resumen
    nets = [r['neto'] for r in filas]
    N = len(nets)
    mu = st.mean(nets); se = st.stdev(nets) / N ** .5 if N > 1 else float('nan')
    peor = min(filas, key=lambda r: r['neto'])
    print('\n2) RESUMEN')
    print(f'Mesas {N} | neto medio {mu:.0f} +- {se:.0f} (EE) | mediana {st.median(nets):.0f} | desv {st.pstdev(nets):.0f}')
    print(f'1.o en la mesa {pct(sum(r["puesto"] == 1 for r in filas), N)}% | quiebra {pct(sum(r["quebro"] for r in filas), N)}% | peor mesa {peor["k"]} ({peor["neto"]})')
    print(f'puesto medio {st.mean(r["puesto"] for r in filas):.2f}; distribucion de puestos {dict(sorted(Counter(r["puesto"] for r in filas).items()))}')

    # ---- 3. torneos de 10 mesas
    print('\n3) "TORNEOS" DE 10 MESAS CONSECUTIVAS (bloques disjuntos; fichas reiniciadas, no arrastradas)')
    for b in range(0, N - N % 10, 10):
        bl = filas[b:b + 10]
        q = [i + 1 for i, r in enumerate(bl) if r['quebro']]
        print(f'  bloque {b // 10 + 1} ({bl[0]["k"][9:]}..{bl[-1]["k"][9:]}): suma neto {sum(r["neto"] for r in bl):7} | quiebra {"en ronda " + str(q[0]) if q else "no"} | puesto medio {st.mean(r["puesto"] for r in bl):.2f}')
    if N < 10: print('  (menos de 10 mesas: no hay bloques completos)')
    elif N % 10: print(f'  ({N % 10} mesas sobrantes sin bloque)')

    # ---- 4. bootstrap
    rnd = random.Random(12345)
    R = 20000; sup = 0; acc = []
    for _ in range(R):
        s = [rnd.choice(filas) for _ in range(10)]
        if any(r['quebro'] for r in s): continue
        sup += 1; acc.append(sum(r['neto'] for r in s))
    print('\n4) BOOTSTRAP (20000 replicas, semilla 12345, 10 mesas con reemplazo)')
    print(f'P(sobrevivir 10 rondas sin quiebra) = {100 * sup / R:.1f}%  (con {sum(r["quebro"] for r in filas)} quiebras en {N} mesas)')
    if acc:
        print('Neto acumulado de 10 rondas | sobrevive: ' + ' '.join(f'p{int(q * 100)}={pctl(acc, q):.0f}' for q in (.1, .25, .5, .75, .9)) + f' | media {st.mean(acc):.0f}')
    print('AVISO: aproximacion. Con fichas arrastradas el stack cambia (stack grande = ventaja, corto = presion), las mesas de practica son de 4 y las del torneo de 5-6, '
          'las mesas se suponen independientes e identicamente distribuidas, y los rivales de practica no son los de la clasificatoria. Una mesa con quiebra cuenta como eliminacion.')

    # ---- 5/6. rivales y Lince
    P = defaultdict(Counter); vs = defaultdict(float); comp = Counter()
    tipos = Counter(); extra = Counter()
    for m in ms:
        g, sal = ganancias(m)
        J = list(sal)
        for k in m:
            if k not in ('hora', 'id_mano', 'fase', 'yo', 'mis_cartas', 'ciega_grande', 'boton', 'jugadores', 'acciones', 'comunitarias', 'repartos', 'mostradas', 'descripcion'):
                extra[k] += 1
        cg = m['ciega_grande']; n = len(J); b = m['boton']
        sb, bb = J[(b + 1) % n], J[(b + 2) % n]
        calle = Counter(); calle[sb] = cg // 2; calle[bb] = cg
        et = 'preflop'; cur = cg; inv = Counter(); inv[sb] = cg // 2; inv[bb] = cg
        vpip = set(); pfr = set()
        for a in (m.get('acciones') or []):
            p = a['id_jugador']; t = a['accion']['tipo']; mo = a['accion'].get('monto', 0)
            tipos[t] += 1
            if a['etapa'] != et:
                et = a['etapa']; calle = Counter(); cur = 0
            c = P[p]
            if t == 'allin': mo = calle[p] + sal[p] - inv[p]
            if et == 'preflop':
                if t in ('call', 'bet', 'raise', 'allin'): vpip.add(p)
                if t in AG: pfr.add(p)
            else:
                if cur > calle[p]:
                    c['enf'] += 1
                    if t == 'fold': c['enf_fold'] += 1
                else:
                    c['op'] += 1
                    if t in AG: c['ag'] += 1
            if t == 'call': add = cur - calle[p]
            elif t in AG: add = mo - calle[p]
            else: add = 0
            add = min(max(add, 0), sal[p] - inv[p])
            if t in AG and mo > cur: cur = mo
            inv[p] += add; calle[p] += add
        for p in J:
            P[p]['manos'] += 1; P[p]['vpip'] += p in vpip; P[p]['pfr'] += p in pfr
        if YO in J:
            for p in J:
                if p != YO: comp[p] += 1
            nl = g[YO]
            rep = Counter({r['id_jugador']: r['monto'] for r in (m.get('repartos') or [])})
            w = {p: rep[p] for p in J if p != YO and rep[p] > 0} if nl < 0 else {p: -g[p] for p in J if p != YO and g[p] < 0}
            s = sum(w.values())
            for p, v in w.items():
                if s: vs[p] += nl * v / s
    print('\n5) POR RIVAL (neto de Lince vs rival: aproximacion repartiendo el neto de cada mano entre los rivales que ganaron/perdieron)')
    print(f'{"rival":10}{"manos":>6}{"netoLince":>10}{"VPIP":>7}{"PFR":>7}{"%fold/ap":>9}{"%apuesta":>9}')
    for p in sorted(comp, key=lambda x: vs[x]):
        c = P[p]
        print(f'{p:10}{comp[p]:6}{vs[p]:10.0f}{pct(c["vpip"], c["manos"]):>7}{pct(c["pfr"], c["manos"]):>7}{pct(c["enf_fold"], c["enf"]):>9}{pct(c["ag"], c["op"]):>9}')
    print('%fold/ap = se retira ante apuesta postflop; %apuesta = bet/raise postflop cuando nadie habia apostado en la calle.')
    c = P[YO]
    print(f'\n6) LINCE: manos {c["manos"]} | VPIP {pct(c["vpip"], c["manos"])} | PFR {pct(c["pfr"], c["manos"])} | retiro ante apuesta postflop {pct(c["enf_fold"], c["enf"])} ({c["enf_fold"]}/{c["enf"]}) | % apuesta postflop {pct(c["ag"], c["op"])}')
    print(f'Tipos de accion vistos: {dict(tipos)}; campos extra en registros: {dict(extra) or "ninguno"}')
    print('Timeouts / acciones invalidas: NO detectables en este formato (solo se graba la accion final aceptada; un timeout seria indistinguible de un fold/check automatico).')


main()
