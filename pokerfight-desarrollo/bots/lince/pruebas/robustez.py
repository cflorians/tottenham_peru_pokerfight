"""Prueba de robustez de Lince contra la mesa real del kit.

Reproduce en chico lo que pasa el día del torneo y comprueba que el bot:

  1. aguanta conectado y en silencio mientras no hay partida;
  2. juega una mesa de 5 y, cuando la mesa cierra, reconecta solo a la siguiente;
  3. cuenta bien las rondas;
  4. detecta la mesa final (6 jugadores, todos con 10 000 fichas);
  5. si se termina su proceso durante la clasificatoria o la final, el supervisor
     lo relanza y la mesa vuelve a aceptar jugadas desde el mismo asiento.

Se corre desde la raíz de pokerfight-desarrollo, con todo compilado:

    go build -o bin/ ./cmd/... ./bots/...
    python bots/lince/pruebas/robustez.py            # ~2 minutos
    python bots/lince/pruebas/robustez.py -silencio 3600   # con 1 hora en silencio

No toca el kit: solo lanza bin/mesa.exe, los bots de ejemplo y el supervisor.
"""

import argparse
import csv
import json
import re
import shutil
import subprocess
import sys
import time
from pathlib import Path

from rivales import leer

REPO = Path(__file__).resolve().parents[3]
BIN = REPO / "bin"
SUPERVISOR = REPO / "bots" / "lince" / "correr.ps1"
PUERTO = 9117
TOKEN = "lince-robustez"
CONEXION = re.compile(r"conectado a \S+ como (\S+) \(silla (\d+)\)")

resultados = []


def comprobar(nombre, ok, detalle=""):
    resultados.append((nombre, ok))
    print(f"  [{'OK' if ok else 'FALLA'}] {nombre}" + (f" ({detalle})" if detalle else ""), flush=True)
    return ok


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
        # Un nombre por ejecución permite terminar únicamente el Lince de esta prueba.
        self.exe_lince = self.salidas / f"lince-{datos.name}.exe"

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
        shutil.copyfile(BIN / "lince.exe", self.exe_lince)
        return self.lanzar("supervisor", [
            "powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", str(SUPERVISOR),
            "-Addr", f"127.0.0.1:{PUERTO}", "-Token", TOKEN, "-Datos", str(self.datos),
            "-Exe", str(self.exe_lince), "-Silencioso", "-Nuevo"])

    def pid_lince(self):
        return pid_del_bot(self.exe_lince.name)

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

    def jugadas_confirmadas(self, mesa):
        # La acción de respaldo del servidor solo puede ser pasar o retirarse.
        # Un pago, apuesta o subida en el historial acredita una respuesta aceptada.
        for mano in leer(sorted(self.datos.glob("manos-*.jsonl"))):
            if mano["id_mano"].rpartition("-mano-")[0] != mesa:
                continue
            for accion in mano.get("acciones") or []:
                tipo = accion["accion"]["tipo"]
                if accion["id_jugador"] == TOKEN and tipo in {"call", "bet", "raise", "allin"}:
                    yield mano["id_mano"], tipo

    def cerrar(self):
        for p in self.procesos:
            if p.poll() is None:
                subprocess.run(["taskkill", "/F", "/T", "/PID", str(p.pid)], capture_output=True)
        self.exe_lince.unlink(missing_ok=True)


def pid_del_bot(nombre="lince.exe"):
    """PID del ejecutable con ese nombre, si está corriendo."""
    salida = subprocess.run(["tasklist", "/FI", f"IMAGENAME eq {nombre}", "/FO", "CSV", "/NH"],
                            capture_output=True).stdout.decode("utf-8", errors="replace")
    for fila in csv.reader(salida.splitlines()):
        if len(fila) > 1 and fila[0].casefold() == nombre.casefold():
            return int(fila[1])
    return None


def probar_reinicio(prueba, mesa, ronda, fase, mano_objetivo):
    if not comprobar(f"la ronda llega a la mano {mano_objetivo}",
                     esperar(lambda: prueba.estado().get("manos_ronda", 0) >= mano_objetivo, 40)):
        return
    pid = prueba.pid_lince()
    registro = prueba.log()
    asientos = CONEXION.findall(registro)
    antes = prueba.estado()
    if not comprobar("Lince está sentado en la mesa antes del reinicio",
                     pid is not None and mesa.poll() is None and bool(asientos)
                     and antes.get("ronda") == ronda and antes.get("fase") == fase):
        return

    parada = subprocess.run(["taskkill", "/F", "/PID", str(pid)], capture_output=True)
    if not comprobar("el proceso anterior termina",
                     parada.returncode == 0 and esperar(lambda: prueba.pid_lince() != pid, 5),
                     f"tras registrar {antes['manos_ronda']} manos"):
        return
    if not comprobar("el supervisor lo relanza", esperar(lambda: prueba.pid_lince() not in (None, pid), 15)):
        return
    comprobar("retoma el estado de disco",
              esperar(lambda: prueba.log().count("estado: retomado") > registro.count("estado: retomado"), 10))
    if not comprobar("reconecta a la mesa", esperar(lambda: len(CONEXION.findall(prueba.log())) > len(asientos), 15)):
        return

    jugador, silla = CONEXION.findall(prueba.log())[-1]
    comprobar("vuelve con el mismo jugador y asiento", (jugador, silla) == asientos[-1], f"silla {silla}")

    # Solo cuentan jugadas que aparezcan en el historial después de reconectar.
    mesa_actual = antes["mesa"]
    jugadas_previas = set(prueba.jugadas_confirmadas(mesa_actual))

    def jugada_nueva():
        return next((jugada for jugada in prueba.jugadas_confirmadas(mesa_actual) if jugada not in jugadas_previas), None)

    esperar(lambda: jugada_nueva() or mesa.poll() is not None, 30)
    jugada = jugada_nueva()
    comprobar("la mesa acepta una jugada de Lince tras el reinicio", jugada is not None,
              f"{jugada[0]}, {jugada[1]}" if jugada else "")
    comprobar("avanza el contador de manos",
              esperar(lambda: prueba.estado().get("manos_ronda", 0) > antes["manos_ronda"], 10),
              f"mano {antes['manos_ronda']} → {prueba.estado().get('manos_ronda')}")
    despues = prueba.estado()
    comprobar(f"conserva la ronda {ronda} y la fase {fase}",
              despues.get("ronda") == ronda and despues.get("fase") == fase)
    comprobar("conserva la misma mesa y hora de inicio de ronda",
              despues.get("mesa") == mesa_actual and despues.get("inicio_ronda") == antes["inicio_ronda"])
    if fase == "final":
        comprobar("conserva la hora de inicio de la final", despues.get("inicio_final") == antes["inicio_final"])
    comprobar("no la tomó como mesa nueva", prueba.log().count("MESA NUEVA") == registro.count("MESA NUEVA"))


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
        comprobar("sigue conectado tras el silencio", "DESCONECTADO" not in t.log() and t.pid_lince() is not None)

        # --- 2. Primera mesa de 5 --------------------------------------------
        print("2. Mesa de 5, 50 manos", flush=True)
        t.rivales("r1", ["darwin", "espia", "base", "estadistico"])
        comprobar("la mesa termina", esperar(lambda: mesa1.poll() is not None, 120))
        registro = t.log()
        comprobar("anota la ronda 1 como clasificatoria", "MESA NUEVA · clasificatoria · ronda 1" in registro)
        comprobar("nota que la mesa cerró", esperar(lambda: "DESCONECTADO" in t.log(), 10))

        # --- 3. Segunda mesa: reconectar y recuperar el proceso ----------------
        print("3. Segunda mesa de 5: reconexión y reinicio en clasificatoria", flush=True)
        time.sleep(4)  # Un rato sin servidor, como entre el fin de una sesión y la siguiente.
        mesa2 = t.mesa("ronda2", 5, 50)
        t.rivales("r2", ["darwin", "espia", "base", "estadistico"])
        comprobar("anota la ronda 2 como clasificatoria",
                  esperar(lambda: "MESA NUEVA · clasificatoria · ronda 2" in t.log(), 30))
        if comprobar("el estado en disco dice ronda 2", esperar(lambda: t.estado().get("ronda") == 2, 5)):
            probar_reinicio(t, mesa2, ronda=2, fase="clasificatoria", mano_objetivo=10)
        comprobar("la segunda mesa termina", esperar(lambda: mesa2.poll() is not None, 120))

        # --- 4 y 5. Mesa final, y matar el proceso a mitad -----------------------
        print("4. Mesa de 6 con 10 000 fichas cada uno: la final", flush=True)
        time.sleep(2)
        mesa3 = t.mesa("final", 6, 100000)
        t.rivales("f", ["conservador"] * 5)
        comprobar("detecta la final", esperar(lambda: "es la FINAL" in t.log(), 10, cada=0.02))
        comprobar("el estado en disco dice fase final", esperar(lambda: t.estado().get("fase") == "final", 5))

        print("5. Se mata el proceso del bot a mitad de la final", flush=True)
        probar_reinicio(t, mesa3, ronda=3, fase="final", mano_objetivo=20)
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
