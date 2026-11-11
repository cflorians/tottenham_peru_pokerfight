# Lince · nuestro bot para PokerFight

Nombre de trabajo; se puede cambiar. Parte de **Gato Darwin** (`bots/darwin`, de AICS PUCP): política con
números evolucionados por un algoritmo genético y fuerza de la mano por Monte Carlo.

**Versión actual: v0.** Juega igual que Darwin con el genoma prudente (`Simulaciones/genoma_prudente.json`); lo
nuevo es la infraestructura para aguantar el campeonato entero. El plan está en `../../../PLAN-BOT.md`.

## Qué agrega sobre Darwin

| | |
| --- | --- |
| Reconexión | Si se corta la conexión, vuelve a entrar con el mismo token (espera de 1 a 5 s, sin límite de intentos) |
| No se cae por errores propios | Un pánico al decidir juega la acción segura (check o fold) y queda en el log |
| Contexto de torneo (`torneo/`) | Ronda, manos, fase (clasificatoria o final), fichas de cada alias. Se guarda en disco al cerrar cada mano |
| Supervisor (`correr.ps1`) | Relanza el proceso si muere; una sola instancia por máquina |
| Log a archivo | Cada decisión, cambio de mesa y reconexión, en `datos/lince-AAAAMMDD.log` |

En v0 el contexto de torneo **solo se registra**: todavía no cambia cómo juega.

## Compilar y probar

Desde la raíz de `pokerfight-desarrollo`:

```bash
go build -o bin/ ./bots/lince
go test -trimpath ./bots/lince/...
python bots/lince/pruebas/robustez.py
```

`robustez.py` levanta mesas reales del kit y comprueba silencio, reconexión entre mesas, conteo de rondas,
detección de la final y que matar el proceso a mitad de una mesa no pierde nada (~2 minutos; con
`-silencio 3600` deja al bot una hora conectado sin partida).

## Correrlo

En una mesa o un torneo de la arena local (termina cuando la mesa cierra, no escribe archivos):

```json
{ "id": "lince", "nombre": "Lince", "comando": ["./bin/lince", "-silencioso", "-sin-archivos", "-una-partida"] }
```

Contra un servidor de la organización, siempre con el supervisor:

```powershell
powershell -ExecutionPolicy Bypass -File bots\lince\correr.ps1 -Addr HOST:PUERTO -Token TU_TOKEN -Nuevo
```

- `-Nuevo` aparta el estado guardado. **Solo al empezar un torneo**, nunca al relanzar a mitad.
- En el **servidor de pruebas** agregar `-Fase clasificatoria`: ahí todos vuelven a 10 000 fichas en cada
  ronda, y una mesa de 6 se confundiría con la final.
- El token también se puede pasar en la variable `POKERFIGHT_TOKEN`.

Flags del bot:

| Flag | Por defecto | Para qué |
| --- | --- | --- |
| `-addr`, `-token` | `localhost:9000`, `lince` | Servidor y credencial |
| `-datos carpeta` | `lince-datos` | Dónde van el estado y el log |
| `-fase` | `auto` | `auto`, `clasificatoria` o `final` |
| `-una-partida` | no | Termina cuando la mesa cierra en vez de reconectar |
| `-sin-archivos` | no | No escribe estado ni log |
| `-silencioso` | no | No imprime cada decisión en la terminal |
| `-genoma`, `-semilla`, `-muestras` | | Los de Darwin |

## Cómo deduce la ronda y la final

La arena no avisa el cambio de ronda. Se toma como mesa nueva si cambia el prefijo del id de mano, si
aparece un jugador que no estaba, o si el número de mano vuelve atrás. Si el id de mano no se puede leer,
valen la cuenta de 50 manos y una pausa de 30 s. La final es una mesa nueva con 6 jugadores, todos con
10 000 fichas, después de haber jugado al menos una mesa. Ante la duda, clasificatoria.

Pendiente de confirmar con el servidor de pruebas: qué forma tienen los id de mano reales.

## Smart App Control

Windows bloquea algunos `.exe` recién compilados ("An Application Control policy has blocked this file"), y
el veredicto cambia con cada compilación: el mismo código puede pasar o no. Después de compilar la versión
que se va a usar, **comprobar que arranca** (`bin\lince.exe -h`) y no volver a compilarla. Si la bloquea,
recompilar con otras opciones (`-trimpath`, `-ldflags=-s`) suele dar un binario que pasa.

## Estructura

| Ruta | Qué hay |
| --- | --- |
| `main.go` | Conexión con reconexión, decisión protegida, logs |
| `torneo/` | Contexto de torneo y estado en disco |
| `estrategia/` | Genoma, política y fuerza (sin cambios respecto de Darwin) |
| `entrenamiento/`, `cmd/entrenar-lince/` | Entrenador genético (sin cambios respecto de Darwin) |
| `genoma.json` | Genoma con el que juega (v0: el prudente). `genoma_darwin.json` es el original |
| `correr.ps1` | Supervisor |
| `pruebas/robustez.py` | Prueba de robustez contra la mesa real |
