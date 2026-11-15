# Ensayo general: 15 ligas reales completas de 30 bots

- **Fecha**: dom 15/11/2026 · **Quién**: Carlos (con Claude Code; el orquestador lo escribió un subagente)
- **Herramienta**: `bots/lince/cmd/liga-real` (nueva): clasificatoria de 10 rondas con fichas arrastradas y mesa final, con los ejecutables de verdad y el servidor de mesa del kit. Lince es un solo proceso durante toda la liga.
- **Comando** (desde `pokerfight-desarrollo/`): `./bin/liga-real.exe -n 30 -ligas 15 -semilla 500 -paralelo 4`
- **Campo**: Lince v1.2 más 29 bots del kit repartidos parejo entre sus tipos.

## Conclusiones

| Medida | Arena real (15 ligas) | Simulador (campo kit, 30 inscritos) |
| --- | ---: | ---: |
| Lince llega a la final | 12 de 15 (80 %) | 86 % |
| Lince termina 1.º de la clasificatoria | 3 de 15 (20 %) | 26 % |
| Lince gana la final, estando en ella | 10 de 12 (83 %) | 87 % sin tope |
| **Lince campeón** | **10 de 15 (67 %)** | ~75 % |

- El simulador y la arena real coinciden también con fichas arrastradas, dentro del margen de 15 ligas (80 % es en realidad 55-93 %). El campo del simulador no es idéntico (otras proporciones de bots).
- **Ensayo del bot: 15 de 15.** En todas las ligas Lince contó bien las rondas y detectó la final en el momento justo. Cero timeouts, ningún pánico ni error en su log.
- Quiebras de Lince: una en la ronda 1, dos en la 3 y una en la 7.
- Una liga completa de 30 bots tarda 62 segundos.

## Límites

- Aquí la conexión se cierra entre rondas y Lince reconecta; en el torneo queda abierta y en silencio. Eso lo cubre la prueba de silencio, que sigue pendiente.
- El tope de 6 minutos por mesa nunca se activó: no está probado.
- Los rivales son solo bots del kit.

## Lo que no dio mejora (mismo día)

Reentrenar la clasificatoria por el premio del bot entero (final jugada con el genoma adaptativo) y con un campo "realista" (el mixto más 4 bots entrenados de 29): `genoma_clasificatoria_r.json`. Llega algo más a la final en kit (+2,4 puntos) y fuerte (+4,5), termina 1.º menos y pierde contra el campo agresivo; en el campo realista queda igual (65,8 % contra 65,0 %). La clasificatoria sigue con la v1.1.

## Salida completa

# Liga real: 30 participantes, 10 rondas de 50 manos, semilla 500

Campo: Lince + 29 bots del kit (darwin, espia, estadistico, base, conservador, aleatorio, novato). Reloj por mesa 6m0s, plazo 5000 ms, 4 mesas en paralelo.

## Liga 1 (semilla 500)

| Ronda | Mesas | Con fichas antes | Con fichas después | Fichas de Lince | Timeouts Lince | Suma de fichas | Duración |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| 1 | 6 | 30 | 22 | 12473 | 0 | ok | 5 s |
| 2 | 4 | 22 | 15 | 20806 | 0 | ok | 3 s |
| 3 | 3 | 15 | 9 | 24775 | 0 | ok | 4 s |
| 4 | 2 | 9 | 7 | 32494 | 0 | ok | 7 s |
| 5 | 2 | 7 | 6 | 63085 | 0 | ok | 7 s |
| 6 | 1 | 6 | 5 | 63558 | 0 | ok | 6 s |
| 7 | 1 | 5 | 5 | 134879 | 0 | ok | 6 s |
| 8 | 1 | 5 | 4 | 133336 | 0 | ok | 7 s |
| 9 | 1 | 4 | 4 | 140334 | 0 | ok | 6 s |
| 10 | 1 | 4 | 3 | 138129 | 0 | ok | 7 s |

Lince terminó 1.º de 30 en la clasificatoria.

Clasificados a la final:

| Puesto | Jugador | Fichas o caída |
| ---: | --- | --- |
| 1 | **Lince** | 138129 |
| 2 | darwin-4 | 84428 |
| 3 | novato-2 | 77443 |
| 4 | novato-3 | eliminado en la ronda 10 |
| 5 | novato-4 | eliminado en la ronda 8 |
| 6 | espia-4 | eliminado en la ronda 6 |

**Final**: ganó **Lince** en 1192 manos (23 s). Orden: **Lince**, espia-4, novato-3, novato-4, novato-2, darwin-4.

Timeouts de Lince en toda la liga: 0. Duración de la liga: 81 s.

**Chequeo del estado de Lince** (estado.json y log):

- Cree estar en la ronda 11, fase final (jugador lince). De verdad jugó 11 mesas y la última fue de la final.
- Fases por ronda según Lince: clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, final.
- Coincide: rondas, fases, fichas al empezar cada ronda y jugadores por mesa.
- Log: 0 líneas con PÁNICO y 0 con ERROR.

## Liga 2 (semilla 501)

| Ronda | Mesas | Con fichas antes | Con fichas después | Fichas de Lince | Timeouts Lince | Suma de fichas | Duración |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| 1 | 6 | 30 | 21 | 11313 | 0 | ok | 6 s |
| 2 | 4 | 21 | 14 | 9201 | 0 | ok | 4 s |
| 3 | 3 | 14 | 7 | **0 (cae aquí)** | 0 | ok | 4 s |
| 4 | 2 | 7 | 7 | — | 0 | ok | 2 s |
| 5 | 2 | 7 | 4 | — | 0 | ok | 2 s |
| 6 | 1 | 4 | 4 | — | 0 | ok | 2 s |
| 7 | 1 | 4 | 3 | — | 0 | ok | 2 s |
| 8 | 1 | 3 | 2 | — | 0 | ok | 2 s |
| 9 | 1 | 2 | 1 | — | 0 | ok | 0 s |
| 10 | 0 | 1 | 1 | — | 0 | ok | 0 s |

Lince cayó en la ronda 3. Lince terminó 11.º de 30 en la clasificatoria.

Clasificados a la final:

| Puesto | Jugador | Fichas o caída |
| ---: | --- | --- |
| 1 | darwin-4 | 300000 |
| 2 | darwin-2 | eliminado en la ronda 9 |
| 3 | novato-3 | eliminado en la ronda 8 |
| 4 | espia-2 | eliminado en la ronda 7 |
| 5 | novato-4 | eliminado en la ronda 5 |
| 6 | novato-1 | eliminado en la ronda 5 |

**Final**: ganó darwin-4 en 477 manos (9 s). Orden: darwin-4, novato-3, novato-1, darwin-2, espia-2, novato-4.

Timeouts de Lince en toda la liga: 0. Duración de la liga: 33 s.

**Chequeo del estado de Lince** (estado.json y log):

- Cree estar en la ronda 3, fase clasificatoria (jugador lince). De verdad jugó 3 mesas y la última fue de la clasificatoria.
- Fases por ronda según Lince: clasificatoria, clasificatoria, clasificatoria.
- Coincide: rondas, fases, fichas al empezar cada ronda y jugadores por mesa.
- Log: 0 líneas con PÁNICO y 0 con ERROR.

## Liga 3 (semilla 502)

| Ronda | Mesas | Con fichas antes | Con fichas después | Fichas de Lince | Timeouts Lince | Suma de fichas | Duración |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| 1 | 6 | 30 | 17 | 3005 | 0 | ok | 3 s |
| 2 | 3 | 17 | 10 | 1171 | 0 | ok | 4 s |
| 3 | 2 | 10 | 6 | **0 (cae aquí)** | 0 | ok | 1 s |
| 4 | 1 | 6 | 4 | — | 0 | ok | 1 s |
| 5 | 1 | 4 | 4 | — | 0 | ok | 0 s |
| 6 | 1 | 4 | 3 | — | 0 | ok | 0 s |
| 7 | 1 | 3 | 3 | — | 0 | ok | 0 s |
| 8 | 1 | 3 | 3 | — | 0 | ok | 0 s |
| 9 | 1 | 3 | 2 | — | 0 | ok | 0 s |
| 10 | 1 | 2 | 2 | — | 0 | ok | 0 s |

Lince cayó en la ronda 3. Lince terminó 8.º de 30 en la clasificatoria.

Clasificados a la final:

| Puesto | Jugador | Fichas o caída |
| ---: | --- | --- |
| 1 | espia-1 | 276794 |
| 2 | novato-2 | 23206 |
| 3 | novato-1 | eliminado en la ronda 9 |
| 4 | novato-3 | eliminado en la ronda 6 |
| 5 | espia-4 | eliminado en la ronda 4 |
| 6 | darwin-5 | eliminado en la ronda 4 |

**Final**: ganó espia-1 en 791 manos (4 s). Orden: espia-1, novato-2, novato-3, novato-1, darwin-5, espia-4.

Timeouts de Lince en toda la liga: 0. Duración de la liga: 15 s.

**Chequeo del estado de Lince** (estado.json y log):

- Cree estar en la ronda 3, fase clasificatoria (jugador lince). De verdad jugó 3 mesas y la última fue de la clasificatoria.
- Fases por ronda según Lince: clasificatoria, clasificatoria, clasificatoria.
- Coincide: rondas, fases, fichas al empezar cada ronda y jugadores por mesa.
- Log: 0 líneas con PÁNICO y 0 con ERROR.

## Liga 4 (semilla 503)

| Ronda | Mesas | Con fichas antes | Con fichas después | Fichas de Lince | Timeouts Lince | Suma de fichas | Duración |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| 1 | 6 | 30 | 18 | 31776 | 0 | ok | 5 s |
| 2 | 3 | 18 | 9 | 33743 | 0 | ok | 3 s |
| 3 | 2 | 9 | 7 | 49850 | 0 | ok | 3 s |
| 4 | 2 | 7 | 6 | 46526 | 0 | ok | 3 s |
| 5 | 1 | 6 | 6 | 63835 | 0 | ok | 6 s |
| 6 | 1 | 6 | 6 | 59671 | 0 | ok | 6 s |
| 7 | 1 | 6 | 5 | 60265 | 0 | ok | 6 s |
| 8 | 1 | 5 | 5 | 62250 | 0 | ok | 7 s |
| 9 | 1 | 5 | 5 | 52013 | 0 | ok | 7 s |
| 10 | 1 | 5 | 5 | 55624 | 0 | ok | 7 s |

Lince terminó 3.º de 30 en la clasificatoria.

Clasificados a la final:

| Puesto | Jugador | Fichas o caída |
| ---: | --- | --- |
| 1 | espia-1 | 131949 |
| 2 | novato-4 | 55979 |
| 3 | **Lince** | 55624 |
| 4 | novato-2 | 41154 |
| 5 | darwin-2 | 15294 |
| 6 | novato-3 | eliminado en la ronda 7 |

**Final**: ganó **Lince** en 576 manos (22 s). Orden: **Lince**, novato-4, darwin-2, novato-3, novato-2, espia-1.

Timeouts de Lince en toda la liga: 0. Duración de la liga: 74 s.

**Chequeo del estado de Lince** (estado.json y log):

- Cree estar en la ronda 11, fase final (jugador lince). De verdad jugó 11 mesas y la última fue de la final.
- Fases por ronda según Lince: clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, final.
- Coincide: rondas, fases, fichas al empezar cada ronda y jugadores por mesa.
- Log: 0 líneas con PÁNICO y 0 con ERROR.

## Liga 5 (semilla 504)

| Ronda | Mesas | Con fichas antes | Con fichas después | Fichas de Lince | Timeouts Lince | Suma de fichas | Duración |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| 1 | 6 | 30 | 18 | 10641 | 0 | ok | 6 s |
| 2 | 3 | 18 | 10 | 14049 | 0 | ok | 4 s |
| 3 | 2 | 10 | 9 | 17644 | 0 | ok | 3 s |
| 4 | 2 | 9 | 6 | 17175 | 0 | ok | 3 s |
| 5 | 1 | 6 | 6 | 11934 | 0 | ok | 5 s |
| 6 | 1 | 6 | 5 | 22196 | 0 | ok | 6 s |
| 7 | 1 | 5 | 5 | 31762 | 0 | ok | 6 s |
| 8 | 1 | 5 | 5 | 33119 | 0 | ok | 6 s |
| 9 | 1 | 5 | 5 | 21213 | 0 | ok | 6 s |
| 10 | 1 | 5 | 5 | 22154 | 0 | ok | 6 s |

Lince terminó 4.º de 30 en la clasificatoria.

Clasificados a la final:

| Puesto | Jugador | Fichas o caída |
| ---: | --- | --- |
| 1 | darwin-5 | 166675 |
| 2 | novato-2 | 62424 |
| 3 | novato-3 | 33549 |
| 4 | **Lince** | 22154 |
| 5 | novato-1 | 15198 |
| 6 | base-4 | eliminado en la ronda 6 |

**Final**: ganó **Lince** en 1247 manos (12 s). Orden: **Lince**, novato-3, novato-2, darwin-5, novato-1, base-4.

Timeouts de Lince en toda la liga: 0. Duración de la liga: 64 s.

**Chequeo del estado de Lince** (estado.json y log):

- Cree estar en la ronda 11, fase final (jugador lince). De verdad jugó 11 mesas y la última fue de la final.
- Fases por ronda según Lince: clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, final.
- Coincide: rondas, fases, fichas al empezar cada ronda y jugadores por mesa.
- Log: 0 líneas con PÁNICO y 0 con ERROR.

## Liga 6 (semilla 505)

| Ronda | Mesas | Con fichas antes | Con fichas después | Fichas de Lince | Timeouts Lince | Suma de fichas | Duración |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| 1 | 6 | 30 | 19 | 15219 | 0 | ok | 6 s |
| 2 | 4 | 19 | 12 | 54151 | 0 | ok | 6 s |
| 3 | 2 | 12 | 9 | 112588 | 0 | ok | 6 s |
| 4 | 2 | 9 | 7 | 108326 | 0 | ok | 5 s |
| 5 | 2 | 7 | 6 | 109228 | 0 | ok | 7 s |
| 6 | 1 | 6 | 3 | 99002 | 0 | ok | 6 s |
| 7 | 1 | 3 | 3 | 202195 | 0 | ok | 6 s |
| 8 | 1 | 3 | 3 | 201163 | 0 | ok | 6 s |
| 9 | 1 | 3 | 2 | 252857 | 0 | ok | 5 s |
| 10 | 1 | 2 | 2 | 256618 | 0 | ok | 5 s |

Lince terminó 1.º de 30 en la clasificatoria.

Clasificados a la final:

| Puesto | Jugador | Fichas o caída |
| ---: | --- | --- |
| 1 | **Lince** | 256618 |
| 2 | novato-3 | 43382 |
| 3 | espia-4 | eliminado en la ronda 9 |
| 4 | darwin-4 | eliminado en la ronda 6 |
| 5 | novato-1 | eliminado en la ronda 6 |
| 6 | darwin-2 | eliminado en la ronda 6 |

**Final**: ganó novato-1 en 2000 manos (19 s). Orden: novato-1, **Lince**, espia-4, darwin-4, novato-3, darwin-2.

Timeouts de Lince en toda la liga: 0. Duración de la liga: 78 s.

**Chequeo del estado de Lince** (estado.json y log):

- Cree estar en la ronda 11, fase final (jugador lince). De verdad jugó 11 mesas y la última fue de la final.
- Fases por ronda según Lince: clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, final.
- Coincide: rondas, fases, fichas al empezar cada ronda y jugadores por mesa.
- Log: 0 líneas con PÁNICO y 0 con ERROR.

## Liga 7 (semilla 506)

| Ronda | Mesas | Con fichas antes | Con fichas después | Fichas de Lince | Timeouts Lince | Suma de fichas | Duración |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| 1 | 6 | 30 | 15 | 18281 | 0 | ok | 3 s |
| 2 | 3 | 15 | 11 | 41657 | 0 | ok | 4 s |
| 3 | 2 | 11 | 8 | 41650 | 0 | ok | 6 s |
| 4 | 2 | 8 | 4 | 44506 | 0 | ok | 5 s |
| 5 | 1 | 4 | 3 | 45556 | 0 | ok | 7 s |
| 6 | 1 | 3 | 3 | 42188 | 0 | ok | 7 s |
| 7 | 1 | 3 | 3 | 37370 | 0 | ok | 8 s |
| 8 | 1 | 3 | 3 | 35189 | 0 | ok | 8 s |
| 9 | 1 | 3 | 3 | 35013 | 0 | ok | 8 s |
| 10 | 1 | 3 | 3 | 30757 | 0 | ok | 7 s |

Lince terminó 2.º de 30 en la clasificatoria.

Clasificados a la final:

| Puesto | Jugador | Fichas o caída |
| ---: | --- | --- |
| 1 | estadistico-3 | 249714 |
| 2 | **Lince** | 30757 |
| 3 | novato-2 | 19529 |
| 4 | espia-1 | eliminado en la ronda 5 |
| 5 | darwin-5 | eliminado en la ronda 4 |
| 6 | novato-4 | eliminado en la ronda 4 |

**Final**: ganó **Lince** en 1446 manos (16 s). Orden: **Lince**, novato-2, novato-4, estadistico-3, espia-1, darwin-5.

Timeouts de Lince en toda la liga: 0. Duración de la liga: 79 s.

**Chequeo del estado de Lince** (estado.json y log):

- Cree estar en la ronda 11, fase final (jugador lince). De verdad jugó 11 mesas y la última fue de la final.
- Fases por ronda según Lince: clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, final.
- Coincide: rondas, fases, fichas al empezar cada ronda y jugadores por mesa.
- Log: 0 líneas con PÁNICO y 0 con ERROR.

## Liga 8 (semilla 507)

| Ronda | Mesas | Con fichas antes | Con fichas después | Fichas de Lince | Timeouts Lince | Suma de fichas | Duración |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| 1 | 6 | 30 | 16 | 14637 | 0 | ok | 5 s |
| 2 | 3 | 16 | 10 | 22003 | 0 | ok | 4 s |
| 3 | 2 | 10 | 9 | 27385 | 0 | ok | 2 s |
| 4 | 2 | 9 | 8 | 60622 | 0 | ok | 3 s |
| 5 | 2 | 8 | 8 | 64322 | 0 | ok | 4 s |
| 6 | 2 | 8 | 5 | 67050 | 0 | ok | 5 s |
| 7 | 1 | 5 | 4 | 131775 | 0 | ok | 7 s |
| 8 | 1 | 4 | 3 | 135403 | 0 | ok | 7 s |
| 9 | 1 | 3 | 3 | 133330 | 0 | ok | 7 s |
| 10 | 1 | 3 | 3 | 135387 | 0 | ok | 7 s |

Lince terminó 2.º de 30 en la clasificatoria.

Clasificados a la final:

| Puesto | Jugador | Fichas o caída |
| ---: | --- | --- |
| 1 | darwin-3 | 151678 |
| 2 | **Lince** | 135387 |
| 3 | novato-4 | 12935 |
| 4 | novato-3 | eliminado en la ronda 8 |
| 5 | novato-1 | eliminado en la ronda 7 |
| 6 | espia-1 | eliminado en la ronda 6 |

**Final**: ganó **Lince** en 1650 manos (12 s). Orden: **Lince**, novato-4, novato-3, novato-1, espia-1, darwin-3.

Timeouts de Lince en toda la liga: 0. Duración de la liga: 62 s.

**Chequeo del estado de Lince** (estado.json y log):

- Cree estar en la ronda 11, fase final (jugador lince). De verdad jugó 11 mesas y la última fue de la final.
- Fases por ronda según Lince: clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, final.
- Coincide: rondas, fases, fichas al empezar cada ronda y jugadores por mesa.
- Log: 0 líneas con PÁNICO y 0 con ERROR.

## Liga 9 (semilla 508)

| Ronda | Mesas | Con fichas antes | Con fichas después | Fichas de Lince | Timeouts Lince | Suma de fichas | Duración |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| 1 | 6 | 30 | 19 | **0 (cae aquí)** | 0 | ok | 5 s |
| 2 | 4 | 19 | 9 | — | 0 | ok | 3 s |
| 3 | 2 | 9 | 6 | — | 0 | ok | 4 s |
| 4 | 1 | 6 | 3 | — | 0 | ok | 4 s |
| 5 | 1 | 3 | 2 | — | 0 | ok | 1 s |
| 6 | 1 | 2 | 2 | — | 0 | ok | 1 s |
| 7 | 1 | 2 | 1 | — | 0 | ok | 1 s |
| 8 | 0 | 1 | 1 | — | 0 | ok | 0 s |
| 9 | 0 | 1 | 1 | — | 0 | ok | 0 s |
| 10 | 0 | 1 | 1 | — | 0 | ok | 0 s |

Lince cayó en la ronda 1. Lince terminó 22.º de 30 en la clasificatoria.

Clasificados a la final:

| Puesto | Jugador | Fichas o caída |
| ---: | --- | --- |
| 1 | darwin-3 | 300000 |
| 2 | novato-1 | eliminado en la ronda 7 |
| 3 | darwin-1 | eliminado en la ronda 5 |
| 4 | estadistico-1 | eliminado en la ronda 4 |
| 5 | espia-1 | eliminado en la ronda 4 |
| 6 | base-1 | eliminado en la ronda 4 |

**Final**: ganó darwin-1 en 55 manos (3 s). Orden: darwin-1, novato-1, base-1, estadistico-1, darwin-3, espia-1.

Timeouts de Lince en toda la liga: 0. Duración de la liga: 21 s.

**Chequeo del estado de Lince** (estado.json y log):

- Cree estar en la ronda 1, fase clasificatoria (jugador lince). De verdad jugó 1 mesas y la última fue de la clasificatoria.
- Fases por ronda según Lince: clasificatoria.
- Coincide: rondas, fases, fichas al empezar cada ronda y jugadores por mesa.
- Log: 0 líneas con PÁNICO y 0 con ERROR.

## Liga 10 (semilla 509)

| Ronda | Mesas | Con fichas antes | Con fichas después | Fichas de Lince | Timeouts Lince | Suma de fichas | Duración |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| 1 | 6 | 30 | 19 | 5806 | 0 | ok | 4 s |
| 2 | 4 | 19 | 14 | 20041 | 0 | ok | 3 s |
| 3 | 3 | 14 | 8 | 29400 | 0 | ok | 5 s |
| 4 | 2 | 8 | 5 | 32764 | 0 | ok | 7 s |
| 5 | 1 | 5 | 4 | 31402 | 0 | ok | 7 s |
| 6 | 1 | 4 | 3 | 28233 | 0 | ok | 6 s |
| 7 | 1 | 3 | 3 | 29982 | 0 | ok | 6 s |
| 8 | 1 | 3 | 3 | 34664 | 0 | ok | 6 s |
| 9 | 1 | 3 | 3 | 73506 | 0 | ok | 6 s |
| 10 | 1 | 3 | 3 | 71704 | 0 | ok | 6 s |

Lince terminó 2.º de 30 en la clasificatoria.

Clasificados a la final:

| Puesto | Jugador | Fichas o caída |
| ---: | --- | --- |
| 1 | espia-4 | 201100 |
| 2 | **Lince** | 71704 |
| 3 | novato-1 | 27196 |
| 4 | darwin-1 | eliminado en la ronda 6 |
| 5 | novato-3 | eliminado en la ronda 5 |
| 6 | espia-1 | eliminado en la ronda 4 |

**Final**: ganó **Lince** en 1166 manos (15 s). Orden: **Lince**, novato-3, novato-1, espia-4, darwin-1, espia-1.

Timeouts de Lince en toda la liga: 0. Duración de la liga: 70 s.

**Chequeo del estado de Lince** (estado.json y log):

- Cree estar en la ronda 11, fase final (jugador lince). De verdad jugó 11 mesas y la última fue de la final.
- Fases por ronda según Lince: clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, final.
- Coincide: rondas, fases, fichas al empezar cada ronda y jugadores por mesa.
- Log: 0 líneas con PÁNICO y 0 con ERROR.

## Liga 11 (semilla 510)

| Ronda | Mesas | Con fichas antes | Con fichas después | Fichas de Lince | Timeouts Lince | Suma de fichas | Duración |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| 1 | 6 | 30 | 17 | 20818 | 0 | ok | 4 s |
| 2 | 3 | 17 | 12 | 16677 | 0 | ok | 6 s |
| 3 | 2 | 12 | 8 | 10909 | 0 | ok | 7 s |
| 4 | 2 | 8 | 8 | 18998 | 0 | ok | 6 s |
| 5 | 2 | 8 | 7 | 49126 | 0 | ok | 6 s |
| 6 | 2 | 7 | 6 | 42783 | 0 | ok | 7 s |
| 7 | 1 | 6 | 6 | 37493 | 0 | ok | 9 s |
| 8 | 1 | 6 | 4 | 47396 | 0 | ok | 8 s |
| 9 | 1 | 4 | 4 | 50662 | 0 | ok | 7 s |
| 10 | 1 | 4 | 4 | 63065 | 0 | ok | 8 s |

Lince terminó 3.º de 30 en la clasificatoria.

Clasificados a la final:

| Puesto | Jugador | Fichas o caída |
| ---: | --- | --- |
| 1 | estadistico-4 | 117835 |
| 2 | novato-3 | 81570 |
| 3 | **Lince** | 63065 |
| 4 | espia-2 | 37530 |
| 5 | novato-1 | eliminado en la ronda 8 |
| 6 | estadistico-1 | eliminado en la ronda 8 |

**Final**: ganó **Lince** en 812 manos (21 s). Orden: **Lince**, novato-1, novato-3, estadistico-4, espia-2, estadistico-1.

Timeouts de Lince en toda la liga: 0. Duración de la liga: 90 s.

**Chequeo del estado de Lince** (estado.json y log):

- Cree estar en la ronda 11, fase final (jugador lince). De verdad jugó 11 mesas y la última fue de la final.
- Fases por ronda según Lince: clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, final.
- Coincide: rondas, fases, fichas al empezar cada ronda y jugadores por mesa.
- Log: 0 líneas con PÁNICO y 0 con ERROR.

## Liga 12 (semilla 511)

| Ronda | Mesas | Con fichas antes | Con fichas después | Fichas de Lince | Timeouts Lince | Suma de fichas | Duración |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| 1 | 6 | 30 | 16 | 17749 | 0 | ok | 5 s |
| 2 | 3 | 16 | 8 | 20229 | 0 | ok | 3 s |
| 3 | 2 | 8 | 7 | 34565 | 0 | ok | 5 s |
| 4 | 2 | 7 | 5 | 40058 | 0 | ok | 6 s |
| 5 | 1 | 5 | 5 | 55808 | 0 | ok | 7 s |
| 6 | 1 | 5 | 5 | 54221 | 0 | ok | 7 s |
| 7 | 1 | 5 | 5 | 52084 | 0 | ok | 7 s |
| 8 | 1 | 5 | 5 | 54045 | 0 | ok | 7 s |
| 9 | 1 | 5 | 5 | 55681 | 0 | ok | 7 s |
| 10 | 1 | 5 | 4 | 61682 | 0 | ok | 7 s |

Lince terminó 3.º de 30 en la clasificatoria.

Clasificados a la final:

| Puesto | Jugador | Fichas o caída |
| ---: | --- | --- |
| 1 | darwin-3 | 153310 |
| 2 | novato-3 | 66212 |
| 3 | **Lince** | 61682 |
| 4 | novato-2 | 18796 |
| 5 | darwin-4 | eliminado en la ronda 10 |
| 6 | espia-4 | eliminado en la ronda 4 |

**Final**: ganó **Lince** en 488 manos (11 s). Orden: **Lince**, novato-3, darwin-4, darwin-3, novato-2, espia-4.

Timeouts de Lince en toda la liga: 0. Duración de la liga: 72 s.

**Chequeo del estado de Lince** (estado.json y log):

- Cree estar en la ronda 11, fase final (jugador lince). De verdad jugó 11 mesas y la última fue de la final.
- Fases por ronda según Lince: clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, final.
- Coincide: rondas, fases, fichas al empezar cada ronda y jugadores por mesa.
- Log: 0 líneas con PÁNICO y 0 con ERROR.

## Liga 13 (semilla 512)

| Ronda | Mesas | Con fichas antes | Con fichas después | Fichas de Lince | Timeouts Lince | Suma de fichas | Duración |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| 1 | 6 | 30 | 21 | 6409 | 0 | ok | 5 s |
| 2 | 4 | 21 | 12 | 19505 | 0 | ok | 3 s |
| 3 | 2 | 12 | 10 | 22476 | 0 | ok | 5 s |
| 4 | 2 | 10 | 8 | 20681 | 0 | ok | 6 s |
| 5 | 2 | 8 | 5 | 18638 | 0 | ok | 7 s |
| 6 | 1 | 5 | 5 | 23326 | 0 | ok | 6 s |
| 7 | 1 | 5 | 5 | 42818 | 0 | ok | 6 s |
| 8 | 1 | 5 | 5 | 41847 | 0 | ok | 6 s |
| 9 | 1 | 5 | 5 | 72468 | 0 | ok | 6 s |
| 10 | 1 | 5 | 4 | 74739 | 0 | ok | 6 s |

Lince terminó 2.º de 30 en la clasificatoria.

Clasificados a la final:

| Puesto | Jugador | Fichas o caída |
| ---: | --- | --- |
| 1 | novato-4 | 155106 |
| 2 | **Lince** | 74739 |
| 3 | novato-2 | 40803 |
| 4 | novato-1 | 29352 |
| 5 | darwin-4 | eliminado en la ronda 10 |
| 6 | darwin-1 | eliminado en la ronda 5 |

**Final**: ganó **Lince** en 746 manos (8 s). Orden: **Lince**, novato-4, novato-1, novato-2, darwin-1, darwin-4.

Timeouts de Lince en toda la liga: 0. Duración de la liga: 65 s.

**Chequeo del estado de Lince** (estado.json y log):

- Cree estar en la ronda 11, fase final (jugador lince). De verdad jugó 11 mesas y la última fue de la final.
- Fases por ronda según Lince: clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, final.
- Coincide: rondas, fases, fichas al empezar cada ronda y jugadores por mesa.
- Log: 0 líneas con PÁNICO y 0 con ERROR.

## Liga 14 (semilla 513)

| Ronda | Mesas | Con fichas antes | Con fichas después | Fichas de Lince | Timeouts Lince | Suma de fichas | Duración |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| 1 | 6 | 30 | 18 | 14369 | 0 | ok | 4 s |
| 2 | 3 | 18 | 10 | 27789 | 0 | ok | 4 s |
| 3 | 2 | 10 | 7 | 22121 | 0 | ok | 7 s |
| 4 | 2 | 7 | 6 | 43065 | 0 | ok | 6 s |
| 5 | 1 | 6 | 5 | 45219 | 0 | ok | 6 s |
| 6 | 1 | 5 | 4 | 79586 | 0 | ok | 7 s |
| 7 | 1 | 4 | 3 | 76929 | 0 | ok | 7 s |
| 8 | 1 | 3 | 3 | 76967 | 0 | ok | 7 s |
| 9 | 1 | 3 | 3 | 161447 | 0 | ok | 7 s |
| 10 | 1 | 3 | 3 | 162545 | 0 | ok | 7 s |

Lince terminó 1.º de 30 en la clasificatoria.

Clasificados a la final:

| Puesto | Jugador | Fichas o caída |
| ---: | --- | --- |
| 1 | **Lince** | 162545 |
| 2 | darwin-4 | 109782 |
| 3 | novato-3 | 27673 |
| 4 | espia-2 | eliminado en la ronda 7 |
| 5 | novato-2 | eliminado en la ronda 6 |
| 6 | novato-1 | eliminado en la ronda 5 |

**Final**: ganó **Lince** en 896 manos (9 s). Orden: **Lince**, novato-1, novato-2, espia-2, novato-3, darwin-4.

Timeouts de Lince en toda la liga: 0. Duración de la liga: 70 s.

**Chequeo del estado de Lince** (estado.json y log):

- Cree estar en la ronda 11, fase final (jugador lince). De verdad jugó 11 mesas y la última fue de la final.
- Fases por ronda según Lince: clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, final.
- Coincide: rondas, fases, fichas al empezar cada ronda y jugadores por mesa.
- Log: 0 líneas con PÁNICO y 0 con ERROR.

## Liga 15 (semilla 514)

| Ronda | Mesas | Con fichas antes | Con fichas después | Fichas de Lince | Timeouts Lince | Suma de fichas | Duración |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| 1 | 6 | 30 | 17 | 12409 | 0 | ok | 4 s |
| 2 | 3 | 17 | 11 | 42207 | 0 | ok | 5 s |
| 3 | 2 | 11 | 8 | 43421 | 0 | ok | 4 s |
| 4 | 2 | 8 | 6 | 9065 | 0 | ok | 9 s |
| 5 | 1 | 6 | 4 | 6722 | 0 | ok | 10 s |
| 6 | 1 | 4 | 3 | 59114 | 0 | ok | 7 s |
| 7 | 1 | 3 | 2 | **0 (cae aquí)** | 0 | ok | 7 s |
| 8 | 1 | 2 | 2 | — | 0 | ok | 1 s |
| 9 | 1 | 2 | 1 | — | 0 | ok | 0 s |
| 10 | 0 | 1 | 1 | — | 0 | ok | 0 s |

Lince cayó en la ronda 7. Lince terminó 3.º de 30 en la clasificatoria.

Clasificados a la final:

| Puesto | Jugador | Fichas o caída |
| ---: | --- | --- |
| 1 | darwin-2 | 300000 |
| 2 | novato-1 | eliminado en la ronda 9 |
| 3 | **Lince** | eliminado en la ronda 7 |
| 4 | estadistico-4 | eliminado en la ronda 6 |
| 5 | estadistico-2 | eliminado en la ronda 5 |
| 6 | novato-4 | eliminado en la ronda 5 |

**Final**: ganó novato-1 en 540 manos (14 s). Orden: novato-1, darwin-2, **Lince**, novato-4, estadistico-4, estadistico-2.

Timeouts de Lince en toda la liga: 0. Duración de la liga: 61 s.

**Chequeo del estado de Lince** (estado.json y log):

- Cree estar en la ronda 8, fase final (jugador lince). De verdad jugó 8 mesas y la última fue de la final.
- Fases por ronda según Lince: clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, clasificatoria, final.
- Coincide: rondas, fases, fichas al empezar cada ronda y jugadores por mesa.
- Log: 0 líneas con PÁNICO y 0 con ERROR.

**Problemas del ensayo:**

- Lince clasificó habiendo quebrado: se relanzó su proceso para la final

## Resumen de 15 ligas

| Métrica | Ligas | % |
| --- | ---: | ---: |
| Lince llegó a la final | 12 de 15 | 80 % |
| Lince 1.º de la clasificatoria | 3 de 15 | 20 % |
| Lince ganó la final | 10 de 15 | 67 % |
| Estado de Lince coincide con la realidad | 15 de 15 | 100 % |

Quiebras de Lince por ronda: R1: 1; R3: 2; R7: 1;

Timeouts de Lince: 0 en total. Duración media de una liga: 62 s.

Tiempo total: 15m36s.
