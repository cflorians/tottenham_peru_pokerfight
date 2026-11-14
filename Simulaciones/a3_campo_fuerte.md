# A3 · ¿Y si otros equipos entrenan como nosotros? Campo fuerte, final aparte y selector

- **Fecha**: sáb 14/11/2026 (tarde) · **Quién**: Carlos (con Claude Code) · **Rama**: `a/campo-fuerte` sobre `6596c60`
- **Herramientas**: `cmd/simular-liga` (nuevos: `-diagnostico`, `-genoma-final`, campo `fuerte`) y `cmd/entrenar-liga` (nuevo: `-con-final`)
- **Tamaño**: 2 000 ligas por bot y campo, 20 inscritos, semilla 2026, salvo donde se indica.
- **Resultado**: **el bot no cambia** (sigue la v1.1 en las dos fases). Quedan un hallazgo importante, tres
  genomas candidatos y dos ideas descartadas con sus números.

## 1. Conclusiones

1. **La v1.1 es muy buena contra bots como los del kit y mala contra bots entrenados como ella.** En un campo
   donde la mitad de los inscritos son genomas nuestros ya entrenados ("fuerte"), llega a la final el 48 % de
   las veces pero la gana solo el **5-13 %** (el azar es 16,7 %). Contra los otros cuatro campos la gana el
   71-92 % sin tope de manos.
2. **No hay un genoma mejor que todos: hay ciclos, como piedra, papel o tijera.** El que más gana contra el
   kit pierde contra los entrenados, y el que les gana a los entrenados pierde contra el kit.
3. **No sabemos cuál de los dos mundos es el real.** Con lo que hay hoy, cambiar de genoma es apostar a ciegas:
   la cobertura intermedia (`e` en la final) solo compensa si hay más de un ~25 % de probabilidad de que en
   la final haya bots que nos exploten así. Por eso **el servidor de pruebas es la prioridad**: es el único
   lugar donde se ve cómo juegan los otros equipos (el bot ya graba las manos; `pruebas/rivales.py`).
4. **Un selector que elige el genoma sobre la marcha según cuánto gana no sirve**: probar un solo bloque con
   el genoma equivocado ya arruina la final. Descartado y quitado del código.

## 2. De dónde quiebra la v1.1 (diagnóstico)

`go run ./bots/lince/cmd/simular-liga -diagnostico -heroes actual -ligas 20000`: mesas sueltas de una ronda,
todos con 10 000 fichas; se anota la última jugada del bot en la mano en que se queda sin fichas.

| Campo | Quiebra por ronda | De qué (porcentaje de las quiebras) |
| --- | ---: | --- |
| agresivo | 13,5 % | 37 % ya estaba casi sin fichas (el golpe fue antes); 9 % paga preflop más del 60 % del stack con manos de ~57 %; 9 % sube preflop comprometiendo todo con ~64 %; 16 % all-in con manos de 95 % (mala suerte) |
| mixto | 5,1 % | 23 % ya estaba casi sin fichas; 24 % all-in con manos de 95 %; 11 % paga en el river con 86 %; 9 % faroles que comprometen todo el stack |
| tight | 2,1 % | 24 % paga en el river con 89 %; 33 % all-in con manos de 95 % |

No hay una sola fuga grande. Lo arreglable (faroles con todo el stack, pagos y subidas preflop con manos
medianas contra bots que van all-in) suma entre el 10 % y el 20 % de las quiebras: poco margen.

## 3. El campo fuerte

`entrenamiento.Campos` agrega el campo **fuerte** si hay genomas en `bots/lince/rivales/`: la mitad de los
inscritos son esos genomas (v1, v1.1, el especialista en finales B2 y `e`) y el resto, los mejores bots del
kit (novato, espía, estadístico, roca). Es un caso extremo a propósito.

En el campo fuerte llegan 9 bots con fichas a la ronda 10 y el corte está en ~14 000 fichas: ahí sí deciden
las fichas, no la ronda en que se cae.

## 4. Genomas candidatos

| Genoma | Cómo se entrenó | Dónde está |
| --- | --- | --- |
| **v1.1** (el del bot) | Liga, 4 campos, M2 fija en 0,65 | `bots/lince/genoma.json` |
| B2 | Solo finales, contra finalistas del kit | `bots/lince/rivales/final_b2.json` |
| `e` | Liga con 5 campos (incluido el fuerte) y **jugando la final de verdad** (`-con-final`) | `bots/lince/rivales/e.json` |
| F2 | Solo finales, con 40 % de finalistas entrenados | `Simulaciones/genoma_final_f2.json` |

### Un solo genoma para todo el torneo (premio esperado U, en soles)

| Campo | v1.1 | `e` | B2 |
| --- | ---: | ---: | ---: |
| kit | **48,9** | 41,2 | |
| agresivo | **40,2** | 34,6 | |
| tight | **40,5** | 36,5 | |
| mixto | **45,1** | 37,6 | 25,6 |
| fuerte | 3,9 | 13,0 | **14,4** |

(El campo fuerte de esta tabla tenía tres genomas entrenados; el de la siguiente, cuatro.)

### v1.1 en la clasificatoria y otro genoma en la final (`-genoma-final`)

Probabilidad de ganar la final estando en ella (M2), sin tope de manos / con tope de 500:

| Campo | Final con v1.1 | Final con `e` | Final con B2 | Final con F2 |
| --- | ---: | ---: | ---: | ---: |
| kit | **89 % / 78 %** | 79 % / 68 % | 72 % / 58 % | 8 % / 10 % |
| agresivo | **71 % / 66 %** | 66 % / 64 % | 60 % / 54 % | 14 % / 15 % |
| tight | **92 % / 76 %** | 83 % / 66 % | 82 % / 67 % | 6 % / 8 % |
| mixto | **88 % / 78 %** | 81 % / 68 % | 75 % / 60 % | 10 % / 10 % |
| fuerte | 5 % / 12 % | 30 % / 26 % | 36 % / 30 % | **54 % / 50 %** |

- F2 es el caso más claro del ciclo: gana más de la mitad de las finales contra bots entrenados y casi
  ninguna contra los del kit.
- `e` es la cobertura: cede ~8 puntos contra el kit y gana ~24 contra los entrenados. Compensa solo si la
  probabilidad de una final "fuerte" pasa del 25 %.

## 5. Lo que no funcionó

- **Selector adaptativo (bandido UCB1 sobre varios genomas en la final)**: juega bloques de manos con cada
  genoma y se queda con el que más ciegas por mano gana. Con v1.1 + F2 gana el 50 % de las finales del campo
  fuerte, pero **baja del 88 % al 40-51 % en el campo mixto**, con cualquier tamaño de bloque (15 a 80 manos)
  y nivel de exploración, con o sin recortar los pozos grandes. Motivo: el resultado en fichas de un bloque
  es puro ruido frente a la diferencia entre genomas, y un solo bloque con F2 contra bots del kit ya cuesta
  la final. Con v1.1 + `e` no arruina nada (84 %) pero tampoco ayuda contra los fuertes (17 %). Quitado del
  código.
- **Entrenar un solo genoma contra los cinco campos** (`e`): mejora en el fuerte y empeora en los otros
  cuatro. No es una mejora, es otro punto del mismo compromiso.

## 6. Qué haría falta para decidir

1. **Ver a los rivales reales** en el servidor de pruebas y clasificarlos: ¿se parecen a los bots del kit o
   juegan como un genoma entrenado (abren al mínimo, apuestan chico y seguido, no pagan apuestas grandes)?
2. Si aparecen bots fuertes, una forma mejor de adaptarse que el selector: **reconocer al rival por cómo
   juega** (sus estadísticas, que se juntan en cada mano y sin arriesgar fichas) y no por cuánto ganamos.
   Es el modelado de rivales del plan (v5).

## 7. Límites

- El campo fuerte está hecho con nuestros propios genomas: es el peor caso para nosotros, no una predicción.
- Todo es simulador rápido; nada de esto se validó en la arena real porque el bot no cambió.
