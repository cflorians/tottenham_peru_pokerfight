# A1 + B1 · Línea base en el simulador de liga y de mesa final

- **Fecha**: vie 13/11/2026 · **Quién**: Carlos (con Claude Code) · **Commit base**: `49ce119` + rama `comun/simulador-liga`
- **Herramienta**: `bots/lince/cmd/simular-liga` (simulador rápido: `entrenamiento/liga.go` y `campos.go`)
- **Comandos** (desde `pokerfight-desarrollo/`):
  - `go run ./bots/lince/cmd/simular-liga -n 15,20,30,40` (secciones 3 en adelante)
  - `go run ./bots/lince/cmd/simular-liga -n 20 -heroes v0 -tabla` (tablas de valor, al final)
- **Tamaño**: 2 000 ligas completas por bot, campo y tamaño (96 000 ligas en total), más 3 finales por cada liga en que el bot clasifica. Semilla 2026, biblioteca de 20 000 repartos con 200 muestras.
- **Formato**: liga de 10 rondas de 50 manos, 10 000 fichas, ciegas 50/100, mesas de 5-6 sorteadas en cada ronda, fichas arrastradas, pasan 6. Final de 6 con 10 000 fichas y tope de 150, 500 o 5 000 manos ("sin reloj").

Métricas: **M1** = llega a la final, **M1b** = termina 1.º en la clasificatoria, **M2** = gana la final estando en ella (el azar da 16,7 %), **U** = 60 × M1 × M2 + 30 × M1b (premio esperado en soles).

## 1. Conclusiones

1. **La v0 no alcanza, y el simulador lo muestra con claridad: Novato le gana en todo.** En el campo mixto con 20 inscritos, la v0 llega a la final el 57,8 % de las veces y Novato el 66,2 %; en premio esperado, 7,2 contra 15,2. Novato es mejor en los cuatro campos y en los cuatro tamaños, salvo en el campo agresivo, donde empatan. La v1 tiene que superar a Novato, no solo a la v0.
2. **La v0 juega la final peor que el azar.** Su M2 está entre 4 % y 18 % en los campos kit, tight y mixto, casi siempre por debajo del azar (16,7 %). El genoma prudente no sirve para la final: confirma que hace falta un genoma propio (v2). Solo en el campo agresivo llega a 23-35 %.
3. **El genoma prudente arregla la primera ronda, no la liga.** En la ronda 1 la v0 quiebra ~10 % (Darwin original, 40-48 %). Pero en las rondas 2 a 6, con las fichas arrastradas, la v0 quiebra entre 13 % y 18 % por ronda en el campo mixto (hasta 27 % en el agresivo), y solo termina la liga con fichas una de cada cuatro veces. La meta del plan es ≤3 % por ronda.
4. **Casi nunca llegan 6 con fichas a la ronda 10.** Con 20 inscritos, en los campos kit, agresivo y mixto, a la ronda 4 ya quedan 6 o menos en el 82-100 % de las ligas: clasifican "los últimos eliminados", y lo que importa es **en qué ronda se cae**, no las fichas. Solo en el campo tight deciden las fichas (la mitad de las veces), y ahí el corte está en ~10 400 fichas: casi el stack inicial.
5. **El campo tight es el peor para nosotros**: la v0 llega a la final solo el 27,6 % de las veces con 20 inscritos (Novato 50,7 %). Contra bots que esperan cartas, la v0 pierde fichas de a poco.
6. **La cantidad de inscritos pesa mucho** (pregunta 1 a la organización): en el campo mixto, M1 de la v0 pasa de 68,6 % con 15 inscritos a 29,3 % con 40.
7. **Quién llega a la final**: sobre todo los bots tight (novato y roca, ~65 % de los suyos) y los prudentes (27-57 %). Los maníacos y aleatorios casi nunca (2-13 %). Los rivales de la final serán tight, como suponía el plan B.

## 2. Qué tan creíble es esto

- **Paridad con la arena real, solo en la ronda 1**: en las 90 mesas reales de 5 (`pruebas_formato_real.md`), Darwin quebró el 44 % y el genoma prudente y Novato el 10 %. En el simulador, en la ronda 1 del campo kit: Darwin 47,6 %, v0 9,8 %, Novato 14,1 %. Coincide dentro del margen de aquellas pruebas (10 % era en realidad 5-18 %). **Falta la prueba de paridad completa** con stacks arrastrados (nivel real de la Fase 1).
- **Los rivales son réplicas aproximadas.** No hay réplica de Michi Espía, Bot Estadístico ni Gato Bandido; en el campo kit los reemplaza Darwin. Novato y Base sí están calcados regla por regla.
- **Supuestos donde las bases callan** (preguntas 2 a 4): entre eliminados queda mejor el que cayó más tarde (ronda, mano, y fichas al empezar esa mano); empates en fichas por sorteo; mesas de 4 (o de 3 y 4) cuando no alcanzan las de 5 y 6; la liga se juega entera aunque queden 6 o menos (no cambia quién clasifica).
- **La hora de la final se aproxima con un tope de manos.** No sabemos cuántas manos entran en una hora.
- El campo y las tablas de valor se midieron en las ligas con la v0 como héroe (1 de N participantes).

## 3. Resultados completos

Simulador rápido · 2000 ligas por héroe, campo y tamaño · semilla 2026 · biblioteca de 20000 repartos con 200 muestras

Liga: 10 rondas de 50 manos, 10000 fichas, ciegas 50/100, pasan 6. Final: 6 jugadores, 10000 fichas, topes de manos 150,500,0 (0 = 5000).

Entre corchetes, el intervalo de confianza del 95 % (Wilson). U = 60 × M1 × M2 + 30 × M1b, con M2 promediada entre los topes.

### Campo kit · N = 15

Rivales: 2 aleatorio, 1 conservador, 3 base, 3 novato, 5 darwin.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 74.5 % [72.5-76.4] | 10.4 % [9.2-11.9] | 18.3 % [16.4-20.4] | 16.1 % [14.3-18.1] | 15.2 % [13.4-17.1] | 10.52 | 16861 |
| Darwin original | 40.6 % [38.5-42.8] | 9.0 % [7.9-10.4] | 22.8 % [20.0-25.8] | 16.0 % [13.6-18.7] | 12.8 % [10.7-15.3] | 6.91 | 12670 |
| Novato | 75.6 % [73.7-77.5] | 14.3 % [12.8-15.9] | 20.7 % [18.7-22.8] | 24.9 % [22.8-27.2] | 24.6 % [22.5-26.8] | 14.91 | 22264 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 11.2 % | 19.5 % | 21.3 % | 19.1 % | 13.9 % | 12.0 % | 11.8 % | 6.7 % | 6.5 % | 6.6 % | 24.8 % |
| Darwin original | 49.0 % | 33.1 % | 27.3 % | 18.3 % | 15.6 % | 11.4 % | 10.2 % | 5.9 % | 5.5 % | 6.6 % | 11.3 % |
| Novato | 14.9 % | 13.0 % | 13.5 % | 11.2 % | 10.2 % | 9.5 % | 8.1 % | 5.3 % | 5.6 % | 4.1 % | 36.4 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 8.6 | 5.5 | 4.1 | 3.3 | 2.8 | 2.5 | 2.3 | 2.1 | 2.0 | 1.9 |
| Ligas en que ya quedan 6 o menos (acumulado) | 4 % | 81 % | 98 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % |

- Fichas del 1.º al terminar: **118390** de media.
- Nunca llegan 6 con fichas al final: clasifican los últimos eliminados.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.74, aleatorio 0.15 (8 % de los suyos), conservador 0.24 (24 % de los suyos), base 0.78 (26 % de los suyos), novato 2.24 (75 % de los suyos), darwin 1.84 (37 % de los suyos).

### Campo kit · N = 20

Rivales: 2 aleatorio, 2 conservador, 4 base, 4 novato, 7 darwin.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 63.3 % [61.2-65.4] | 8.6 % [7.4-9.9] | 16.9 % [14.9-19.1] | 15.2 % [13.3-17.2] | 13.3 % [11.6-15.3] | 8.33 | 18593 |
| Darwin original | 29.4 % [27.5-31.5] | 6.1 % [5.1-7.2] | 18.0 % [15.1-21.3] | 11.5 % [9.2-14.4] | 11.9 % [9.5-14.7] | 4.27 | 11460 |
| Novato | 67.8 % [65.8-69.9] | 11.5 % [10.2-13.0] | 20.3 % [18.2-22.5] | 21.4 % [19.3-23.6] | 23.7 % [21.5-26.0] | 12.31 | 24208 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 9.8 % | 17.7 % | 18.1 % | 17.7 % | 15.8 % | 13.1 % | 10.5 % | 9.8 % | 9.6 % | 7.3 % | 24.8 % |
| Darwin original | 47.6 % | 36.2 % | 30.5 % | 25.8 % | 20.3 % | 13.5 % | 12.6 % | 9.1 % | 4.8 % | 4.4 % | 8.6 % |
| Novato | 14.1 % | 13.1 % | 13.1 % | 13.0 % | 10.8 % | 8.8 % | 7.9 % | 6.4 % | 7.8 % | 4.9 % | 34.6 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 12.0 | 7.4 | 5.5 | 4.3 | 3.6 | 3.1 | 2.8 | 2.6 | 2.4 | 2.2 |
| Ligas en que ya quedan 6 o menos (acumulado) | 0 % | 27 % | 78 % | 94 % | 99 % | 100 % | 100 % | 100 % | 100 % | 100 % |

- Fichas del 1.º al terminar: **145298** de media.
- Nunca llegan 6 con fichas al final: clasifican los últimos eliminados.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.63, aleatorio 0.06 (3 % de los suyos), conservador 0.09 (5 % de los suyos), base 0.63 (16 % de los suyos), novato 2.62 (65 % de los suyos), darwin 1.97 (28 % de los suyos).

### Campo kit · N = 30

Rivales: 3 aleatorio, 3 conservador, 7 base, 6 novato, 10 darwin.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 44.5 % [42.3-46.6] | 5.6 % [4.7-6.7] | 13.7 % [11.6-16.1] | 10.7 % [8.8-12.9] | 12.4 % [10.4-14.7] | 4.95 | 18165 |
| Darwin original | 18.6 % [16.9-20.3] | 4.5 % [3.7-5.5] | 14.0 % [10.9-17.9] | 8.4 % [5.9-11.6] | 10.8 % [8.0-14.3] | 2.58 | 11463 |
| Novato | 52.6 % [50.4-54.8] | 7.7 % [6.6-9.0] | 20.6 % [18.3-23.2] | 19.7 % [17.4-22.2] | 22.4 % [20.0-25.1] | 8.91 | 26544 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 9.8 % | 18.3 % | 20.8 % | 18.2 % | 16.8 % | 11.9 % | 12.6 % | 8.3 % | 11.1 % | 7.2 % | 23.1 % |
| Darwin original | 47.0 % | 34.2 % | 33.5 % | 28.2 % | 23.1 % | 16.8 % | 14.1 % | 15.3 % | 7.7 % | 9.1 % | 6.5 % |
| Novato | 13.6 % | 15.4 % | 15.4 % | 11.3 % | 9.5 % | 9.8 % | 7.6 % | 7.4 % | 7.4 % | 6.5 % | 33.2 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 17.8 | 10.7 | 7.6 | 6.1 | 5.1 | 4.4 | 3.9 | 3.5 | 3.2 | 3.0 |
| Ligas en que ya quedan 6 o menos (acumulado) | 0 % | 0 % | 24 % | 58 % | 80 % | 91 % | 96 % | 98 % | 99 % | 100 % |

- Fichas del 1.º al terminar: **189892** de media.
- Fichas del 6.º (el corte), cuando llegan 6 o más con fichas (3 % de las ligas): **12236** de media.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.44, aleatorio 0.03 (1 % de los suyos), conservador 0.01 (0 % de los suyos), base 0.49 (7 % de los suyos), novato 3.20 (53 % de los suyos), darwin 1.84 (18 % de los suyos).

### Campo kit · N = 40

Rivales: 4 aleatorio, 4 conservador, 9 base, 9 novato, 13 darwin.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 31.7 % [29.7-33.8] | 3.9 % [3.1-4.8] | 14.8 % [12.3-17.8] | 10.3 % [8.1-12.9] | 9.8 % [7.7-12.3] | 3.38 | 16516 |
| Darwin original | 10.7 % [9.4-12.1] | 3.6 % [2.9-4.5] | 14.1 % [10.0-19.4] | 6.6 % [4.0-10.7] | 9.9 % [6.5-14.6] | 1.73 | 10341 |
| Novato | 43.0 % [40.9-45.2] | 4.8 % [3.9-5.8] | 20.2 % [17.7-23.0] | 22.3 % [19.6-25.2] | 19.5 % [17.0-22.3] | 6.76 | 26669 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 10.5 % | 18.8 % | 18.8 % | 17.2 % | 16.4 % | 13.8 % | 11.9 % | 12.3 % | 11.9 % | 8.8 % | 21.9 % |
| Darwin original | 47.1 % | 35.2 % | 34.1 % | 30.5 % | 20.4 % | 20.4 % | 14.6 % | 10.6 % | 8.6 % | 10.1 % | 6.2 % |
| Novato | 13.8 % | 13.8 % | 13.9 % | 10.9 % | 10.4 % | 8.3 % | 8.8 % | 7.5 % | 5.2 % | 6.9 % | 34.9 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 24.1 | 14.7 | 10.7 | 8.5 | 7.2 | 6.2 | 5.4 | 4.9 | 4.4 | 4.1 |
| Ligas en que ya quedan 6 o menos (acumulado) | 0 % | 0 % | 2 % | 13 % | 34 % | 55 % | 71 % | 82 % | 88 % | 93 % |

- Fichas del 1.º al terminar: **215756** de media.
- Fichas del 6.º (el corte), cuando llegan 6 o más con fichas (19 % de las ligas): **18819** de media.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.32, aleatorio 0.01 (0 % de los suyos), conservador 0.00 (0 % de los suyos), base 0.33 (4 % de los suyos), novato 3.84 (43 % de los suyos), darwin 1.50 (12 % de los suyos).

### Campo agresivo · N = 15

Rivales: 4 maniaco, 3 aleatorio, 4 darwin, 2 estacion, 1 base.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 95.5 % [94.5-96.3] | 15.3 % [13.8-17.0] | 23.9 % [22.1-25.9] | 34.5 % [32.4-36.6] | 35.1 % [33.0-37.2] | 22.46 | 24953 |
| Darwin original | 69.0 % [66.9-70.9] | 18.4 % [16.8-20.2] | 25.9 % [23.6-28.3] | 26.1 % [23.9-28.5] | 25.5 % [23.3-27.9] | 16.23 | 27613 |
| Novato | 91.4 % [90.1-92.6] | 19.7 % [18.0-21.4] | 22.9 % [21.1-24.9] | 30.6 % [28.5-32.7] | 33.7 % [31.6-35.9] | 21.84 | 30571 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 17.9 % | 26.1 % | 20.2 % | 13.5 % | 10.0 % | 8.6 % | 8.4 % | 5.7 % | 5.5 % | 3.7 % | 27.1 % |
| Darwin original | 48.2 % | 40.3 % | 18.8 % | 11.6 % | 6.3 % | 6.5 % | 2.6 % | 0.3 % | 0.8 % | 1.1 % | 18.6 % |
| Novato | 19.1 % | 21.0 % | 14.5 % | 11.2 % | 9.0 % | 8.4 % | 7.4 % | 6.3 % | 4.8 % | 4.3 % | 32.0 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 4.6 | 2.3 | 1.8 | 1.6 | 1.4 | 1.3 | 1.3 | 1.2 | 1.2 | 1.2 |
| Ligas en que ya quedan 6 o menos (acumulado) | 98 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % |

- Fichas del 1.º al terminar: **144181** de media.
- Nunca llegan 6 con fichas al final: clasifican los últimos eliminados.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.95, maniaco 0.63 (16 % de los suyos), aleatorio 0.57 (19 % de los suyos), darwin 2.61 (65 % de los suyos), estacion 0.59 (29 % de los suyos), base 0.65 (65 % de los suyos).

### Campo agresivo · N = 20

Rivales: 6 maniaco, 4 aleatorio, 5 darwin, 2 estacion, 2 base.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 83.3 % [81.6-84.9] | 11.5 % [10.1-12.9] | 25.9 % [23.8-28.0] | 34.2 % [32.0-36.5] | 35.2 % [33.0-37.6] | 19.31 | 25734 |
| Darwin original | 56.4 % [54.2-58.5] | 14.3 % [12.9-16.0] | 26.4 % [23.9-29.0] | 26.7 % [24.2-29.4] | 24.8 % [22.3-27.4] | 13.08 | 28523 |
| Novato | 84.0 % [82.3-85.5] | 15.4 % [13.9-17.0] | 23.3 % [21.4-25.4] | 31.4 % [29.2-33.6] | 30.5 % [28.3-32.7] | 18.93 | 34902 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 21.3 % | 27.3 % | 22.6 % | 14.0 % | 10.5 % | 7.0 % | 7.3 % | 6.8 % | 4.2 % | 5.7 % | 24.8 % |
| Darwin original | 47.9 % | 39.8 % | 28.9 % | 15.5 % | 9.5 % | 5.6 % | 3.7 % | 2.3 % | 2.6 % | 1.7 % | 14.5 % |
| Novato | 18.6 % | 20.3 % | 18.4 % | 11.9 % | 9.4 % | 6.3 % | 6.7 % | 6.2 % | 4.8 % | 4.4 % | 31.5 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 6.0 | 3.0 | 2.1 | 1.7 | 1.5 | 1.4 | 1.3 | 1.3 | 1.2 | 1.2 |
| Ligas en que ya quedan 6 o menos (acumulado) | 71 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % |

- Fichas del 1.º al terminar: **191122** de media.
- Nunca llegan 6 con fichas al final: clasifican los últimos eliminados.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.83, maniaco 0.59 (10 % de los suyos), aleatorio 0.53 (13 % de los suyos), darwin 2.71 (54 % de los suyos), estacion 0.39 (19 % de los suyos), base 0.94 (47 % de los suyos).

### Campo agresivo · N = 30

Rivales: 9 maniaco, 6 aleatorio, 8 darwin, 3 estacion, 3 base.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 69.7 % [67.6-71.7] | 6.5 % [5.5-7.7] | 23.2 % [21.1-25.5] | 31.4 % [29.0-33.9] | 34.6 % [32.1-37.1] | 14.39 | 25877 |
| Darwin original | 41.6 % [39.5-43.8] | 10.2 % [8.9-11.6] | 23.0 % [20.2-25.9] | 22.7 % [20.0-25.7] | 21.9 % [19.2-24.8] | 8.66 | 30481 |
| Novato | 74.0 % [72.0-75.9] | 10.2 % [8.9-11.6] | 19.4 % [17.5-21.5] | 29.1 % [26.9-31.5] | 29.7 % [27.5-32.1] | 14.64 | 36846 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 22.9 % | 28.9 % | 21.7 % | 14.5 % | 10.2 % | 7.9 % | 6.8 % | 6.0 % | 5.5 % | 5.8 % | 23.6 % |
| Darwin original | 49.2 % | 40.4 % | 33.7 % | 18.7 % | 14.4 % | 7.1 % | 8.8 % | 5.5 % | 3.1 % | 1.4 % | 10.7 % |
| Novato | 19.8 % | 19.9 % | 19.3 % | 14.8 % | 11.3 % | 9.2 % | 6.5 % | 6.5 % | 5.5 % | 3.4 % | 28.4 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 8.7 | 4.1 | 2.6 | 2.1 | 1.8 | 1.6 | 1.5 | 1.4 | 1.3 | 1.3 |
| Ligas en que ya quedan 6 o menos (acumulado) | 2 % | 97 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % |

- Fichas del 1.º al terminar: **283271** de media.
- Nunca llegan 6 con fichas al final: clasifican los últimos eliminados.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.70, maniaco 0.33 (4 % de los suyos), aleatorio 0.37 (6 % de los suyos), darwin 3.42 (43 % de los suyos), estacion 0.22 (7 % de los suyos), base 0.97 (32 % de los suyos).

### Campo agresivo · N = 40

Rivales: 12 maniaco, 8 aleatorio, 11 darwin, 4 estacion, 4 base.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 63.6 % [61.5-65.7] | 5.2 % [4.3-6.3] | 25.9 % [23.6-28.4] | 34.1 % [31.5-36.7] | 35.4 % [32.8-38.1] | 13.71 | 29179 |
| Darwin original | 35.2 % [33.1-37.3] | 7.8 % [6.7-9.1] | 21.4 % [18.6-24.6] | 22.2 % [19.2-25.4] | 22.4 % [19.5-25.7] | 6.99 | 31104 |
| Novato | 67.7 % [65.6-69.7] | 7.3 % [6.2-8.5] | 20.7 % [18.6-22.9] | 27.6 % [25.3-30.1] | 26.2 % [23.9-28.6] | 12.28 | 39654 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 21.6 % | 27.2 % | 22.9 % | 13.3 % | 10.4 % | 9.9 % | 6.0 % | 5.4 % | 5.7 % | 5.4 % | 24.4 % |
| Darwin original | 49.6 % | 41.0 % | 34.1 % | 24.7 % | 17.3 % | 12.3 % | 9.3 % | 3.1 % | 5.9 % | 4.0 % | 8.5 % |
| Novato | 20.6 % | 21.5 % | 18.0 % | 16.8 % | 10.8 % | 10.5 % | 6.9 % | 6.0 % | 6.1 % | 5.0 % | 26.5 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 11.5 | 5.1 | 3.2 | 2.4 | 2.0 | 1.7 | 1.6 | 1.5 | 1.4 | 1.3 |
| Ligas en que ya quedan 6 o menos (acumulado) | 0 % | 84 % | 99 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % |

- Fichas del 1.º al terminar: **371181** de media.
- Nunca llegan 6 con fichas al final: clasifican los últimos eliminados.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.64, maniaco 0.18 (1 % de los suyos), aleatorio 0.24 (3 % de los suyos), darwin 3.86 (35 % de los suyos), estacion 0.15 (4 % de los suyos), base 0.94 (23 % de los suyos).

### Campo tight · N = 15

Rivales: 4 roca, 3 novato, 3 base, 1 conservador, 3 prudente.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 38.6 % [36.5-40.8] | 5.2 % [4.4-6.3] | 11.7 % [9.6-14.1] | 7.8 % [6.1-9.9] | 9.6 % [7.7-11.9] | 3.81 | 7615 |
| Darwin original | 18.9 % [17.3-20.7] | 4.0 % [3.3-5.0] | 11.9 % [9.0-15.5] | 6.6 % [4.5-9.6] | 7.9 % [5.6-11.1] | 2.21 | 5124 |
| Novato | 61.3 % [59.1-63.4] | 12.7 % [11.3-14.2] | 19.5 % [17.4-21.8] | 27.9 % [25.5-30.5] | 31.4 % [28.9-34.1] | 13.46 | 17998 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 9.6 % | 15.2 % | 15.6 % | 15.1 % | 13.6 % | 13.3 % | 12.0 % | 11.9 % | 11.4 % | 8.0 % | 26.0 % |
| Darwin original | 37.7 % | 33.9 % | 27.7 % | 20.1 % | 22.7 % | 18.5 % | 16.7 % | 9.6 % | 15.5 % | 11.5 % | 8.4 % |
| Novato | 8.6 % | 8.3 % | 8.6 % | 7.5 % | 5.3 % | 6.6 % | 5.7 % | 4.6 % | 5.3 % | 3.9 % | 51.3 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 12.6 | 10.1 | 8.6 | 7.5 | 6.8 | 6.2 | 5.6 | 5.1 | 4.7 | 4.4 |
| Ligas en que ya quedan 6 o menos (acumulado) | 0 % | 1 % | 7 % | 21 % | 39 % | 56 % | 71 % | 81 % | 89 % | 93 % |

- Fichas del 1.º al terminar: **71108** de media.
- Fichas del 6.º (el corte), cuando llegan 6 o más con fichas (20 % de las ligas): **7019** de media.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.39, roca 2.37 (59 % de los suyos), novato 1.83 (61 % de los suyos), base 0.24 (8 % de los suyos), conservador 0.00 (0 % de los suyos), prudente 1.16 (39 % de los suyos).

### Campo tight · N = 20

Rivales: 5 roca, 4 novato, 4 base, 2 conservador, 4 prudente.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 27.6 % [25.6-29.5] | 4.4 % [3.6-5.4] | 11.6 % [9.2-14.6] | 7.8 % [5.8-10.3] | 6.7 % [4.9-9.1] | 2.76 | 8636 |
| Darwin original | 14.1 % [12.6-15.6] | 3.7 % [3.0-4.6] | 15.3 % [11.6-20.0] | 8.9 % [6.1-12.8] | 6.4 % [4.1-9.9] | 1.97 | 5607 |
| Novato | 50.7 % [48.5-52.9] | 10.9 % [9.6-12.3] | 21.8 % [19.4-24.4] | 26.1 % [23.5-28.9] | 27.6 % [24.9-30.4] | 10.93 | 18394 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 9.7 % | 15.3 % | 14.8 % | 14.0 % | 14.5 % | 12.3 % | 13.2 % | 12.7 % | 10.7 % | 10.7 % | 25.4 % |
| Darwin original | 37.5 % | 30.2 % | 25.8 % | 25.0 % | 17.1 % | 15.1 % | 16.1 % | 10.8 % | 15.6 % | 15.7 % | 9.1 % |
| Novato | 8.0 % | 9.1 % | 7.5 % | 8.5 % | 7.6 % | 6.4 % | 5.9 % | 4.9 % | 4.7 % | 6.0 % | 49.0 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 16.8 | 13.2 | 11.0 | 9.6 | 8.5 | 7.7 | 7.1 | 6.5 | 6.0 | 5.5 |
| Ligas en que ya quedan 6 o menos (acumulado) | 0 % | 0 % | 0 % | 3 % | 10 % | 21 % | 36 % | 49 % | 61 % | 71 % |

- Fichas del 1.º al terminar: **81464** de media.
- Fichas del 6.º (el corte), cuando llegan 6 o más con fichas (50 % de las ligas): **10422** de media.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.28, roca 2.39 (48 % de los suyos), novato 2.09 (52 % de los suyos), base 0.17 (4 % de los suyos), conservador 0.00 (0 % de los suyos), prudente 1.08 (27 % de los suyos).

### Campo tight · N = 30

Rivales: 8 roca, 6 novato, 6 base, 3 conservador, 6 prudente.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 15.6 % [14.1-17.3] | 3.6 % [2.9-4.6] | 10.9 % [7.9-14.8] | 8.0 % [5.5-11.6] | 9.0 % [6.3-12.7] | 1.96 | 8634 |
| Darwin original | 9.3 % [8.2-10.7] | 3.7 % [3.0-4.6] | 10.2 % [6.6-15.3] | 9.6 % [6.2-14.7] | 8.0 % [4.9-12.8] | 1.63 | 6869 |
| Novato | 39.1 % [37.0-41.3] | 7.2 % [6.1-8.4] | 22.3 % [19.5-25.3] | 23.4 % [20.6-26.5] | 24.8 % [21.9-28.0] | 7.67 | 20057 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 8.4 % | 14.2 % | 15.1 % | 15.2 % | 16.2 % | 16.1 % | 13.7 % | 12.1 % | 10.1 % | 11.4 % | 24.0 % |
| Darwin original | 36.1 % | 30.1 % | 25.2 % | 21.9 % | 20.5 % | 18.6 % | 14.5 % | 16.7 % | 12.1 % | 11.8 % | 9.3 % |
| Novato | 8.9 % | 8.2 % | 9.2 % | 6.8 % | 8.1 % | 6.1 % | 6.0 % | 4.4 % | 5.4 % | 6.1 % | 48.9 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 25.2 | 20.1 | 16.7 | 14.6 | 13.0 | 11.7 | 10.5 | 9.6 | 8.8 | 8.2 |
| Ligas en que ya quedan 6 o menos (acumulado) | 0 % | 0 % | 0 % | 0 % | 0 % | 0 % | 1 % | 4 % | 10 % | 17 % |

- Fichas del 1.º al terminar: **91541** de media.
- Fichas del 6.º (el corte), cuando llegan 6 o más con fichas (94 % de las ligas): **19350** de media.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.16, roca 2.47 (31 % de los suyos), novato 2.29 (38 % de los suyos), base 0.13 (2 % de los suyos), conservador 0.00 (0 % de los suyos), prudente 0.95 (16 % de los suyos).

### Campo tight · N = 40

Rivales: 11 roca, 8 novato, 8 base, 4 conservador, 8 prudente.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 11.7 % [10.4-13.2] | 2.7 % [2.1-3.5] | 10.7 % [7.3-15.3] | 4.3 % [2.3-7.7] | 7.3 % [4.6-11.3] | 1.33 | 8412 |
| Darwin original | 6.5 % [5.5-7.7] | 2.5 % [1.9-3.3] | 12.3 % [7.7-19.1] | 6.9 % [3.7-12.6] | 6.2 % [3.2-11.7] | 1.09 | 5388 |
| Novato | 31.2 % [29.2-33.3] | 6.1 % [5.1-7.2] | 18.4 % [15.6-21.7] | 25.0 % [21.8-28.5] | 28.8 % [25.4-32.5] | 6.34 | 20121 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 10.2 % | 15.2 % | 14.0 % | 15.6 % | 14.7 % | 15.5 % | 13.6 % | 10.4 % | 11.0 % | 12.4 % | 24.1 % |
| Darwin original | 38.4 % | 29.5 % | 26.4 % | 23.3 % | 21.2 % | 19.4 % | 17.0 % | 18.9 % | 16.2 % | 13.1 % | 7.7 % |
| Novato | 7.7 % | 8.5 % | 6.6 % | 8.5 % | 7.3 % | 7.8 % | 7.4 % | 5.2 % | 5.0 % | 4.1 % | 49.4 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 33.8 | 26.9 | 22.4 | 19.5 | 17.3 | 15.5 | 14.1 | 12.9 | 11.9 | 10.9 |
| Ligas en que ya quedan 6 o menos (acumulado) | 0 % | 0 % | 0 % | 0 % | 0 % | 0 % | 0 % | 0 % | 1 % | 2 % |

- Fichas del 1.º al terminar: **98721** de media.
- Fichas del 6.º (el corte), cuando llegan 6 o más con fichas (99 % de las ligas): **28728** de media.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.12, roca 2.39 (22 % de los suyos), novato 2.46 (31 % de los suyos), base 0.12 (2 % de los suyos), conservador 0.00 (0 % de los suyos), prudente 0.90 (11 % de los suyos).

### Campo mixto · N = 15

Rivales: 2 aleatorio, 2 conservador, 2 base, 2 novato, 2 roca, 1 maniaco, 1 estacion, 1 darwin, 1 prudente.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 68.6 % [66.5-70.6] | 9.3 % [8.2-10.7] | 17.6 % [15.6-19.7] | 12.2 % [10.6-14.1] | 16.3 % [14.4-18.3] | 9.12 | 14410 |
| Darwin original | 44.8 % [42.6-46.9] | 12.0 % [10.6-13.5] | 18.7 % [16.2-21.3] | 13.3 % [11.2-15.7] | 12.6 % [10.6-15.0] | 7.59 | 16328 |
| Novato | 74.0 % [72.0-75.8] | 17.4 % [15.8-19.2] | 24.1 % [22.0-26.3] | 27.2 % [25.0-29.5] | 29.0 % [26.8-31.4] | 17.11 | 24771 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 10.1 % | 14.6 % | 16.2 % | 14.0 % | 12.5 % | 11.0 % | 11.0 % | 9.4 % | 7.3 % | 7.5 % | 29.8 % |
| Darwin original | 37.5 % | 22.8 % | 20.0 % | 17.5 % | 14.6 % | 11.4 % | 11.9 % | 7.5 % | 7.4 % | 8.8 % | 16.6 % |
| Novato | 12.5 % | 9.1 % | 9.9 % | 9.8 % | 8.8 % | 5.8 % | 6.1 % | 5.9 % | 5.2 % | 3.9 % | 44.8 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 10.0 | 7.1 | 5.5 | 4.6 | 4.0 | 3.6 | 3.3 | 3.0 | 2.8 | 2.6 |
| Ligas en que ya quedan 6 o menos (acumulado) | 0 % | 32 % | 79 % | 93 % | 98 % | 99 % | 100 % | 100 % | 100 % | 100 % |

- Fichas del 1.º al terminar: **99175** de media.
- Fichas del 6.º (el corte), cuando llegan 6 o más con fichas (0 % de las ligas): **5058** de media.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.69, aleatorio 0.16 (8 % de los suyos), conservador 0.27 (14 % de los suyos), base 0.63 (31 % de los suyos), novato 1.48 (74 % de los suyos), roca 1.52 (76 % de los suyos), maniaco 0.05 (5 % de los suyos), estacion 0.08 (8 % de los suyos), darwin 0.45 (44 % de los suyos), prudente 0.68 (68 % de los suyos).

### Campo mixto · N = 20

Rivales: 3 aleatorio, 2 conservador, 2 base, 2 novato, 2 roca, 2 maniaco, 2 estacion, 2 darwin, 2 prudente.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 57.8 % [55.6-59.9] | 6.8 % [5.7-7.9] | 17.0 % [14.9-19.2] | 12.6 % [10.8-14.6] | 15.6 % [13.6-17.8] | 7.24 | 15118 |
| Darwin original | 35.8 % [33.7-37.9] | 9.0 % [7.9-10.4] | 17.5 % [14.9-20.4] | 12.7 % [10.5-15.4] | 13.0 % [10.7-15.7] | 5.80 | 15885 |
| Novato | 66.2 % [64.1-68.2] | 14.3 % [12.9-16.0] | 23.0 % [20.8-25.4] | 29.2 % [26.8-31.7] | 29.8 % [27.4-32.4] | 15.16 | 29010 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 11.6 % | 15.3 % | 17.7 % | 16.6 % | 13.7 % | 13.3 % | 11.2 % | 9.6 % | 10.8 % | 8.2 % | 25.3 % |
| Darwin original | 40.3 % | 27.7 % | 20.0 % | 17.7 % | 17.1 % | 16.1 % | 10.4 % | 8.5 % | 9.3 % | 10.2 % | 13.2 % |
| Novato | 14.7 % | 11.4 % | 8.9 % | 10.3 % | 8.2 % | 8.6 % | 5.7 % | 6.7 % | 5.4 % | 5.0 % | 41.0 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 11.9 | 8.3 | 6.2 | 5.1 | 4.4 | 3.9 | 3.5 | 3.2 | 2.9 | 2.7 |
| Ligas en que ya quedan 6 o menos (acumulado) | 0 % | 14 % | 57 % | 82 % | 93 % | 97 % | 99 % | 100 % | 100 % | 100 % |

- Fichas del 1.º al terminar: **131608** de media.
- Fichas del 6.º (el corte), cuando llegan 6 o más con fichas (0 % de las ligas): **6543** de media.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.58, aleatorio 0.15 (5 % de los suyos), conservador 0.13 (7 % de los suyos), base 0.46 (23 % de los suyos), novato 1.32 (66 % de los suyos), roca 1.38 (69 % de los suyos), maniaco 0.04 (2 % de los suyos), estacion 0.08 (4 % de los suyos), darwin 0.71 (36 % de los suyos), prudente 1.14 (57 % de los suyos).

### Campo mixto · N = 30

Rivales: 4 aleatorio, 4 conservador, 3 base, 3 novato, 3 roca, 3 maniaco, 3 estacion, 3 darwin, 3 prudente.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 41.5 % [39.4-43.7] | 4.9 % [4.0-5.9] | 16.5 % [14.1-19.2] | 12.7 % [10.6-15.1] | 14.8 % [12.6-17.4] | 5.12 | 17244 |
| Darwin original | 25.1 % [23.3-27.1] | 7.4 % [6.4-8.7] | 13.5 % [10.8-16.8] | 9.5 % [7.3-12.4] | 8.2 % [6.1-10.9] | 3.80 | 17775 |
| Novato | 54.2 % [52.1-56.4] | 8.9 % [7.7-10.2] | 21.8 % [19.5-24.4] | 28.0 % [25.4-30.8] | 29.9 % [27.2-32.7] | 11.32 | 29468 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 12.7 % | 15.1 % | 14.9 % | 15.8 % | 13.1 % | 11.1 % | 9.3 % | 10.7 % | 8.9 % | 8.1 % | 27.9 % |
| Darwin original | 37.5 % | 25.1 % | 20.7 % | 19.9 % | 15.3 % | 11.9 % | 17.6 % | 8.8 % | 11.4 % | 9.8 % | 13.3 % |
| Novato | 12.8 % | 11.1 % | 9.2 % | 9.2 % | 8.7 % | 7.4 % | 7.7 % | 8.2 % | 6.1 % | 7.2 % | 39.9 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 18.5 | 12.7 | 9.2 | 7.5 | 6.5 | 5.7 | 5.1 | 4.6 | 4.2 | 3.9 |
| Ligas en que ya quedan 6 o menos (acumulado) | 0 % | 0 % | 7 % | 28 % | 49 % | 66 % | 79 % | 88 % | 92 % | 96 % |

- Fichas del 1.º al terminar: **162820** de media.
- Fichas del 6.º (el corte), cuando llegan 6 o más con fichas (15 % de las ligas): **12226** de media.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.41, aleatorio 0.07 (2 % de los suyos), conservador 0.01 (0 % de los suyos), base 0.30 (10 % de los suyos), novato 1.64 (54 % de los suyos), roca 1.62 (54 % de los suyos), maniaco 0.01 (0 % de los suyos), estacion 0.03 (1 % de los suyos), darwin 0.71 (24 % de los suyos), prudente 1.19 (40 % de los suyos).

### Campo mixto · N = 40

Rivales: 5 aleatorio, 5 conservador, 5 base, 4 novato, 4 roca, 4 maniaco, 4 estacion, 4 darwin, 4 prudente.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 29.3 % [27.4-31.4] | 3.7 % [3.0-4.6] | 15.2 % [12.5-18.3] | 13.5 % [10.9-16.5] | 12.8 % [10.3-15.7] | 3.54 | 15869 |
| Darwin original | 17.3 % [15.7-19.0] | 6.2 % [5.3-7.4] | 12.7 % [9.6-16.6] | 6.9 % [4.7-10.1] | 7.8 % [5.4-11.1] | 2.83 | 18421 |
| Novato | 42.0 % [39.9-44.2] | 7.4 % [6.3-8.6] | 22.5 % [19.8-25.4] | 27.2 % [24.3-30.3] | 28.1 % [25.1-31.2] | 8.76 | 29572 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 12.2 % | 17.5 % | 16.5 % | 14.0 % | 14.0 % | 12.9 % | 11.9 % | 10.9 % | 10.0 % | 8.9 % | 25.1 % |
| Darwin original | 40.1 % | 27.0 % | 23.0 % | 19.5 % | 15.1 % | 13.3 % | 14.3 % | 12.3 % | 11.0 % | 8.2 % | 12.2 % |
| Novato | 13.9 % | 12.8 % | 11.5 % | 12.0 % | 8.8 % | 7.4 % | 6.6 % | 7.4 % | 4.9 % | 7.3 % | 37.6 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 24.4 | 16.4 | 11.9 | 9.7 | 8.3 | 7.3 | 6.6 | 6.0 | 5.4 | 5.0 |
| Ligas en que ya quedan 6 o menos (acumulado) | 0 % | 0 % | 0 % | 5 % | 16 % | 31 % | 46 % | 60 % | 72 % | 80 % |

- Fichas del 1.º al terminar: **185973** de media.
- Fichas del 6.º (el corte), cuando llegan 6 o más con fichas (38 % de las ligas): **18427** de media.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.29, aleatorio 0.03 (1 % de los suyos), conservador 0.00 (0 % de los suyos), base 0.28 (6 % de los suyos), novato 1.78 (45 % de los suyos), roca 1.69 (42 % de los suyos), maniaco 0.01 (0 % de los suyos), estacion 0.01 (0 % de los suyos), darwin 0.72 (18 % de los suyos), prudente 1.18 (30 % de los suyos).

## 4. Tablas de valor (N = 20)

Probabilidad de llegar a la final y de terminar 1.º según la ronda que empieza y el stack, medida sobre todos los participantes. Son la base de las "tablas de valor" del plan A (pieza 1).

### Campo kit · N = 20
Tabla de valor · P(llegar a la final) según la ronda que empieza y el stack (como fracción del stack medio de los vivos). "-" = menos de 30 casos.

| Ronda | 0-0.25 | 0.25-0.5 | 0.5-0.75 | 0.75-1 | 1-1.5 | 1.5-2 | 2-3 | ≥3 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | - | - | - | - | 30 % | - | - | - |
| 2 | 10 % | 33 % | 57 % | 57 % | 61 % | 67 % | 71 % | 72 % |
| 3 | 52 % | 77 % | 80 % | 82 % | 85 % | 90 % | 92 % | 96 % |
| 4 | 88 % | 93 % | 95 % | 95 % | 97 % | 97 % | 99 % | 99 % |
| 5 | 96 % | 98 % | 98 % | 99 % | 99 % | 99 % | 100 % | 100 % |
| 6 | 99 % | 99 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % |
| 7 | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % |
| 8 | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | - |
| 9 | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | - |
| 10 | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | - |

Tabla de valor · P(terminar 1.º) según la ronda que empieza y el stack (como fracción del stack medio de los vivos). "-" = menos de 30 casos.

| Ronda | 0-0.25 | 0.25-0.5 | 0.5-0.75 | 0.75-1 | 1-1.5 | 1.5-2 | 2-3 | ≥3 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | - | - | - | - | 5 % | - | - | - |
| 2 | 0 % | 3 % | 7 % | 9 % | 11 % | 13 % | 17 % | 18 % |
| 3 | 2 % | 6 % | 9 % | 13 % | 17 % | 23 % | 28 % | 39 % |
| 4 | 3 % | 7 % | 12 % | 16 % | 22 % | 34 % | 42 % | 51 % |
| 5 | 2 % | 8 % | 13 % | 20 % | 29 % | 43 % | 56 % | 60 % |
| 6 | 2 % | 7 % | 13 % | 24 % | 38 % | 54 % | 65 % | 79 % |
| 7 | 2 % | 5 % | 15 % | 22 % | 44 % | 67 % | 73 % | 88 % |
| 8 | 2 % | 5 % | 12 % | 18 % | 53 % | 77 % | 84 % | - |
| 9 | 1 % | 2 % | 10 % | 17 % | 60 % | 84 % | 90 % | - |
| 10 | 0 % | 1 % | 8 % | 10 % | 67 % | 93 % | 96 % | - |


### Campo agresivo · N = 20

Rivales: 6 maniaco, 4 aleatorio, 5 darwin, 2 estacion, 2 base.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 83.3 % [81.6-84.9] | 11.5 % [10.1-12.9] | 25.9 % [23.8-28.0] | 34.2 % [32.0-36.5] | 35.2 % [33.0-37.6] | 19.31 | 25734 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 21.3 % | 27.3 % | 22.6 % | 14.0 % | 10.5 % | 7.0 % | 7.3 % | 6.8 % | 4.2 % | 5.7 % | 24.8 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 6.0 | 3.0 | 2.1 | 1.7 | 1.5 | 1.4 | 1.3 | 1.3 | 1.2 | 1.2 |
| Ligas en que ya quedan 6 o menos (acumulado) | 71 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % |

- Fichas del 1.º al terminar: **191122** de media.
- Nunca llegan 6 con fichas al final: clasifican los últimos eliminados.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.83, maniaco 0.59 (10 % de los suyos), aleatorio 0.53 (13 % de los suyos), darwin 2.71 (54 % de los suyos), estacion 0.39 (19 % de los suyos), base 0.94 (47 % de los suyos).

Tabla de valor · P(llegar a la final) según la ronda que empieza y el stack (como fracción del stack medio de los vivos). "-" = menos de 30 casos.

| Ronda | 0-0.25 | 0.25-0.5 | 0.5-0.75 | 0.75-1 | 1-1.5 | 1.5-2 | 2-3 | ≥3 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | - | - | - | - | 30 % | - | - | - |
| 2 | 91 % | 90 % | 91 % | 92 % | 97 % | 95 % | 70 % | - |
| 3 | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | - |
| 4 | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | - |
| 5 | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | - |
| 6 | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | - |
| 7 | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | - | - |
| 8 | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | - | - |
| 9 | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | - | - |
| 10 | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | - | - |

Tabla de valor · P(terminar 1.º) según la ronda que empieza y el stack (como fracción del stack medio de los vivos). "-" = menos de 30 casos.

| Ronda | 0-0.25 | 0.25-0.5 | 0.5-0.75 | 0.75-1 | 1-1.5 | 1.5-2 | 2-3 | ≥3 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | - | - | - | - | 5 % | - | - | - |
| 2 | 4 % | 8 % | 13 % | 16 % | 22 % | 21 % | 11 % | - |
| 3 | 5 % | 15 % | 19 % | 29 % | 46 % | 60 % | 62 % | - |
| 4 | 4 % | 15 % | 25 % | 38 % | 69 % | 78 % | 74 % | - |
| 5 | 5 % | 14 % | 26 % | 40 % | 82 % | 84 % | 71 % | - |
| 6 | 4 % | 18 % | 30 % | 40 % | 88 % | 87 % | 73 % | - |
| 7 | 2 % | 15 % | 30 % | 37 % | 92 % | 91 % | - | - |
| 8 | 1 % | 12 % | 30 % | 40 % | 94 % | 94 % | - | - |
| 9 | 1 % | 10 % | 23 % | 44 % | 95 % | 95 % | - | - |
| 10 | 0 % | 3 % | 15 % | 44 % | 97 % | 99 % | - | - |


### Campo tight · N = 20

Rivales: 5 roca, 4 novato, 4 base, 2 conservador, 4 prudente.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 27.6 % [25.6-29.5] | 4.4 % [3.6-5.4] | 11.6 % [9.2-14.6] | 7.8 % [5.8-10.3] | 6.7 % [4.9-9.1] | 2.76 | 8636 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 9.7 % | 15.3 % | 14.8 % | 14.0 % | 14.5 % | 12.3 % | 13.2 % | 12.7 % | 10.7 % | 10.7 % | 25.4 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 16.8 | 13.2 | 11.0 | 9.6 | 8.5 | 7.7 | 7.1 | 6.5 | 6.0 | 5.5 |
| Ligas en que ya quedan 6 o menos (acumulado) | 0 % | 0 % | 0 % | 3 % | 10 % | 21 % | 36 % | 49 % | 61 % | 71 % |

- Fichas del 1.º al terminar: **81464** de media.
- Fichas del 6.º (el corte), cuando llegan 6 o más con fichas (50 % de las ligas): **10422** de media.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.28, roca 2.39 (48 % de los suyos), novato 2.09 (52 % de los suyos), base 0.17 (4 % de los suyos), conservador 0.00 (0 % de los suyos), prudente 1.08 (27 % de los suyos).

Tabla de valor · P(llegar a la final) según la ronda que empieza y el stack (como fracción del stack medio de los vivos). "-" = menos de 30 casos.

| Ronda | 0-0.25 | 0.25-0.5 | 0.5-0.75 | 0.75-1 | 1-1.5 | 1.5-2 | 2-3 | ≥3 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | - | - | - | - | 30 % | - | - | - |
| 2 | 4 % | 11 % | 28 % | 43 % | 45 % | 51 % | 53 % | 59 % |
| 3 | 8 % | 27 % | 42 % | 49 % | 56 % | 62 % | 68 % | 77 % |
| 4 | 15 % | 38 % | 49 % | 57 % | 66 % | 74 % | 79 % | 86 % |
| 5 | 22 % | 46 % | 59 % | 67 % | 74 % | 80 % | 87 % | 92 % |
| 6 | 32 % | 56 % | 68 % | 74 % | 82 % | 86 % | 93 % | 98 % |
| 7 | 43 % | 63 % | 74 % | 80 % | 87 % | 93 % | 96 % | 100 % |
| 8 | 51 % | 70 % | 79 % | 86 % | 92 % | 95 % | 98 % | 100 % |
| 9 | 59 % | 75 % | 86 % | 91 % | 95 % | 97 % | 99 % | 100 % |
| 10 | 64 % | 80 % | 92 % | 95 % | 98 % | 99 % | 100 % | 99 % |

Tabla de valor · P(terminar 1.º) según la ronda que empieza y el stack (como fracción del stack medio de los vivos). "-" = menos de 30 casos.

| Ronda | 0-0.25 | 0.25-0.5 | 0.5-0.75 | 0.75-1 | 1-1.5 | 1.5-2 | 2-3 | ≥3 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | - | - | - | - | 5 % | - | - | - |
| 2 | 1 % | 1 % | 3 % | 6 % | 7 % | 10 % | 15 % | 12 % |
| 3 | 1 % | 2 % | 4 % | 6 % | 10 % | 13 % | 19 % | 32 % |
| 4 | 1 % | 2 % | 4 % | 7 % | 12 % | 16 % | 26 % | 39 % |
| 5 | 0 % | 2 % | 4 % | 8 % | 13 % | 19 % | 33 % | 48 % |
| 6 | 0 % | 2 % | 4 % | 8 % | 14 % | 22 % | 39 % | 62 % |
| 7 | 0 % | 1 % | 4 % | 8 % | 14 % | 26 % | 48 % | 75 % |
| 8 | 0 % | 1 % | 3 % | 7 % | 14 % | 29 % | 59 % | 87 % |
| 9 | 0 % | 1 % | 2 % | 6 % | 14 % | 35 % | 71 % | 92 % |
| 10 | 0 % | 0 % | 1 % | 4 % | 13 % | 43 % | 81 % | 95 % |


### Campo mixto · N = 20

Rivales: 3 aleatorio, 2 conservador, 2 base, 2 novato, 2 roca, 2 maniaco, 2 estacion, 2 darwin, 2 prudente.

| Bot | M1 llega a la final | M1b termina 1.º | M2 tope 150 | M2 tope 500 | M2 tope sin reloj | U | Fichas medias |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 57.8 % [55.6-59.9] | 6.8 % [5.7-7.9] | 17.0 % [14.9-19.2] | 12.6 % [10.8-14.6] | 15.6 % [13.6-17.8] | 7.24 | 15118 |

Quiebra por ronda (de las ligas en que empezó la ronda con fichas):

| Bot | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | Llega vivo al final |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| v0 (prudente) | 11.6 % | 15.3 % | 17.7 % | 16.6 % | 13.7 % | 13.3 % | 11.2 % | 9.6 % | 10.8 % | 8.2 % | 25.3 % |

El campo (ligas con v0 (prudente) como héroe):

| | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | 
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Vivos al terminar la ronda (media) | 11.9 | 8.3 | 6.2 | 5.1 | 4.4 | 3.9 | 3.5 | 3.2 | 2.9 | 2.7 |
| Ligas en que ya quedan 6 o menos (acumulado) | 0 % | 14 % | 57 % | 82 % | 93 % | 97 % | 99 % | 100 % | 100 % | 100 % |

- Fichas del 1.º al terminar: **131608** de media.
- Fichas del 6.º (el corte), cuando llegan 6 o más con fichas (0 % de las ligas): **6543** de media.
- Quién llega a la final (finalistas por liga, de 6): héroe 0.58, aleatorio 0.15 (5 % de los suyos), conservador 0.13 (7 % de los suyos), base 0.46 (23 % de los suyos), novato 1.32 (66 % de los suyos), roca 1.38 (69 % de los suyos), maniaco 0.04 (2 % de los suyos), estacion 0.08 (4 % de los suyos), darwin 0.71 (36 % de los suyos), prudente 1.14 (57 % de los suyos).

Tabla de valor · P(llegar a la final) según la ronda que empieza y el stack (como fracción del stack medio de los vivos). "-" = menos de 30 casos.

| Ronda | 0-0.25 | 0.25-0.5 | 0.5-0.75 | 0.75-1 | 1-1.5 | 1.5-2 | 2-3 | ≥3 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | - | - | - | - | 30 % | - | - | - |
| 2 | 14 % | 34 % | 58 % | 62 % | 64 % | 63 % | 61 % | 60 % |
| 3 | 37 % | 70 % | 77 % | 82 % | 83 % | 84 % | 87 % | 90 % |
| 4 | 72 % | 88 % | 91 % | 92 % | 94 % | 95 % | 97 % | 98 % |
| 5 | 88 % | 95 % | 95 % | 97 % | 98 % | 98 % | 99 % | 100 % |
| 6 | 95 % | 97 % | 99 % | 98 % | 100 % | 99 % | 100 % | 100 % |
| 7 | 97 % | 99 % | 99 % | 100 % | 100 % | 100 % | 100 % | 100 % |
| 8 | 99 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % |
| 9 | 99 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | - |
| 10 | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | 100 % | - |

Tabla de valor · P(terminar 1.º) según la ronda que empieza y el stack (como fracción del stack medio de los vivos). "-" = menos de 30 casos.

| Ronda | 0-0.25 | 0.25-0.5 | 0.5-0.75 | 0.75-1 | 1-1.5 | 1.5-2 | 2-3 | ≥3 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | - | - | - | - | 5 % | - | - | - |
| 2 | 1 % | 3 % | 8 % | 10 % | 12 % | 14 % | 14 % | 17 % |
| 3 | 1 % | 5 % | 8 % | 13 % | 16 % | 20 % | 27 % | 35 % |
| 4 | 2 % | 6 % | 11 % | 14 % | 20 % | 27 % | 39 % | 50 % |
| 5 | 2 % | 6 % | 11 % | 18 % | 24 % | 33 % | 51 % | 65 % |
| 6 | 1 % | 5 % | 12 % | 17 % | 29 % | 44 % | 61 % | 74 % |
| 7 | 1 % | 4 % | 11 % | 18 % | 34 % | 53 % | 71 % | 78 % |
| 8 | 0 % | 4 % | 8 % | 18 % | 38 % | 64 % | 79 % | 89 % |
| 9 | 0 % | 2 % | 9 % | 15 % | 42 % | 73 % | 90 % | - |
| 10 | 0 % | 1 % | 4 % | 10 % | 51 % | 83 % | 96 % | - |
