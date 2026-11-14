"""Perfil de cada rival a partir de las manos que grabó Lince.

Lince guarda cada mano completa en datos/manos-AAAAMMDD.jsonl (acciones de
todos, cartas mostradas). Este script las resume por jugador, para ver cómo
juegan los rivales de verdad en el servidor de la organización y comparar con
los bots del kit.

    python bots/lince/pruebas/rivales.py                       # todo lo que hay en bots/lince/datos
    python bots/lince/pruebas/rivales.py ruta/manos-20261114.jsonl
    python bots/lince/pruebas/rivales.py -min 30               # solo rivales con 30 manos o más

Columnas:

    Manos     manos en que recibió cartas
    Entra     % de manos en que puso fichas preflop por decisión propia (VPIP)
    Sube pre  % de manos en que subió preflop (PFR)
    Agresión  apuestas y subidas / (apuestas + subidas + pagos)
    Se retira % de veces que se retiró cuando le apostaron después del flop
    All-in    % de manos en que fue all-in
    Showdown  % de manos en que mostró cartas
    Neto      fichas ganadas o perdidas por mano, en ciegas grandes cada 100 manos
    Quiebras  veces que terminó una mano sin fichas
"""

import argparse
import collections
import glob
import json
import os
import sys

AGRESIVAS = {"bet", "raise", "allin"}


def leer(rutas):
    for ruta in rutas:
        with open(ruta, encoding="utf-8") as f:
            for linea in f:
                linea = linea.strip()
                if not linea:
                    continue
                try:
                    yield json.loads(linea)
                except json.JSONDecodeError:
                    continue  # Una línea a medio escribir (el bot se cortó ahí).


def nuevo():
    return collections.Counter()


def analizar(manos):
    perfiles = collections.defaultdict(nuevo)
    yo = None
    total = 0
    for m in manos:
        total += 1
        yo = m.get("yo") or yo
        cg = m.get("ciega_grande") or 100
        saldos = {j["id"]: j["saldo"] for j in m.get("jugadores", [])}
        puesto = collections.Counter()   # Fichas puestas por cada uno en la mano.
        en_ronda = collections.Counter()  # Apuesta de cada uno en la ronda en curso.
        etapa = None
        entro, subio_pre, allin = set(), set(), set()
        hubo_apuesta = False
        # Las ciegas no figuran como acciones: se reconstruyen por posición.
        ids = [j["id"] for j in m.get("jugadores", []) if j["saldo"] > 0]
        for j in ids:
            perfiles[j]["manos"] += 1
        acciones = m.get("acciones") or []
        for a in acciones:
            jugador, tipo = a["id_jugador"], a["accion"]["tipo"]
            monto = a["accion"].get("monto", 0) or 0
            if a["etapa"] != etapa:
                etapa, hubo_apuesta = a["etapa"], False
                en_ronda.clear()
            p = perfiles[jugador]
            preflop = etapa == "preflop"
            if tipo in AGRESIVAS:
                p["agresivas"] += 1
                if preflop:
                    entro.add(jugador)
                    subio_pre.add(jugador)
                if tipo == "allin":
                    allin.add(jugador)
            elif tipo == "call":
                p["pasivas"] += 1
                if preflop:
                    entro.add(jugador)
            if not preflop and hubo_apuesta and tipo in ("fold", "call", "raise", "allin"):
                p["enfrento"] += 1
                if tipo == "fold":
                    p["se_retiro"] += 1
            if tipo in AGRESIVAS:
                hubo_apuesta = True
            if tipo in ("bet", "raise") and monto:
                en_ronda[jugador] = monto
        for j in entro:
            perfiles[j]["entro"] += 1
        for j in subio_pre:
            perfiles[j]["subio_pre"] += 1
        for j in allin:
            perfiles[j]["allin"] += 1
        for j in (m.get("mostradas") or {}):
            perfiles[j]["showdown"] += 1
        # Neto de la mano: lo que cobró menos lo que puso no se puede sacar
        # solo de las acciones (faltan ciegas y pagos sin monto), así que se
        # usa la diferencia de fichas entre esta mano y la siguiente del mismo
        # jugador, que se calcula abajo.
        m["_saldos"] = saldos
        m["_cg"] = cg

    return perfiles, yo, total


def netos(manos, perfiles):
    """Suma el cambio de fichas de cada jugador entre manos consecutivas de la misma mesa."""
    anterior = {}
    for m in manos:
        mesa = m["id_mano"].rsplit("-mano-", 1)[0]
        saldos = m["_saldos"]
        for j, saldo in saldos.items():
            clave = (mesa, j)
            if clave in anterior:
                previo, cg = anterior[clave]
                perfiles[j]["neto_bb"] += (saldo - previo) / cg
                perfiles[j]["manos_neto"] += 1
                if saldo <= 0 < previo:
                    perfiles[j]["quiebras"] += 1
            anterior[clave] = (saldo, m["_cg"])


def pct(a, b):
    return f"{100 * a / b:.0f} %" if b else "-"


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("archivos", nargs="*", help="archivos manos-*.jsonl (por defecto, los de bots/lince/datos)")
    ap.add_argument("-min", type=int, default=20, help="mínimo de manos para listar a un jugador")
    args = ap.parse_args()

    rutas = args.archivos or sorted(glob.glob(os.path.join("bots", "lince", "datos", "manos-*.jsonl")))
    if not rutas:
        sys.exit("no hay archivos manos-*.jsonl: el bot los escribe al jugar sin -sin-archivos")
    manos = list(leer(rutas))
    perfiles, yo, total = analizar(manos)
    netos(manos, perfiles)

    mesas = {m["id_mano"].rsplit("-mano-", 1)[0] for m in manos}
    print(f"{total} manos en {len(mesas)} mesas · {len(perfiles)} jugadores vistos · nosotros: `{yo}`\n")
    print("| Jugador | Manos | Entra | Sube pre | Agresión | Se retira | All-in | Showdown | Neto (bb/100) | Quiebras |")
    print("| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |")
    filas = sorted(perfiles.items(), key=lambda kv: -kv[1]["manos"])
    for jugador, p in filas:
        if p["manos"] < args.min:
            continue
        nombre = f"**{jugador} (Lince)**" if jugador == yo else jugador
        neto = f"{100 * p['neto_bb'] / p['manos_neto']:+.0f}" if p["manos_neto"] else "-"
        print(
            f"| {nombre} | {p['manos']} | {pct(p['entro'], p['manos'])} | {pct(p['subio_pre'], p['manos'])} "
            f"| {pct(p['agresivas'], p['agresivas'] + p['pasivas'])} | {pct(p['se_retiro'], p['enfrento'])} "
            f"| {pct(p['allin'], p['manos'])} | {pct(p['showdown'], p['manos'])} | {neto} | {p['quiebras']} |"
        )
    ocultos = sum(1 for _, p in filas if p["manos"] < args.min)
    if ocultos:
        print(f"\n{ocultos} jugadores con menos de {args.min} manos no se listan.")


if __name__ == "__main__":
    main()
