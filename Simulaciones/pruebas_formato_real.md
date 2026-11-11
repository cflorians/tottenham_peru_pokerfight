# Pruebas en el formato real (mesas de 5-6, una ronda de la clasificatoria)

Sin modificar código del kit. Configuraciones `mesa-unica` de la arena con las
condiciones de una ronda de la clasificatoria:

- 10 000 fichas, ciegas 50/100 (100 ciegas grandes), 50 manos, plazo 5 s
- `reponer_stack: false`: quien se queda sin fichas queda fuera (neto = -10 000)
- 30 semillas por alineación, 90 mesas en total

Alineaciones:

| | Bots |
| - | ---- |
| A | darwin, espia, estadistico, base, novato |
| B | darwin, espia, estadistico, novato, conservador |
| C | darwin, espia, estadistico, base, aleatorio, conservador (6) |

`bandido` no pudo jugar: Windows (Smart App Control) bloquea `bin/bandido.exe`
("An Application Control policy has blocked this file"), incluso recompilado.

## Prueba 1: bots del kit tal como vienen

| Bot | Mesas | Neto medio | Mediana | % mesas en que quiebra |
| --- | ----: | ---------: | ------: | ---------------------: |
| novato | 60 | +4 301 | +2 622 | **10 %** |
| espia | 90 | +3 073 | -146 | 36 % |
| darwin | 90 | +2 969 | -2 641 | **44 %** |
| estadistico | 90 | +988 | -1 048 | 41 % |
| base | 60 | -3 581 | -10 000 | 60 % |
| conservador | 60 | -6 729 | -6 742 | 5 % |
| aleatorio | 30 | -9 072 | -10 000 | 97 % |

Primer puesto por mesa: darwin 31/90, espia 20/90, novato 16/60, estadistico 16/90.

Por alineación:

| Bot | A: neto / quiebra | B: neto / quiebra | C: neto / quiebra |
| --- | ----------------: | ----------------: | ----------------: |
| darwin | +3 954 / 40 % | -829 / 50 % | +5 782 / 43 % |
| espia | -2 016 / 50 % | +1 699 / 40 % | +9 537 / 17 % |
| estadistico | -730 / 40 % | +1 231 / 37 % | +2 461 / 47 % |
| novato | +4 241 / 13 % | +4 361 / 7 % | — |

## Prueba 2: darwin con un genoma prudente (`genoma_prudente.json`)

Mismas 90 mesas, mismas semillas, cambiando solo darwin por
`./bin/darwin -genoma genoma_prudente.json`. Respecto del genoma original cambian 4 genes:
`aversion_riesgo` 0,0125 → 0,35, `umbral_allin` 0,89 → 0,95, `farol_pre` 0,28 → 0,08,
`farol_post` 0,18 → 0,08.

| Bot | Neto medio | Mediana | % quiebra |
| --- | ---------: | ------: | --------: |
| estadistico | +3 098 | +282 | 31 % |
| **darwin prudente** | **+2 579** | **+386** | **10 %** |
| novato | +2 548 | +3 088 | 20 % |
| espia | +942 | +659 | 31 % |
| base | -1 758 | -10 000 | 52 % |
| conservador | -6 264 | -6 244 | 8 % |
| aleatorio | -8 907 | -10 000 | 97 % |

Con un ajuste a ojo, el % de quiebra bajó de 44 % a 10 % y el neto medio casi no
cambió. Con 10 rondas y eliminación, sobrevivir las 10 rondas pasa de ~0,3 %
(0,56¹⁰) a ~35 % (0,90¹⁰), si cada ronda fuera independiente.

## Por qué las pruebas iniciales (`pruebas_iniciales.md`) no alcanzan

Usan `torneo-go.json`: heads-up, 4 000 fichas con ciegas 10/20 (200 ciegas grandes),
fichas repuestas en cada mano y sin eliminación. Además, `aleatorio` pierde entre
93 000 y 209 000 fichas por torneo, y ese dinero se reparte entre los demás:
gran parte del ranking mide quién le saca más al bot aleatorio. Sirven para ver
la habilidad heads-up, pero no reflejan la clasificatoria. En la Prueba 1 inicial,
darwin y aleatorio tienen 11 cruces jugados en vez de 12: uno no se jugó.
