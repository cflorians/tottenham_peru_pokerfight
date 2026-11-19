# Lince · nuestro bot para PokerFight

[README breve de entrega para el jurado](README-ENTREGA.md)

Nombre de trabajo; se puede cambiar. Parte de **Gato Darwin** (`bots/darwin`, de AICS PUCP): política con
números evolucionados por un algoritmo genético y fuerza de la mano por Monte Carlo.

**Versión actual: v1.2.** La política es la de Darwin, pero el genoma se entrenó con lo que premia el torneo
(llegar a la final y terminar 1.º en ligas simuladas de 10 rondas) en vez de las fichas por mano, y alrededor
está la infraestructura para aguantar el campeonato entero. Resultados: `Simulaciones/a2_b2_v1.md`. El plan está en `../../../PLAN-BOT.md` y las
reglas de trabajo en equipo (ramas, PR, archivos compartidos, token) en `../../../AGENTS.md`.

## Qué agrega sobre Darwin

| | |
| --- | --- |
| Reconexión | Si se corta la conexión, vuelve a entrar con el mismo token (espera de 1 a 5 s, sin límite de intentos) |
| No se cae por errores propios | Un pánico al decidir juega la acción segura (check o fold) y queda en el log |
| Contexto de torneo (`torneo/`) | Ronda, manos, fase (clasificatoria o final), fichas de cada alias. Se guarda en disco al cerrar cada mano |
| Supervisor (`correr.ps1`) | Relanza el proceso si muere; una sola instancia por máquina |
| Log a archivo | Cada decisión, cambio de mesa y reconexión, en `datos/lince-AAAAMMDD.log` |
| Registro de manos | Cada mano completa (acciones de todos, cartas mostradas) en `datos/manos-AAAAMMDD.jsonl`, para estudiar a los rivales |
| Cartas a salvo | Las cartas propias se guardan en `datos/mano.json`: si la conexión o el proceso se cortan a mitad de una mano, al volver la sigue jugando (la mesa no las repite) |

El contexto de torneo decide **con qué genoma se juega**: `genoma.json` en la clasificatoria y
`genoma_final.json` al detectar la mesa final. El de la final se **adapta a cada rival**: el bot cuenta cuán
seguido apuesta cada alias (`rivales.go`, guardado en `datos/rivales.json`) y le resube y paga más liviano al
que apuesta casi siempre (`Simulaciones/b4_adaptacion_rival.md`).

## Compilar y probar

Desde la raíz de `pokerfight-desarrollo`:

```bash
go build -o bin/ ./bots/lince
go test -trimpath ./bots/lince/...
python bots/lince/pruebas/robustez.py
```

`robustez.py` (no corre si hay otro supervisor de Lince en la máquina, por ejemplo el del servidor de pruebas:
solo se permite una instancia) levanta mesas reales del kit y comprueba silencio, reconexión entre mesas, conteo de rondas,
detección de la final y que matar el proceso a mitad de una mesa no pierde nada (~2 minutos; con
`-silencio 3600` deja al bot una hora conectado sin partida).

## Simular el torneo entero

`simular-liga` juega miles de torneos completos (clasificatoria de 10 rondas y mesa final) en el simulador
rápido y reporta, en markdown, las métricas del plan: M1 (llega a la final), M1b (termina 1.º) y M2 (gana la
final). La primera corrida precalcula una biblioteca de repartos (unos minutos) y la guarda en `datos/`.

```bash
go run ./bots/lince/cmd/simular-liga                         # v0, Darwin original y Novato; 4 campos; 20 inscritos
go run ./bots/lince/cmd/simular-liga -n 15,20,30,40 -tabla   # todos los tamaños y las tablas de valor
go run ./bots/lince/cmd/simular-liga -heroes v0,otro.json    # comparar un genoma nuevo con la v0
```

Los supuestos donde las bases no dicen nada (orden de los eliminados, mesas con pocos jugadores) están
comentados en `entrenamiento/liga.go`. Resultados de referencia: `Simulaciones/a1_b1_linea_base.md`.

Otros modos: `-diagnostico` cuenta con qué jugada se queda sin fichas el bot; `-paridad` compara el simulador
con la arena real; `-genoma-final otro.json` hace que el bot juegue la mesa final con otro genoma. Si hay
genomas en `rivales/`, se agrega el campo **fuerte** (la mitad de los inscritos son esos genomas).

Con más de un bot en `-heroes`, todos juegan las mismas ligas y el reporte trae la **diferencia contra el
primero, liga por liga, con su intervalo**: es la comparación de la prueba de promoción.

## Entrenar con el premio del torneo

`entrenar-liga` evoluciona un genoma midiendo lo que premia el torneo, no las ciegas ganadas por mano:

```bash
go run ./bots/lince/cmd/entrenar-liga -modo clasificatoria -m2 0.65   # aptitud = premio esperado de la liga -> genoma_clasificatoria.json
go run ./bots/lince/cmd/entrenar-liga -modo final            # aptitud = % de mesas finales ganadas  -> genoma_final.json
go run ./bots/lince/cmd/entrenar-liga -rapido                # prueba de humo (no escribe en la carpeta del bot)
```

Con `-modo final -equilibrado`, la mitad de las finales son contra los genomas de `rivales/` y la aptitud es
el peor de los dos resultados (o su promedio ponderado con `-peso-fuertes`; el genoma de final de la v1.2 se
entrenó con `-peso-fuertes 0.25 -coevolucion 0`). Con `-campos kit,mixto` se eligen los campos de la clasificatoria. Con `-con-final`, cada liga en que el genoma clasifica juega también su mesa final y la aptitud es el premio
realmente cobrado. No pisa `genoma.json`: deja el campeón en un archivo aparte, elegido con torneos que el genético nunca vio.
Tarda unos 10 minutos con los valores por defecto.

## Validar en la arena real

Para probar una mesa local de cinco jugadores desde PowerShell, ejecuta estos comandos en
`pokerfight-desarrollo`:

```powershell
go build -o bin/ ./cmd/... ./bots/...
.\bin\arena.exe validar -config .\pruebas-lince\mesa5.json
.\bin\arena.exe correr -config .\pruebas-lince\mesa5.json
```

El historial queda en `pruebas-lince/resultados/`. La arena local genera los tokens de los bots.

`pruebas/mesas.py` juega muchas mesas con los bots de verdad del kit (una ronda de clasificatoria, o una mesa
final con `-final`) y resume neto, quiebras y timeouts de Lince con cada genoma:

```bash
python bots/lince/pruebas/mesas.py -mesas 40 v0=bots/lince/genoma.json v1=bots/lince/genoma_clasificatoria.json
python bots/lince/pruebas/mesas.py -final -mesas 20 clasif=bots/lince/genoma.json final=bots/lince/genoma_final.json
```

## Ensayar el torneo entero con los bots de verdad

`liga-real` juega una clasificatoria completa (10 rondas, mesas sorteadas en cada una, fichas arrastradas,
eliminación, corte de los 6) y la mesa final con los ejecutables reales del kit y el servidor de mesa del
kit. Lince corre como **un solo proceso durante toda la liga**, igual que en el torneo, y al terminar se
comprueba que haya contado bien las rondas y detectado la final. Una liga de 30 bots tarda un minuto.

```bash
go build -o bin/ ./bots/lince/cmd/liga-real
./bin/liga-real -n 30                 # una liga de 30 bots y su final
./bin/liga-real -n 30 -ligas 15       # quince, con resumen: finales alcanzadas, ganadas, quiebras por ronda
./bin/liga-real -n 10 -rondas 3 -manos 20   # prueba corta
```

Deja todo en `datos/liga-real/` (ignorado por git): el log y el estado de Lince, y un `liga.json` por liga.
Límites: aquí la conexión se cierra entre rondas y Lince reconecta (en el torneo queda abierta y en
silencio), y los rivales son los bots del kit. Resultados: `Simulaciones/ensayo_liga_real.md`.

## Estudiar a los rivales

Cuando el bot juega con archivos (sin `-sin-archivos`) graba cada mano. `pruebas/rivales.py` las resume por
jugador: cuánto entra, cuánto sube, agresión, cuánto se retira ante una apuesta, all-ins y resultado. Es la
forma de ver cómo juegan los otros equipos en el servidor de la organización.

```bash
python bots/lince/pruebas/rivales.py            # todo lo grabado en bots/lince/datos
python bots/lince/pruebas/rivales.py -min 50    # solo rivales con 50 manos o más
```

## Correrlo

En una mesa o un torneo de la arena local (termina cuando la mesa cierra, no escribe archivos):

```json
{ "id": "lince", "nombre": "Lince", "comando": ["./bin/lince", "-silencioso", "-sin-archivos", "-una-partida"] }
```

Contra un servidor de la organización, siempre con el supervisor:

```powershell
powershell -ExecutionPolicy Bypass -File bots\lince\correr.ps1 -Addr HOST:PUERTO -Nuevo
```

- `-Nuevo` aparta el estado guardado. **Solo al empezar un torneo**, nunca al relanzar a mitad.
- En el **servidor de pruebas** agregar `-Fase clasificatoria`: ahí todos vuelven a 10 000 fichas en cada
  ronda, y una mesa de 6 se confundiría con la final.
- El token se toma de la variable `POKERFIGHT_TOKEN` (ver cómo guardarla en `AGENTS.md`). No hace falta pasar
  `-Token`, y así no queda en la línea de comandos.

Flags del bot:

| Flag | Por defecto | Para qué |
| --- | --- | --- |
| `-addr`, `-token` | `localhost:9000`, variable `POKERFIGHT_TOKEN` | Servidor y credencial |
| `-datos carpeta` | `lince-datos` | Dónde van el estado y el log |
| `-fase` | `auto` | `auto`, `clasificatoria` o `final` |
| `-una-partida` | no | Termina cuando la mesa cierra en vez de reconectar |
| `-sin-archivos` | no | No escribe estado ni log |
| `-silencioso` | no | No imprime cada decisión en la terminal |
| `-genoma`, `-semilla`, `-muestras` | | Los de Darwin |
| `-genoma-final` | | Genoma de la final en lugar del incluido. Con `-genoma` y sin este flag, la final usa el de `-genoma` |

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
| `entrenamiento/liga.go`, `campos.go` | Simulador del torneo entero: liga de 10 rondas con fichas arrastradas, mesa final y campos de rivales |
| `cmd/simular-liga/` | Mide M1, M1b y M2 de un bot en miles de torneos simulados |
| `entrenamiento/genetico_liga.go`, `cmd/entrenar-liga/` | Entrenador con aptitud de liga o de mesa final |
| `pruebas/mesas.py` | Mesas reales en condiciones de torneo para comparar genomas |
| `cmd/liga-real/` | Torneo entero (clasificatoria y final) con los ejecutables de verdad: ensayo general |
| `registro.go`, `pruebas/rivales.py` | Registro de manos y perfil de cada rival |
| `rivales.go`, `estrategia/lectura.go` | Lectura de cada rival (cuán seguido apuesta) que usa la política al decidir |
| `genoma.json` | Genoma de la clasificatoria (v1.1). Lo incluye el binario |
| `genoma_final.json` | Genoma de la mesa final (v1.2, adaptativo al rival). Lo incluye el binario |
| `genoma_clasificatoria.json` | Última salida de `entrenar-liga -modo clasificatoria` (candidato; no lo usa el bot) |
| `rivales/*.json` | Genomas nuestros ya entrenados: son los rivales del campo "fuerte" del simulador |
| `genoma_v0.json`, `genoma_darwin.json` | El prudente de la v0 y el original de Darwin: referencia y rivales del simulador |
| `correr.ps1` | Supervisor |
| `pruebas/robustez.py` | Prueba de robustez contra la mesa real |
