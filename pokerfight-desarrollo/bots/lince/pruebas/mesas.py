"""Mesas reales en condiciones de torneo, para comparar genomas de Lince.

Es el "nivel real" liviano: levanta la arena del kit (formato mesa-unica: una
ronda de la clasificatoria, o una mesa final con -final) con los bots de
verdad, muchas veces, y resume cuánto gana y cuántas veces quiebra Lince con
cada genoma. Sirve para validar lo que dice el simulador rápido.

Desde pokerfight-desarrollo, con todo compilado en bin/:

    python bots/lince/pruebas/mesas.py -mesas 30 v0=bots/lince/genoma.json v1=bots/lince/genoma_clasificatoria.json
    python bots/lince/pruebas/mesas.py -final -mesas 50 v0=bots/lince/genoma.json final=bots/lince/genoma_final.json

Con -final hay tres alineaciones: F1 y F2 (Lince contra bots del kit) y F3
(Lince contra cuatro copias de nuestro bot con otros genomas de
bots/lince/rivales/ -v1, v1.1, b2 y e- más el novato: rivales "fuertes").

Cada variante es nombre=genoma.json. Todas juegan las mismas alineaciones y
semillas, pero la arena real NO es reproducible (algún bot del kit usa azar sin
sembrar): son muestras independientes, no comparaciones pareadas.
"""

import argparse
import concurrent.futures
import json
import math
import os
import re
import subprocess
import sys
import tempfile

RIVALES = {
    "darwin": {"id": "darwin", "nombre": "Gato Darwin", "comando": ["./bin/darwin", "-silencioso"]},
    "espia": {"id": "espia", "nombre": "Michi Espia", "comando": ["./bin/espia", "-silencioso"]},
    "estadistico": {"id": "estadistico", "nombre": "Bot Estadistico", "comando": ["./bin/estadistico", "-simulaciones", "2000"]},
    "base": {"id": "base", "nombre": "Bot Base", "comando": ["./bin/base"]},
    "conservador": {"id": "conservador", "nombre": "Bot Conservador", "comando": ["./bin/conservador"]},
    "aleatorio": {"id": "aleatorio", "nombre": "Bot Aleatorio", "comando": ["./bin/aleatorio"]},
    "novato": {"id": "novato", "nombre": "Michi Novato", "comando": [sys.executable, "bots/novato/bot.py", "-silencioso"]},
}


def _rival_fuerte(clave, nombre, genoma):
    """Rival 'fuerte': el mismo ejecutable de Lince con otro genoma (sin la palabra Lince en id ni nombre)."""
    return {
        "id": f"rival-{clave}",
        "nombre": f"Rival {nombre}",
        "comando": ["./bin/lince", "-silencioso", "-sin-archivos", "-una-partida", "-fase", "clasificatoria", "-genoma", f"bots/lince/rivales/{genoma}.json"],
    }


for _clave, _nombre, _genoma in [("v1", "v1", "v1"), ("v1_1", "v1.1", "v1_1"), ("b2", "b2", "final_b2"), ("e", "e", "e")]:
    RIVALES[f"rival-{_clave}"] = _rival_fuerte(_clave, _nombre, _genoma)

# Alineaciones de la clasificatoria (Lince más estos rivales): las mismas ideas
# que Simulaciones/pruebas_formato_real.md.
ALINEACIONES = {
    "A": ["darwin", "espia", "estadistico", "base"],
    "B": ["espia", "estadistico", "novato", "conservador"],
    "C": ["darwin", "espia", "estadistico", "base", "aleatorio"],
}
# Mesas finales de 6: los tipos de bot que llegan a la final en el simulador.
FINALES = {
    "F1": ["novato", "espia", "estadistico", "darwin", "base"],
    "F2": ["novato", "espia", "estadistico", "darwin", "conservador"],
    # Rivales fuertes: el mismo bot con otros genomas (bots/lince/rivales/).
    "F3": ["rival-v1", "rival-v1_1", "rival-b2", "rival-e", "novato"],
}


def configuracion(rivales, genoma, semilla, manos, salida):
    lince = {
        "id": "lince",
        "nombre": "Lince",
        "comando": ["./bin/lince", "-silencioso", "-sin-archivos", "-una-partida", "-fase", "clasificatoria", "-genoma", genoma],
    }
    return {
        "nombre": "lince-mesas",
        "formato": "mesa-unica",
        "puntuacion": "neto",
        "manos_por_enfrentamiento": manos,
        "ida_y_vuelta": False,
        "reponer_stack": False,
        "semilla": semilla,
        "paralelismo": 1,
        "penalizacion_timeout": 0,
        "guardar_historial": False,
        "salida": salida,
        "mesa": {"ciega_chica": 50, "ciega_grande": 100, "stack": 10000, "timeout_ms": 5000, "espera_jugadores_ms": 20000},
        "participantes": [lince] + [RIVALES[r] for r in rivales],
    }


def jugar(tarea):
    variante, genoma, alineacion, rivales, semilla, manos = tarea
    with tempfile.TemporaryDirectory(prefix="lince-mesas-") as tmp:
        ruta = os.path.join(tmp, "mesa.json")
        with open(ruta, "w", encoding="utf-8") as f:
            json.dump(configuracion(rivales, genoma, semilla, manos, os.path.join(tmp, "resultados")), f)
        arena = os.path.join("bin", "arena.exe" if os.name == "nt" else "arena")
        try:
            r = subprocess.run([arena, "correr", "-config", ruta], capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=900)
        except subprocess.TimeoutExpired:
            return variante, alineacion, semilla, None, None, "la arena no terminó"
    salida = r.stdout + r.stderr
    # Fila de la tabla final: puesto, nombre, fichas, jugados, ganados, perdidos, neto, timeouts.
    m = re.search(r"^(\d+)\s+Lince\s+(-?\d+)\s+\d+\s+\d+\s+\d+\s+[+-]?\d+\s+(\d+)\s*$", salida, re.M)
    if not m:
        return variante, alineacion, semilla, None, None, salida[-300:]
    # Gana la mesa quien queda 1.º en la tabla de la arena (más fichas al terminar).
    return variante, alineacion, semilla, int(m.group(2)), (int(m.group(1) == "1"), int(m.group(3))), None


def wilson(k, n):
    if n == 0:
        return "-"
    z = 1.96
    p = k / n
    den = 1 + z * z / n
    centro = (p + z * z / (2 * n)) / den
    radio = z * math.sqrt(p * (1 - p) / n + z * z / (4 * n * n)) / den
    return f"{100 * p:.0f} % [{100 * max(centro - radio, 0):.0f}-{100 * min(centro + radio, 1):.0f}]"


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("variantes", nargs="+", help="nombre=ruta/al/genoma.json")
    ap.add_argument("-mesas", type=int, default=30, help="mesas por variante y alineación")
    ap.add_argument("-final", action="store_true", help="mesas finales: 6 jugadores hasta que quede uno (o -manos)")
    ap.add_argument("-manos", type=int, default=0, help="manos por mesa (por defecto 50; en la final, 2000)")
    ap.add_argument("-alineaciones", default="", help="cuáles usar, separadas por coma (por defecto todas)")
    ap.add_argument("-semilla", type=int, default=1000, help="primera semilla")
    ap.add_argument("-paralelo", type=int, default=3, help="mesas a la vez (muchas a la vez pueden causar timeouts)")
    args = ap.parse_args()

    grupos = FINALES if args.final else ALINEACIONES
    if args.alineaciones:
        grupos = {k: grupos[k] for k in args.alineaciones.split(",")}
    manos = args.manos or (2000 if args.final else 50)

    variantes = []
    for v in args.variantes:
        nombre, _, ruta = v.partition("=")
        if not os.path.exists(ruta):
            sys.exit(f"no existe el genoma {ruta!r}")
        variantes.append((nombre, ruta.replace("\\", "/")))

    tareas = [
        (nombre, ruta, alin, rivales, args.semilla + i, manos)
        for i in range(args.mesas)
        for alin, rivales in grupos.items()
        for nombre, ruta in variantes
    ]
    resultados = {}
    fallas = 0
    with concurrent.futures.ThreadPoolExecutor(max_workers=args.paralelo) as ex:
        for n, (variante, alin, semilla, neto, extra, error) in enumerate(ex.map(jugar, tareas), 1):
            if error:
                fallas += 1
                print(f"  mesa fallida ({variante}, {alin}, semilla {semilla}): {error}", file=sys.stderr)
                continue
            resultados.setdefault((variante, alin), []).append((neto, *extra))
            if n % 10 == 0:
                print(f"  {n}/{len(tareas)} mesas", file=sys.stderr)

    tipo = "Mesa final (6 jugadores, gana el que tiene más fichas al terminar)" if args.final else "Ronda de clasificatoria"
    print(f"{tipo} · {manos} manos · 10 000 fichas · ciegas 50/100 · plazo 5 s · {args.mesas} mesas por variante y alineación\n")
    for alin, rivales in grupos.items():
        print(f"- **{alin}**: Lince + {', '.join(rivales)}")
    print()
    cabecera = "| Variante | Alineación | Mesas | Neto medio | Quiebra | Gana la mesa | Timeouts |"
    print(cabecera + "\n| --- | --- | ---: | ---: | ---: | ---: | ---: |")
    for nombre, _ in variantes:
        todas = []
        for alin in list(grupos) + ["todas"]:
            filas = todas if alin == "todas" else resultados.get((nombre, alin), [])
            if alin != "todas":
                todas += filas
            if not filas:
                continue
            n = len(filas)
            neto = sum(f[0] for f in filas) / n
            quiebras = sum(1 for f in filas if f[0] <= -10000)
            ganadas = sum(f[1] for f in filas)
            timeouts = sum(f[2] for f in filas)
            print(f"| {nombre} | {alin} | {n} | {neto:+.0f} | {wilson(quiebras, n)} | {wilson(ganadas, n)} | {timeouts} |")
    if fallas:
        print(f"\n{fallas} mesas fallaron y no se cuentan.")


if __name__ == "__main__":
    main()
