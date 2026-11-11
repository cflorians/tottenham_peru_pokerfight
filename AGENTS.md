# AGENTS.md

Guía para los agentes de código que trabajan en este repositorio (**Codex** lee este archivo; **Claude Code**
lo recibe a través de `CLAUDE.md`, que solo lo importa). Es la única fuente de reglas: si una regla cambia,
se cambia aquí y no en `CLAUDE.md`.

Repositorio del equipo para **PokerFight** (competencia de bots de póker de AICS PUCP). Somos **2 personas**,
cada una con su agente (una con Claude Code, la otra con Codex), trabajando sobre el mismo repositorio de
GitHub. El equipo trabaja en español: **responder en español** y escribir en español el código, los
comentarios y la documentación.

## Reglas de carpetas (estrictas)

| Carpeta | Regla |
| --- | --- |
| `pokerfight-desarrollo/` | **La única carpeta donde se escribe código.** Nuestro bot es `pokerfight-desarrollo/bots/lince/` (nombre de trabajo; ver su README) y nuestras herramientas van en carpetas nuevas propias. Dentro de ella tampoco se edita el kit (`internal/`, `cmd/`, `pkg/`, `docs/`, los bots de ejemplo, su `README.md`): el servidor real no usa nuestra copia. El único archivo del kit que podemos editar es `torneo-go.json`, para sumar nuestro bot a las pruebas locales. Ver `pokerfight-desarrollo/AGENTS.md`. |
| `pokerfight-arena-participantes/` | **No se toca nunca.** Es la copia de referencia intacta del kit de la organización. Ni ediciones, ni compilaciones, ni `bin/` o `resultados/` adentro. Solo se lee, por ejemplo para comparar con `pokerfight-desarrollo`. |
| `Simulaciones/` | Resultados de simulaciones (markdown) y artefactos de prueba como `genoma_prudente.json`. Los resultados nuevos van aquí, **un archivo por experimento** (ver `Simulaciones/README.md`). |
| `Bases_Poker_Fight_de_Bots.docx`, `PokerFight - Guía de inicio para equipos.docx` | Bases oficiales y guía para equipos. Solo lectura. Para sacar el texto sin dependencias: descomprimir `word/document.xml` y extraer los elementos `<w:t>`. |

## El plan

**[`PLAN-BOT.md`](PLAN-BOT.md) es el plan de trabajo: leerlo antes de empezar cualquier trabajo en el bot.**
En resumen:

- Base: copia de **Gato Darwin** (`bots/darwin`, Go, algoritmo genético + Monte Carlo), más el modelado de
  rivales de **Michi Espía**.
- El objetivo es el premio mayor (ganar la mesa final): P(campeón) = P(llegar a la final) × P(ganar la final).
  Son dos juegos distintos: la clasificatoria (mesas de 5-6, fichas arrastradas 10 rondas, quebrar elimina,
  pasan los 6 con más fichas o los últimos eliminados) premia **sobrevivir y acumular lo suficiente**; la
  final (6 bots, 10 000 fichas, un ganador) premia acumular sin aversión al riesgo. Darwin original quiebra en
  el 44 % de las mesas de 5; un genoma prudente ajustado a mano lo baja a 10 %, pero se retira con manos
  ganadoras ante all-ins grandes: es solo el punto de partida (`Simulaciones/pruebas_formato_real.md`).
- `PLAN-BOT.md` es el plan maestro (infraestructura común, simulador, calendario, escalera de versiones,
  reparto del equipo, preguntas a la organización, riesgos). La estrategia está en dos planes que terminan
  siendo dos modos del mismo bot (solo se permite un bot por equipo):
  [`PLAN-CLASIFICATORIA.md`](PLAN-CLASIFICATORIA.md) (plan A: entrar entre los 6 y pelear el 1.º de la
  clasificatoria, que tiene premio propio) y [`PLAN-FINAL.md`](PLAN-FINAL.md) (plan B: ganar la mesa final).
  Leer el que corresponda al trabajo. Premio esperado = 60 × P(final) × P(ganar la final) + 30 × P(1.º en la
  clasificatoria).
- Escalera de versiones, siempre con un bot entregable y primero lo barato de cada plan: v0 Darwin + genoma
  prudente + infraestructura (reconexión, supervisor, estado en disco, log) → v1 (A) genes actuales
  reentrenados con fitness de liga en nuestro simulador de 10 rondas → v2 (B) genoma de la final + detección
  de la final → v3 (A) política nueva: equidad contra rangos, tablas de valor, juego según el stack → v4 (B)
  modo final completo: reloj, valor de la final, mesa corta → v5 modelado de rivales (lo primero que se
  recorta). Una versión reemplaza a la anterior solo si la supera en las pruebas de promoción sin bajar la
  métrica del otro plan. Congelar el código el vie 20/11, entrega el sáb 21/11, torneo el dom 22/11/2026.
- Medir los cambios con mesas de 5-6 jugadores en condiciones de competencia (10 000 fichas, ciegas 50/100,
  50 manos, `timeout_ms` 5000, `reponer_stack: false`), no con el `torneo-go.json` mano a mano. Cada mesa
  `mesa-unica` corre en ~3 s; correr muchas semillas.

Al terminar una tarea o cambiar una decisión del plan, actualizar las casillas y la sección "Estado" del plan
que corresponda (`PLAN-BOT.md`, `PLAN-CLASIFICATORIA.md` o `PLAN-FINAL.md`), en el mismo commit o PR que el
trabajo.

## Trabajo en equipo (2 personas, 2 agentes)

Los dos agentes no comparten memoria ni conversación: **lo único compartido es el repositorio**. Todo lo que
la otra persona necesite saber (decisiones, resultados, cosas aprendidas, tareas tomadas) se escribe en los
planes, en `Simulaciones/` o en la descripción del PR. No dejar conocimiento solo en el chat.

### Git

- Remoto: `origin` en GitHub (**repositorio privado; no hacerlo público** antes de la entrega). Rama
  principal: `master`.
- **`master` siempre es entregable**: compila, pasa las pruebas y el bot juega. Cada versión promovida se
  etiqueta sobre `master` (`v0`, `v1`...).
- Trabajar en **ramas cortas** que salen de `master` actualizado: `a/<tema>` (plan A), `b/<tema>` (plan B),
  `comun/<tema>` (infraestructura, simulador, docs). Una rama = una tarea del plan; se fusiona en uno o dos
  días, no se deja envejecer.
- Antes de empezar: `git pull --rebase origin master`. Antes de abrir el PR: volver a traer `master` y
  resolver los conflictos en la rama propia.
- Integrar por **pull request** en GitHub. Si el cambio toca archivos compartidos (lista de abajo), lo revisa
  la otra persona antes de fusionar; si solo toca archivos propios de la tarea y las comprobaciones pasan, lo
  puede fusionar quien lo abrió. Los cambios que solo son de documentación o de estado de un plan pueden ir
  directo a `master`.
- Commits chicos y en español, que digan qué cambió y, si es un cambio medido, con qué resultado. Los agentes
  firman con su línea `Co-Authored-By`.
- Nunca `push --force` sobre `master` ni reescribir etiquetas ya subidas.
- Un agente **no hace commit ni push si la persona no se lo pidió**.

### Comprobaciones antes de cada PR

Desde `pokerfight-desarrollo/`:

```bash
go build -o bin/ ./cmd/... ./bots/...
go vet ./bots/lince/...
go test -trimpath ./bots/lince/...
```

Y si el cambio afecta cómo juega o cómo se conecta el bot: mesas en condiciones de torneo
(`./bin/arena correr -config pruebas-lince/mesa5.json`, varias semillas) y, si toca conexión o estado,
`python bots/lince/pruebas/robustez.py`.

### Archivos compartidos (avisar antes de tocarlos)

Los dos planes pasan por los mismos archivos del bot. Para no pisarse: avisar a la otra persona, hacer el
cambio en un PR chico y fusionarlo rápido.

- `bots/lince/estrategia/` (`situacion.go`, `genoma.go`, `politica.go`): la política es común a los dos modos.
- `bots/lince/entrenamiento/simulador.go` y el resto del entrenador: base del simulador de liga y de final.
- `bots/lince/main.go` y `bots/lince/torneo/`: conexión, estado en disco, detección de fase.
- `PLAN-BOT.md` y este `AGENTS.md`.

Los genomas (`genoma*.json`, `preflop.json`) son **salida del entrenador**: no se editan ni se fusionan a
mano. Si hay conflicto, se queda el de la versión promovida. Cada modo tendrá su propio archivo de genoma, así
A y B no chocan.

### Token y servidor de la organización

- El **token del equipo es secreto y es uno solo**: nunca va al repositorio, ni a un log subido, ni a un PR.
  Se pasa por la variable `POKERFIGHT_TOKEN` o por un `.env` local (ignorado por git).
- **Una sola conexión a la vez**: dos conexiones con el mismo token se pisan. Antes de conectar al servidor de
  pruebas, avisar a la otra persona. El día del torneo corre una sola máquina.

### Resultados y entrenamientos

- Cada experimento va en su propio archivo en `Simulaciones/` (así no hay conflictos), con fecha, autor,
  commit, comando y configuración usados. Ver `Simulaciones/README.md`.
- Los entrenamientos largos (noches de máquina) se anotan en el "Estado" del plan: quién entrena qué, para no
  duplicar.

## Cosas que fallan y conviene saber

- **Smart App Control de Windows** (pasa en la PC de Carlos; puede pasar en cualquier Windows 11) bloquea de
  forma impredecible los `.exe` recién compilados ("An Application Control policy has blocked this file"): el
  veredicto cambia con cada compilación, así que el mismo código puede pasar o no (bloqueó `bin/bandido.exe` y
  el binario de `go test`). Después de compilar, comprobar que el `.exe` arranca (`bin/lince.exe -h`); si
  está bloqueado, recompilar con otras opciones (`go test -trimpath`, `-ldflags=-s`). `go run` no es una
  salida segura: también genera un `.exe`. Nunca recompilar el binario final el día del torneo.
- Las mesas de la arena real **no son reproducibles**: la misma `semilla` con los mismos bots da resultados
  distintos, porque algún bot del kit usa azar sin sembrar. Las comparaciones pareadas "con las mismas
  cartas" solo existen en el simulador rápido de Darwin.
- En las configuraciones de la arena, correr nuestro bot con `-silencioso -sin-archivos -una-partida`; sin
  `-una-partida` reconecta para siempre, por diseño.
- El SDK de Go **no** reconecta, aunque el README de Darwin diga lo contrario.
- Compilar siempre con `-o bin/` (con la barra final): en Windows así Go agrega el `.exe`.
