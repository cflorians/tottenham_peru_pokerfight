"""Prueba de robustez de Lince contra la mesa real del kit.

Reproduce en chico lo que pasa el día del torneo y comprueba que el bot:

  1. aguanta conectado y en silencio mientras no hay partida;
  2. juega una mesa de 5 y, cuando la mesa cierra, reconecta solo a la siguiente;
  3. cuenta bien las rondas;
  4. detecta la mesa final (6 jugadores, todos con 10 000 fichas);
  5. si le matan el proceso a mitad de una mesa, el supervisor lo relanza y el
     bot retoma su silla sabiendo en qué ronda y en qué fase estaba.

Se corre desde la raíz de pokerfight-desarrollo, con todo compilado:

    go build -o bin/ ./cmd/... ./bots/...
    python bots/lince/pruebas/robustez.py            # ~2 minutos
    python bots/lince/pruebas/robustez.py -silencio 3600   # con 1 hora en silencio

No toca el kit: solo lanza bin/mesa.exe, los bots de ejemplo y el supervisor.
"""

import argparse
import json
import os
import subprocess
import sys
import time
from pathlib import Path

REPO = Path(__file__).resolve().parents[3]
BIN = REPO / "bin"
SUPERVISOR = REPO / "bots" / "lince" / "correr.ps1"
PUERTO = 9117
TOKEN = "lince-robustez"

resultados = []


def comprobar(nombre, ok, detalle=""):
    resultados.append((nombre, ok))
    print(f"  [{'OK' if ok else 'FALLA'}] {nombre}" + (f" ({detalle})" if detalle else ""), flush=True)


def esperar(condicion, segundos, cada=0.1):
    """Espera hasta que condicion() sea verdadera. Devuelve si lo fue a tiempo."""
    limite = time.time() + segundos
    while time.time() < limite:
        if condicion():
            return True
        time.sleep(cada)
    return condicion()


class Prueba:
    def __init__(self, datos: Path):
        self.datos = datos
        self.procesos = []
        self.salidas = datos / "salidas"
        self.salidas.mkdir(parents=True, exist_ok=True)

    def lanzar(self, nombre, comando):
        salida = open(self.salidas / f"{nombre}.txt", "w", encoding="utf-8", errors="replace")
        p = subprocess.Popen(comando, cwd=REPO, stdout=salida, stderr=subprocess.STDOUT,
                             creationflags=subprocess.CREATE_NEW_PROCESS_GROUP)
        self.procesos.append(p)
        return p

    def mesa(self, nombre, jugadores, manos):
        return self.lanzar(nombre, [
            str(BIN / "mesa.exe"), "-addr", f"127.0.0.1:{PUERTO}", "-id", nombre,
            "-jugadores", str(jugadores), "-min-jugadores", str(jugadores),
            "-rondas", str(manos), "-stack", "10000",
            "-ciega-chica", "50", "-ciega-grande", "100", "-timeout-ms", "5000"])

    def rivales(self, etiqueta, bots):
        for i, bot in enumerate(bots):
            args = ["-silencioso"] if bot in ("darwin", "espia") else []
            self.lanzar(f"{etiqueta}-{bot}{i}", [str(BIN / f"{bot}.exe"), "-addr", f"127.0.0.1:{PUERTO}",
                                                "-token", f"{etiqueta}-{bot}{i}"] + args)

    def supervisor(self):
        return self.lanzar("supervisor", [
            "powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", str(SUPERVISOR),
            "-Addr", f"127.0.0.1:{PUERTO}", "-Token", TOKEN, "-Datos", str(self.datos), "-Silencioso", "-Nuevo"])

    def log(self):
        texto = ""
        for archivo in sorted(self.datos.glob("lince-*.log")):
            texto += archivo.read_text(encoding="utf-8", errors="replace")
        return texto

    def estado(self):
        try:
            return json.loads((self.datos / "estado.json").read_text(encoding="utf-8"))
        except (OSError, ValueError):
            return {}

    def cerrar(self):
        for p in self.procesos:
            if p.poll() is None:
                subprocess.run(["taskkill", "/F", "/T", "/PID", str(p.pid)], capture_output=True)


def pid_del_bot():
    """PID de lince.exe (el que lanzó el supervisor)."""
    salida = subprocess.run(["tasklist", "/FI", "IMAGENAME eq lince.exe", "/FO", "CSV", "/NH"],
                            capture_output=True).stdout.decode("utf-8", errors="replace")
    for linea in salida.splitlines():
        partes = [x.strip('"') for x in linea.split('","')]
        if partes and partes[0].strip('"').lower() == "lince.exe":
            return int(partes[1])
    return None


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("-silencio", type=int, default=40, help="segundos conectado en silencio antes de la primera mesa")
    opciones = parser.parse_args()

    if pid_del_bot():
        sys.exit("Ya hay un lince.exe corriendo: cerrarlo antes de probar.")

    datos = REPO / "pruebas-lince" / ("robustez-" + time.strftime("%Y%m%d-%H%M%S"))
    datos.mkdir(parents=True)
    t = Prueba(datos)
    try:
        # --- 1. Silencio ---------------------------------------------------
        print(f"1. Conectado en silencio {opciones.silencio} s (la mesa espera jugadores)", flush=True)
        mesa1 = t.mesa("ronda1", 5, 50)
        time.sleep(1)
        t.supervisor()
        comprobar("el bot conecta", esperar(lambda: "conectado a" in t.log(), 20))
        time.sleep(opciones.silencio)
        comprobar("sigue conectado tras el silencio", "DESCONECTADO" not in t.log() and pid_del_bot() is not None)

        # --- 2. Primera mesa de 5 --------------------------------------------
        print("2. Mesa de 5, 50 manos", flush=True)
        t.rivales("r1", ["darwin", "espia", "base", "estadistico"])
        comprobar("la mesa termina", esperar(lambda: mesa1.poll() is not None, 120))
        registro = t.log()
        comprobar("anota la ronda 1 como clasificatoria", "MESA NUEVA · clasificatoria · ronda 1" in registro)
        comprobar("nota que la mesa cerró", esperar(lambda: "DESCONECTADO" in t.log(), 10))

        # --- 3. Segunda mesa: tiene que reconectar solo ------------------------
        print("3. Segunda mesa de 5: reconexión sin ayuda", flush=True)
        time.sleep(4)  # Un rato sin servidor, como entre el fin de una sesión y la siguiente.
        mesa2 = t.mesa("ronda2", 5, 50)
        t.rivales("r2", ["darwin", "espia", "base", "estadistico"])
        comprobar("la segunda mesa se juega (el bot volvió solo)", esperar(lambda: mesa2.poll() is not None, 120))
        comprobar("anota la ronda 2 como clasificatoria", "MESA NUEVA · clasificatoria · ronda 2" in t.log())
        comprobar("el estado en disco dice ronda 2", esperar(lambda: t.estado().get("ronda") == 2, 5), f"ronda {t.estado().get('ronda')}")

        # --- 4 y 5. Mesa final, y matar el proceso a mitad -----------------------
        print("4. Mesa de 6 con 10 000 fichas cada uno: la final", flush=True)
        time.sleep(2)
        mesa3 = t.mesa("final", 6, 100000)
        t.rivales("f", ["conservador"] * 5)
        comprobar("detecta la final", esperar(lambda: "es la FINAL" in t.log(), 10, cada=0.02))
        comprobar("el estado en disco dice fase final", esperar(lambda: t.estado().get("fase") == "final", 5))
        inicio_final = t.estado().get("inicio_final")

        print("5. Se mata el proceso del bot a mitad de la final", flush=True)
        esperar(lambda: t.estado().get("manos_ronda", 0) >= 20, 30, cada=0.02)
        pid = pid_del_bot()
        manos_antes = t.estado().get("manos_ronda", 0)
        mesas_antes = t.log().count("MESA NUEVA")
        subprocess.run(["taskkill", "/F", "/PID", str(pid)], capture_output=True)
        comprobar("el proceso murió", esperar(lambda: pid_del_bot() != pid, 5), f"en la mano {manos_antes} de la final")
        comprobar("el supervisor lo relanza", esperar(lambda: pid_del_bot() not in (None, pid), 15))
        comprobar("retoma el estado de disco", esperar(lambda: "estado: retomado" in t.log(), 10))
        comprobar("vuelve a su silla", esperar(lambda: t.log().count("conectado a") >= 4, 15))
        comprobar("sigue jugando la misma mesa", esperar(lambda: t.estado().get("manos_ronda", 0) > manos_antes, 60),
                  f"mano {manos_antes} → {t.estado().get('manos_ronda')}")
        e = t.estado()
        comprobar("sigue en la ronda 3, fase final", e.get("ronda") == 3 and e.get("fase") == "final", f"ronda {e.get('ronda')}, fase {e.get('fase')}")
        comprobar("conserva la hora de inicio de la final", e.get("inicio_final") == inicio_final)
        comprobar("no la tomó como mesa nueva", t.log().count("MESA NUEVA") == mesas_antes)
        comprobar("la mesa sigue viva", mesa3.poll() is None)

        registro = t.log()
        comprobar("ningún pánico ni error en el log", "PÁNICO" not in registro and "ERROR" not in registro)
        comprobar("ningún aviso de decisión lenta", "AVISO" not in registro)
    finally:
        t.cerrar()

    fallas = [n for n, ok in resultados if not ok]
    print(f"\n{len(resultados) - len(fallas)} de {len(resultados)} comprobaciones bien. Logs en {datos}")
    if fallas:
        print("FALLARON: " + "; ".join(fallas))
        sys.exit(1)


if __name__ == "__main__":
    main()
