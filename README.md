# Tottenham Perú · PokerFight 2026

Repositorio del equipo para el **Poker Fight de Bots** de AICS PUCP: una competencia de bots de Texas Hold'em
No-Limit. Aquí están nuestro bot (**Lince**, nombre de trabajo), el kit de la organización, los planes y los
resultados de las simulaciones.

**Repositorio privado.** No hacerlo público antes de la entrega y no subir nunca el token del equipo.

| | |
| --- | --- |
| Sesión de dudas | dom 15/11/2026 |
| Congelar el código | vie 20/11/2026 |
| **Entrega** | **sáb 21/11/2026** |
| Torneo | dom 22/11/2026 |
| Versión actual del bot | **v0** (Gato Darwin + genoma prudente + infraestructura) |
| Próximo paso | Fase 1: simulador de liga y de mesa final (ver [`PLAN-BOT.md`](PLAN-BOT.md)) |

## La competencia en dos líneas

- **Clasificatoria**: mesas de 5 (algunas de 6), 10 000 fichas, ciegas fijas 50/100, 10 rondas de 50 manos
  con las fichas arrastradas de una ronda a la otra. Quedarse sin fichas es quedar eliminado. Pasan los 6 con
  más fichas. El 1.º tiene premio propio (S/. 30).
- **Mesa final**: los 6 clasificados, 10 000 fichas cada uno, gana el último en pie o el que tenga más fichas
  tras 1 hora (S/. 60).

Son dos juegos distintos (sobrevivir y acumular contra ganar a todo o nada), y solo se permite un bot por
equipo: por eso el bot tiene dos modos y hay dos planes.

## Qué hay en el repositorio

| Ruta | Qué es | ¿Se edita? |
| --- | --- | --- |
| [`PLAN-BOT.md`](PLAN-BOT.md) | **Plan maestro**: estado, reparto del equipo, escalera de versiones, calendario, simulador, preguntas a la organización, riesgos, día del torneo | Sí: mantener casillas y "Estado" |
| [`PLAN-CLASIFICATORIA.md`](PLAN-CLASIFICATORIA.md) | Plan A: entrar entre los 6 y pelear el 1.º de la clasificatoria | Sí |
| [`PLAN-FINAL.md`](PLAN-FINAL.md) | Plan B: ganar la mesa final | Sí |
| [`AGENTS.md`](AGENTS.md) | Reglas para los agentes de código (Codex y Claude Code) y **flujo de trabajo en equipo** | Sí, con aviso al otro |
| [`CLAUDE.md`](CLAUDE.md) | Solo importa `AGENTS.md` para Claude Code | No |
| [`pokerfight-desarrollo/`](pokerfight-desarrollo/) | Copia de trabajo del kit: arena, mesa, SDK de Go, bots de ejemplo. **La única carpeta donde se escribe código** | Solo lo nuestro |
| [`pokerfight-desarrollo/bots/lince/`](pokerfight-desarrollo/bots/lince/) | **Nuestro bot** (ver su [README](pokerfight-desarrollo/bots/lince/README.md)) | Sí |
| `pokerfight-desarrollo/pruebas-lince/` | Configuraciones de la arena para probar el bot en condiciones de torneo | Sí |
| `pokerfight-desarrollo/docs/` | Documentación del kit: protocolo, reglas, formato de la competencia, guía del bot | No (es del kit) |
| [`pokerfight-arena-participantes/`](pokerfight-arena-participantes/) | Copia de referencia intacta del kit de la organización | **Nunca** |
| [`Simulaciones/`](Simulaciones/) | Resultados de pruebas y simulaciones (ver su [README](Simulaciones/README.md)) | Sí: un archivo por experimento |
| `Bases_Poker_Fight_de_Bots.docx`, `PokerFight - Guía de inicio para equipos.docx` | Bases oficiales y guía para equipos | No |

## Empezar (para quien recién clona)

Hace falta **Go 1.26 o más nuevo** y, para la prueba de robustez, **Python 3**. Sin otras dependencias.

```bash
git clone https://github.com/cflorians/tottenham_peru_pokerfight.git
```

```bash
cd tottenham_peru_pokerfight/pokerfight-desarrollo
```

Compilar la arena y todos los bots (siempre con `-o bin/`, con la barra final):

```bash
go build -o bin/ ./cmd/... ./bots/...
```

Pruebas de nuestro bot:

```bash
go test -trimpath ./bots/lince/...
```

Una mesa de 5 en condiciones de torneo (Lince contra Darwin, Espía, Estadístico y Base; ~3 s):

```bash
./bin/arena correr -config pruebas-lince/mesa5.json
```

Prueba de robustez contra mesas reales (silencio, reconexión, conteo de rondas, detección de la final; ~2 min):

```bash
python bots/lince/pruebas/robustez.py
```

Si la mesa imprime una tabla de posiciones y las pruebas pasan, todo funciona. En Windows 11, si un `.exe`
recién compilado no arranca ("An Application Control policy has blocked this file"), es Smart App Control:
ver "Cosas que fallan" en [`AGENTS.md`](AGENTS.md).

Después, leer en este orden: [`PLAN-BOT.md`](PLAN-BOT.md) → el plan que te toque
([A](PLAN-CLASIFICATORIA.md) o [B](PLAN-FINAL.md)) → el [README de Lince](pokerfight-desarrollo/bots/lince/README.md)
→ `pokerfight-desarrollo/docs/competencia.md` y `docs/protocolo.md`.

## Cómo trabajamos (2 personas, 2 agentes)

Una persona trabaja con **Claude Code** y la otra con **Codex**. Los dos agentes leen las mismas reglas:
[`AGENTS.md`](AGENTS.md) en la raíz y [`pokerfight-desarrollo/AGENTS.md`](pokerfight-desarrollo/AGENTS.md)
para el kit. Los `CLAUDE.md` solo importan esos archivos, así no hay dos versiones de las reglas.

El detalle está en `AGENTS.md`; lo esencial:

1. **`master` siempre es entregable.** Cada versión promovida lleva su etiqueta (`v0`, `v1`...).
2. **Ramas cortas por tarea**: `a/<tema>` (plan A), `b/<tema>` (plan B), `comun/<tema>`. Se integran por
   pull request; si tocan archivos compartidos (política, simulador, conexión, plan maestro), revisa la otra
   persona.
3. **Antes de cada PR**: compilar, `go vet` y `go test` del bot; si cambia el juego, mesas de torneo.
4. **Lo que no está en el repositorio no existe**: los agentes no comparten memoria. Decisiones y avances van
   al "Estado" del plan; resultados, a `Simulaciones/`.
5. **El token es uno solo y es secreto**: va en la variable `POKERFIGHT_TOKEN`, nunca en git. Una sola
   conexión a la vez: avisar antes de usar el servidor de pruebas.
6. **Reparto**: una persona lleva el Plan A y la otra el Plan B; el simulador y el ensayo general son
   comunes. La tabla está en la sección 4 de [`PLAN-BOT.md`](PLAN-BOT.md).

## El bot en una línea

Lince parte de **Gato Darwin** (bot de ejemplo del kit, de AICS PUCP): una política cuyos números evoluciona
un algoritmo genético, con la fuerza de la mano estimada por Monte Carlo. Sobre eso agregamos reconexión,
supervisor, estado en disco y log (v0), y el plan suma un genoma por modo, política nueva y modelado de
rivales. La escalera completa está en [`PLAN-BOT.md`](PLAN-BOT.md).
