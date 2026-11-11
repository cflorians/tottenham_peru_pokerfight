# AGENTS.md

Guidance for coding agents working in this folder (Codex reads this file; Claude Code gets it through `CLAUDE.md`, which only imports it). The team rules (git workflow, shared files, token) are in `../AGENTS.md`: read that first.

## What this is

The participant starter kit for **PokerFight** (AICS PUCP's poker-bot competition): a local Texas Hold'em No-Limit table server (`mesa`), a round-robin tournament runner (`arena`), a Go bot SDK, and sparring bots. The aim of working here is to build and test a competitive bot; the competition server itself is run by the organizers and is **not** in this repo, but it speaks the same protocol.

- Everything (code identifiers, comments, docs, CLI flags, JSON config keys) is in **Spanish**. Keep that convention.
- This tree was **exported from the organizers' repo** (see `.exportado-de`). Treat `internal/`, `cmd/`, `pkg/` and `docs/` as upstream code: fixes there come from the organizers via `git pull`. If the code and `docs/protocolo.md` disagree, that is an upstream bug to report, not something a bot should exploit (exploiting it disqualifies). Our own bot belongs in `bots/<nombre>/`.
- `../pokerfight-arena-participantes/` is the pristine reference copy of the kit: **never modify, build or run anything inside it**. This folder (`pokerfight-desarrollo`) is the only place to write code. `../Bases_Poker_Fight_de_Bots.docx` and `../PokerFight - Guía de inicio para equipos.docx` are the official rules/guide.
- The working plan for our bot is `../PLAN-BOT.md` (master plan) with `../PLAN-CLASIFICATORIA.md` (plan A) and `../PLAN-FINAL.md` (plan B). Our bot is `bots/lince/` (see its README). Simulation results are in `../Simulaciones/`. Read the plan before starting bot work, and keep its checkboxes and "Estado" section up to date.
- Module path: `github.com/aicspucp/pokerfight-arena-participantes` (Go 1.26+). No external dependencies.

## Commands

Platform is Windows: built binaries are `bin\*.exe`. Always build with `-o bin/` (trailing slash) so Go adds `.exe`; `-o bin/mi-bot` produces a file neither PowerShell nor the arena can run.

```bash
go build -o bin/ ./cmd/... ./bots/...       # build arena, mesa, casino and all sample bots  (make build)
go build -o bin/ ./bots/mi-bot              # build a single bot
go vet ./...
./bin/arena correr -config torneo-go.json   # round-robin among the Go sample bots       (make torneo-local)
./bin/arena correr -config torneo.json      # also includes the Python/JS bots (needs python + node)
./bin/arena validar -config torneo.json     # check config and list pairings without playing
./bin/arena correr -config t.json -verboso  # show bots' stderr
go run ./bots/darwin/cmd/entrenar-darwin    # retrain Darwin's genome (overwrites bots/darwin/genoma.json); -rapido for a quick smoke run in a temp dir
```

The kit itself has **no `_test.go` files**; only our bot does (`go test -trimpath ./bots/lince/...`, plus `python bots/lince/pruebas/robustez.py` against real tables). Beyond that, verification is done by playing: a manual table plus bots, or an arena tournament (`./bin/arena correr -config pruebas-lince/mesa5.json` is a 5-player table in competition conditions).

Manual table (one process per terminal; the arena appends `-addr` and `-token` to a bot's command the same way):

```bash
./bin/mesa -addr :9000 -jugadores 6 -min-jugadores 3 -rondas 200 -timeout-ms 5000
./bin/conservador -addr localhost:9000 -token conservador
./bin/mi-bot      -addr localhost:9000 -token mi-bot
```

`mesa` defaults (`-timeout-ms 2000`, `-stack 1000`, blinds 10/20) are **not** the competition's. To mimic the qualifier, use `-jugadores 5 -stack 10000 -ciega-chica 50 -ciega-grande 100 -rondas 50 -timeout-ms 5000`.

Add a bot to a tournament by appending to `participantes` in `torneo-go.json`: `{ "id": "mi-bot", "nombre": "Mi Bot", "comando": ["./bin/mi-bot"] }`. Unknown/misspelled JSON keys make loading fail on purpose. Results and per-hand history (with everyone's hole cards, JSON Lines) go to `resultados/`; that history is the main debugging tool.

## Architecture

Dependency rule: **`internal/protocolo` is the only package shared by everything**. Mesa, crupier and bots never import each other's internals; they communicate through protocolo types. Casino never imports mesa (and vice versa): the `cmd/*` binaries are the only orchestrators that know both, injecting `mesa.ValidarToken` and converting `mesa.ResumenPartida` into casino results.

- `internal/protocolo` – wire contract: JSON Lines over TCP (`codec.go`, with max line sizes), message types, cards, actions. `VersionProtocolo` must match the bot's `saludo` exactly. `AccionSegura` = check if free, else fold. `docs/protocolo.md` is the normative spec.
- `internal/crupier` – deck (`baraja.go`: crypto/rand normally, seeded source for reproducible tournaments), main/side pots (`pozo.go`), hand evaluation. `crupier.Evaluacion` is an alias of `pkg/poker.Evaluacion`.
- `internal/mesa` – runs a game: blinds, button, betting rounds, chip accounting, showdown. `mesa.go` is the game loop (`Jugar` → `jugarMano` → `rondaApuestas` → `turnoJugador` → `repartirPozo`); all mutable table state is guarded by `Mesa.Mu` (functions in `saldo.go` assume it's held; `*SinLock` helpers too). `turnos.go` distinguishes `Jugador.Activo` (still has chips in the game) from `Jugador.EnMano` (dealt into the current hand). `servidor.go` is the TCP server: handshake, per-IP handshake limits, reconnect-by-token (the same token replaces the old socket and keeps the seat/chips), timeouts. Connections go through the `ConexionMesaJugador` interface; `ConexionPrueba` is an in-memory implementation for deterministic tests.
- `internal/arena` – tournament runner. Extension points are registries chosen by name from the config: `Formato` (`RegistrarFormato`; `round-robin`, `mesa-unica`), `Puntuador` (`RegistrarPuntuador`; `neto`, `puestos`, `victorias`; must be a pure function because parallel matches finish in arbitrary order), and `Lanzador` (subprocess vs. remote bot). `ida_y_vuelta` replays each pairing with swapped seats and the same seed; `semilla_secreta` keeps deals reproducible only for whoever holds the secret. In open mode each launched bot gets a fresh secret token per match.
- `internal/casino` + `internal/almacen` – accounts, sessions, bot versioning, historical ranking, persisted as a JSON file. Only used with `-casino-db`.
- `pkg/botsdk` – public Go SDK. A bot implements `Nombre()` and `Decidir(ctx, Turno) Accion`, then calls `botsdk.CorrerDesdeFlags(bot)`. Optional hooks are detected by interface: `Inicializable`, `ObservaManos`, `ObservaEstado`. `tipos.go` re-exports protocolo types as aliases because `internal/` can't be imported from outside the module; any protocol type a bot must name has to be aliased there too.
- `pkg/poker` – the exact hand evaluator the table uses (`poker.Mejor(mano, comunitarias)`); bots should use it rather than reimplement.

Sample bots (`bots/`) each demonstrate one technique: `aleatorio`, `conservador` (minimal SDK example), `base` (Chen formula), `estadistico` (Monte Carlo equity), `bandido` (UCB1 bandit), `espia` (Bayesian opponent modelling), `darwin` (genetic algorithm; `estrategia/` is the runtime policy, `entrenamiento/` + `cmd/entrenar-darwin` the offline trainer, `genoma.json`/`preflop.json` its trained output), plus Python (`novato`, `ejemplos/python`) and JS (`ejemplos/javascript`) bots with no dependencies.

## Bot rules that are easy to get wrong

- `monto` in an action is the **total** bet for the betting round, not the increment. Prefer the `Turno` builders (`Pagar`, `Apostar(total)`, `ApostarPozo(frac)`, `AllIn`) which clamp to legal amounts; only send actions listed in `acciones_validas`.
- An invalid action costs exactly the same as a timeout (safe action played, timeout counted, which hurts ranking). Deadline is 5 s in competition; aim to answer in 2–3 s.
- Echo the request's `id_mano` in every `accion`; never reply to `estado`/`mano_fin`. A late reply can be taken as the answer to the next request in the same hand.
- Identify yourself by the `id_jugador` from `bienvenida`, not by token/name. In competition, opponents are random but stable aliases (`j-xxxxxx`) across rounds.
- Bots with custom flags must define them before `CorrerDesdeFlags` and **must not call `flag.Parse`**; read flag values in `Iniciar()`.
- The Go SDK does **not** reconnect: `botsdk.Correr` returns on disconnect, so wrap it in a retry loop (example in `docs/guia-del-bot.md#reconectar`). EOF is the normal end of a game.
- Competition format (`docs/competencia.md`): 10 000 starting chips, fixed blinds 50/100, tables of 5 (some 6), up to 50 hands or 6 minutes per table, 10 rounds with chips carried over and re-drawn tables, busting out eliminates. One connection lasts the whole tournament and stays **silent between rounds**; it is not a drop and must not be closed. Top 6 by chips go to a final table (10 000 chips, last standing or most chips after 1 h).
