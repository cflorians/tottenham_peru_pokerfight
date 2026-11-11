# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Workspace for our team's entry in **PokerFight** (AICS PUCP poker-bot competition). The team works in Spanish: answer in Spanish and keep code, comments and docs in Spanish.

## Folder rules (strict)

| Folder | Rule |
| --- | --- |
| `pokerfight-desarrollo/` | **The only folder where code is written.** Our bot is `pokerfight-desarrollo/bots/lince/` (working name; git branch `lince`, one tag per version: `v0`, `v1`...; see its README), and our tools in their own new folders. Inside it, still don't edit the kit itself (`internal/`, `cmd/`, `pkg/`, `docs/`, the sample bots): the real server doesn't use our copy. The only kit file we may edit is `torneo-go.json`, to add our bot to local tests. See `pokerfight-desarrollo/CLAUDE.md`. |
| `pokerfight-arena-participantes/` | **Never touch.** It's the untouched reference copy of the organizers' kit. No edits, no builds, no `bin/` or `resultados/` inside it. Only read it, e.g. to diff against `pokerfight-desarrollo`. |
| `Simulaciones/` | Simulation results (markdown) and test artifacts such as `genoma_prudente.json`. Add new results here. |
| `Bases_Poker_Fight_de_Bots.docx`, `PokerFight - Guía de inicio para equipos.docx` | Official rules and team guide. Read-only. To get their text without dependencies, unzip `word/document.xml` and extract the `<w:t>` elements. |

## The plan

**[`PLAN-BOT.md`](PLAN-BOT.md) is the working plan: read it before starting any bot work.** In short:

- Base: copy of **Gato Darwin** (`bots/darwin`, Go, genetic algorithm + Monte Carlo), plus opponent modelling from **Michi Espía**.
- Goal is the grand prize (winning the final table): P(champion) = P(reach the final) × P(win the final). They are two different games: the qualifier (tables of 5-6, chips carried over 10 rounds, busting eliminates, top 6 by chips or the last ones eliminated go through) rewards **survival plus enough accumulation**; the final (6 bots, 10 000 chips, one winner) rewards risk-neutral accumulation. Original Darwin busts in 44 % of 5-player tables; a hand-tuned risk-averse genome brings that to 10 % but folds winning hands to big all-ins, so it is only the starting point (`Simulaciones/pruebas_formato_real.md`).
- `PLAN-BOT.md` is the master plan (shared infrastructure, simulator, calendar, version ladder, questions for the organizers, risks). The strategy lives in two plans that end up as two modes of the same bot (only one bot per team): [`PLAN-CLASIFICATORIA.md`](PLAN-CLASIFICATORIA.md) (plan A: reach the top 6 and fight for 1st place in the qualifier, which has its own prize) and [`PLAN-FINAL.md`](PLAN-FINAL.md) (plan B: win the final table). Read the one that matches the work. Expected prize = 60 × P(final) × P(win final) + 30 × P(1st in qualifier).
- Version ladder, always keeping a deliverable bot, cheap wins of each plan first: v0 Darwin + prudent genome + infrastructure (reconnection, supervisor, on-disk state, logging) → v1 (A) current genes retrained with league fitness on our own 10-round league simulator → v2 (B) final-table genome + final detection → v3 (A) new policy: equity against ranges, value tables, stack-depth play → v4 (B) full final mode: clock, final value, short-handed play → v5 opponent modelling (first thing to cut). A version replaces the previous one only if it beats it in the promotion tests without lowering the other plan's metric. Freeze code on Fri 20/11, delivery Sat 21/11, tournament Sun 22/11/2026.
- Measure changes with 5-6 player tables in competition conditions (10 000 chips, blinds 50/100, 50 hands, `timeout_ms` 5000, `reponer_stack: false`), not with the heads-up `torneo-go.json`. Each `mesa-unica` table runs in ~3 s; run many seeds.

When a task is done or a plan decision changes, update the checkboxes and the "Estado" section of the plan it belongs to (`PLAN-BOT.md`, `PLAN-CLASIFICATORIA.md` or `PLAN-FINAL.md`).

## Known pitfalls on this machine

- Windows Smart App Control blocks freshly built `.exe` files unpredictably ("An Application Control policy has blocked this file"): the verdict changes with every build, so the same code may pass or not (it blocked `bin/bandido.exe` and our `go test` binary). After building, check that the `.exe` starts (`bin/lince.exe -h`); if it is blocked, rebuild with other options (`go test -trimpath`, `-ldflags=-s`). `go run` is not a safe fallback: it also produces an `.exe`. Never rebuild the final binary on tournament day.
- Real-arena tables are **not reproducible**: the same `semilla` with the same bots gives different results, because some kit bot uses unseeded randomness. Paired "same cards" comparisons only exist in Darwin's fast simulator.
- Run our bot in arena configs with `-silencioso -sin-archivos -una-partida`; without `-una-partida` it reconnects forever by design.
- The Go SDK does **not** reconnect, despite what Darwin's README says.
