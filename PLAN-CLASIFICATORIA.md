# Plan A · Clasificatoria: entrar entre los 6 (y pelear el 1.º de la liga)

Parte del plan maestro [`PLAN-BOT.md`](PLAN-BOT.md). El otro plan es [`PLAN-FINAL.md`](PLAN-FINAL.md).
Lo común (infraestructura, simulador, calendario, preguntas, riesgos) está en el maestro.

## Estado

_Última actualización: vie 13/11/2026._ **A1 y A2 hechas; juega la v1.1** (`Simulaciones/a2_b2_v1.md`). En el campo mixto con 20 inscritos (campos nuevos, con Espía y Estadístico): M1 de la v1.1 **84,4 %** (v1 71,9 %, v0 53,1 %, Novato 65,3 %) y M1b 34,2 % (v0 7,9 %). Quiebra por ronda en el mixto: 5,0 %, 5,8 %, 4,1 % y después por debajo de 2 %; en kit y tight ya está en la meta de 3 %. En mesas reales de una ronda quiebra el 6 % (v0: 9 %). Punto flojo: el campo agresivo (12 % y 17 % en las rondas 1 y 2) y que en el campo tight termina 1.º menos que la v1. Sigue A3. **Aviso**: contra un campo de bots entrenados como el nuestro la v1.1 llega a la final el 48 % de las veces y casi no la gana (`Simulaciones/a3_campo_fuerte.md`).

## 1. Objetivo

1. **Principal: llegar a la final** (estar entre los 6). Sin esto no hay premio mayor.
2. **Secundario: terminar 1.º en la clasificatoria** (premio "El mejor de la liga", S/. 30).

Los dos objetivos tiran para lados distintos: para entrar entre 6 conviene cuidar el stack; para ser 1.º hay
que acumular más que nadie. No se elige uno: el bot maximiza el **premio esperado**.

> U(ronda, stack) = 60 × m2 × P(llegar a la final) + 30 × P(terminar 1.º)

- `m2` = probabilidad de ganar la final estando en ella (la mide el Plan B; hasta tener el dato, 0,20).
- Las dos probabilidades salen de tablas del simulador de liga, según ronda y stack relativo.
- Consecuencia: con stack chico o cerca del corte manda la supervivencia; **con stack grande y la final casi
  asegurada, el bot sigue acumulando** para pelear el 1.º, en vez de encerrarse. Quebrar pierde los dos premios.

Métricas (intervalo de confianza del 95 %):

- **M1** = % de ligas simuladas en que llega a la final.
- **M1b** = % de ligas en que termina 1.º.
- Secundarias: supervivencia por ronda, fichas finales, ronda de eliminación.

## 2. Reglas que mandan en esta fase

| Regla | Consecuencia |
| --- | --- |
| Mesas de 5 (algunas de 6), sorteadas en cada ronda | Estrategia para 4-5 rivales. Al quebrar jugadores la mesa se achica: saber jugar de a 3 y de a 2 |
| 10 000 fichas, ciegas fijas 50/100, 50 manos o 6 min | Las ciegas cuestan 30 fichas por mano de a 5: **1 500 por ronda, 15 000 en 10 rondas**. Solo retirarse es quedar eliminado hacia la ronda 7 |
| Fichas arrastradas 10 rondas | Stacks de 3 000 a 40 000, propios y ajenos |
| Sin fichas = eliminado | 10 % de quiebra por ronda es 35 % de llegar a la ronda 10. Meta: ≤3 % por ronda |
| Pasan los 6 con más fichas; si quedan menos de 6, los últimos eliminados | También es una carrera de supervivencia. Con 6 o menos vivos, todos clasificados (a confirmar). Quebrar tarde vale más que quebrar temprano |
| Límite de 6 minutos | Solo aprieta con bots lentos. Contar las manos reales, no suponer 50 |
| Alias estables | Lo aprendido de un rival vale en rondas posteriores. Su stack muestra cómo va el torneo |

## 3. Lo que ya sabemos (`Simulaciones/pruebas_formato_real.md`)

- Darwin original quiebra en el **44 %** de las mesas de 5; Novato, el 10 %.
- El genoma prudente (4 genes a mano) baja a Darwin a **10 %**, pero según `politica.go` nunca dispara su
  regla de all-in y **se retira con la mejor jugada posible ante un all-in grande**: un rival agresivo lo
  desangra. Es el punto de partida (v0), no la solución.
- Quebró el **38 % de los asientos** por mesa. Si el campo real se parece al kit, con 20 inscritos quedarían
  menos de 6 vivos hacia la ronda 3 o 4.
- Límites: 90 mesas es poca muestra (10 % es en realidad 5-18 %), miden una ronda y no la liga, y solo hay
  bots del kit.

## 4. Estrategia

### Por etapa del torneo

| Etapa | Qué pasa | Qué hace el bot |
| --- | --- | --- |
| Rondas 1-3 | Están los bots débiles (aleatorios, maníacos, los que pagan todo) y reparten sus fichas | **Cobrarles** con apuestas de valor grandes y sin faroles; no jugarse el stack en volados parejos |
| Rondas 4-7 | Campo más chico y más fuerte, stacks dispares | Robar ciegas a los tight, presionar a los stacks cortos, evitar choques con stacks que nos cubren |
| Rondas 8-10 | Se define el corte | Según la posición: arriba del corte, cuidar (o ir por el 1.º si el stack da); **debajo del corte, arriesgar**: ser 7.º es igual que ser último |

### Los dos regímenes

El simulador dirá cuál es más probable; el bot tiene que servir en ambos.

- **Quedan menos de 6 con fichas** (campo agresivo): gana el que sobrevive más. Vale durar una ronda más.
- **Llegan más de 6 a la ronda 10** (campo tight): deciden las fichas. Hay que estar por encima del 6.º.

### Piezas

1. **Tablas de valor** `P(final)` y `P(1.º)` por ronda y stack relativo, del simulador. Se usan en las
   decisiones grandes (más de ~25 % del stack): pagar o apostar si mejora `U`. Reemplazan a `aversion_riesgo` y
   `umbral_allin`.
2. **Equidad contra rangos**: contra el X % mejor de las manos posibles del rival en esa mesa, con X según su
   acción (entró pagando ≈ 50 %, subió ≈ 25 %, resubió o fue all-in ≈ 10 %; después son genes). La biblioteca
   de repartos precalcula la fuerza contra 4 rangos fijos, para no frenar el entrenamiento.
3. **Tamaño de apuesta con compromiso**: fracciones del pozo, y el stack entero solo cuando el pozo lo
   justifica. Darwin original va all-in con cualquier pozo apenas pasa un umbral.
4. **Juego según el stack** (en ciegas grandes): <15 BB empujar o retirarse, 15-40 BB juego reducido,
   >100 BB presionar a los cortos.
5. **Posición estimada**: stack medio de los vivos y stack del líder, con los stacks de los alias vistos.
   Señales de campo chico: mesa inicial de menos de 5 o los mismos rivales dos rondas seguidas.
6. **Modelado mínimo** de rivales (perfiles de Espía): detectar en pocas manos al que paga todo, al maníaco y
   al que se retira siempre.

## 5. Tareas

Cada paso se mide contra la versión anterior con la prueba de promoción del maestro.

### A1 · Línea base (mié 11 - vie 13, junto con el simulador)

- [x] Campos de prueba **kit**, **agresivo**, **tight** y **mixto**, con N = 15, 20, 30 y 40 (`entrenamiento/campos.go`).
- [x] Curva de bots vivos por ronda y ronda en que quedan 6 o menos, por campo. Con 20 inscritos, a la
      ronda 4 ya quedan 6 o menos en el 82-100 % de las ligas, salvo en el campo tight.
- [x] Fichas del 6.º y del 1.º al terminar la ronda 10, por campo. Solo en el campo tight llegan 6 con
      fichas (la mitad de las veces); el corte está en ~10 400.
- [x] Tablas `P(final)` y `P(1.º)` por ronda y stack relativo (impresas con `-tabla`; todavía no están en un
      formato que el bot pueda leer).
- [x] M1 y M1b de v0 (genoma prudente), de Darwin original y de Novato. En `Simulaciones/a1_b1_linea_base.md`.

### A2 · Reentrenar lo que ya existe (vie 13) → versión v1

- [x] Fitness = `U` media en ligas simuladas (mismas cartas para toda la generación), en vez de bb/100
      (`entrenamiento/genetico_liga.go`, `cmd/entrenar-liga`).
- [x] Reentrenar los 18 genes actuales sin tocar la política. Tarda 10 minutos, no horas.
- [x] Validar en la arena real (300 mesas por versión, `pruebas/mesas.py`) y promover.

### A3 · Política nueva (sáb 14 - lun 16) → versión v3

De a una pieza: ampliar `Situacion` y el simulador, reentrenar, promover o descartar.

- [ ] Equidad contra rangos (biblioteca con 4 rangos por asiento y calle).
- [ ] Decisiones grandes con las tablas de valor.
- [ ] Tamaño de apuesta con compromiso.
- [ ] Juego según el stack.
- [ ] Posición estimada, con las señales de campo chico.

### A4 · Corte y 1.º de la liga (lun 16 - mar 17)

- [ ] Rondas 8-10: comprobar en el simulador que el bot arriesga debajo del corte y cuida arriba.
- [ ] Con la final asegurada (6 o menos vivos, o `P(final)` ≈ 1): jugar por `P(1.º)`.
- [ ] Stack corto: comparar "empujar ya" contra "aguantar una ronda más" cuando quedan pocos vivos.

### A5 · Modelado mínimo (mié 18) · recortable

- [ ] Perfiles por alias persistidos en disco: cuánto entra, cuánto agrede, cuánto se retira, frecuencia de
      all-in, fuerza de lo que muestra en el showdown.
- [ ] Tres ajustes, cada uno con su prueba: pagar más liviano a maníacos y aleatorios, más valor y cero faroles
      contra quien paga todo, robar ciegas a quien se retira mucho.

## 6. Metas

- M1 por encima de v0 y de Novato en los cuatro campos, con el intervalo sin tocar el cero en el mixto.
- Quiebra por ronda ≤3 % en las rondas 1-7 (en las últimas puede subir a propósito si va debajo del corte).
- M1b: no bajar respecto de v1 al agregar prudencia. Si una pieza sube M1 y baja M1b, decide `U`.
