# Cómo debería jugar el bot: el formato, los profesionales y Lince hoy

- **Fecha**: dom 15/11/2026 · **Quién**: Carlos (con Claude Code; la búsqueda en la web la hicieron dos subagentes)
- **Para qué**: decidir en qué gastar los días que quedan (se congela el código el vie 20/11). Junta tres
  cosas: qué pide el formato, cómo juegan los mejores programas y jugadores, y cómo juega Lince medido en mesas reales.
- **Servidor de pruebas**: el dom 15/11 a las 14:59 seguía sin responder.

## 1. Conclusiones

1. **Lince juega muy distinto de un profesional, y contra los bots del kit eso le conviene.** Entra en el
   doble de manos, apuesta siempre chico y casi nunca paga una apuesta después del flop (se retira o resube).
   Es un estilo que el entrenamiento eligió porque les saca fichas a bots que se retiran ante apuestas chicas
   y no les paga a los que solo apuestan con mano.
2. **Su mayor debilidad frente a un rival bueno es retirarse demasiado**: ante una apuesta después del flop
   se retira el 76 % de las veces. La cuenta que usan profesionales y programas dice que contra una apuesta
   de un tercio del pozo hay que seguir ~75 % de las veces para que al rival no le rinda apostar con cualquier
   cosa. Es la fuga que explotan los bots entrenados como el nuestro (`a3_campo_fuerte.md`), y la adaptación
   al rival de la v1.2 la cubre solo en la final y solo contra quien apuesta muy seguido.
3. **Dos recetas de manual, probadas hoy, casi no mueven la aguja sobre la v1.1**: empujar o retirarse con
   pocas fichas (+2 puntos de llegar a la final solo contra bots que van all-in seguido) y apostar más grande
   con manos más fuertes (+2 soles de premio esperado en el campo kit, nada en los demás). El genoma entrenado
   ya está cerca de lo mejor que da esta política: reentrenado con esos genes libres, el entrenamiento apagó
   el primero, casi no usó el segundo y dio otro compromiso, no una mejora (sección 7).
4. **Lo que sí importa para ganar no es parecerse más a un profesional, sino no ser explotable por los pocos
   rivales buenos que haya entre 30 equipos**, y eso solo se sabe viéndolos jugar.

## 2. Qué pide el formato

Texas Hold'em sin límite. Bases y `pokerfight-desarrollo/docs/competencia.md`:

| Fase | Reglas | Qué premia |
| --- | --- | --- |
| Clasificatoria | 30 equipos. Mesas de 5 (alguna de 6), 10 000 fichas (100 ciegas grandes), ciegas fijas 50/100, 10 rondas de 50 manos con las fichas arrastradas y mesas sorteadas de nuevo. Sin fichas = eliminado. Pasan 6 (20 %): los que más fichas tienen o, si quedan menos de 6, los últimos en caer | **No quebrar** y acumular lo justo. Solo retirarse cuesta 30 fichas por mano: 15 000 en 10 rondas, más que el stack inicial |
| Final | Los 6 empiezan con 10 000. Gana el último en pie, o el que tiene más fichas tras 1 hora | **Ganar todo**: el segundo no cobra. Sin aversión al riesgo |

Particularidades frente a un torneo normal:

- **Las ciegas no suben.** En un torneo normal el stack "se achica" solo y al final todo es empujar o
  retirarse. Aquí se juega siempre profundo (100 ciegas o más), como una partida de dinero, salvo quien perdió fichas.
- **El corte es por fichas entre 30, no por mesa.** Ganar la mesa no sirve si se termina 7.º.
- **En el simulador casi nunca llegan 6 con fichas a la ronda 10** contra campos tipo kit: clasifican los
  últimos eliminados y manda la ronda en que se cae. Contra un campo fuerte sí llegan 9 y deciden las fichas.
- **Los rivales son programas**, con alias estables: lo que se aprende de uno vale después.

## 3. Cómo juegan los mejores programas

Fuentes: prensa y artículos de 2017-2019 sobre Pluribus y Libratus, y glosarios de estrategia. Lo marcado
"sin verificar" no se pudo confirmar en una fuente fiable.

| Lección | Qué hacen | Fuente |
| --- | --- | --- |
| Entrar subiendo, nunca pagando la ciega | Pluribus (6 jugadores) probó pagar la ciega para entrar y lo descartó, salvo desde la ciega chica | [NPR](https://www.npr.org/2019/07/11/740661470/bet-on-the-bot-ai-beats-the-professionals-at-6-player-texas-hold-em), [PokerNews](https://www.pokernews.com/news/2019/07/pluribus-first-ai-to-beat-humans-in-multiplayer-no-limit-34910.htm) |
| Varios tamaños de apuesta, más grandes que los humanos | Pluribus y Libratus apuestan fuerte y variado; Libratus sobreapuesta en el river | [NPR](https://www.npr.org/2019/07/11/740661470/bet-on-the-bot-ai-beats-the-professionals-at-6-player-texas-hold-em), [PokerNews](https://www.pokernews.com/news/2017/12/top-10-poker-stories-computer-beats-pros-29592.htm) |
| Mezclar jugadas con la misma mano | No ser predecible fue su ventaja principal según los profesionales | [Gizmodo](https://gizmodo.com/superhuman-ai-crushes-poker-pros-at-six-player-texas-1836257695) |
| Defensa mínima ante una apuesta | Seguir al menos pozo / (pozo + apuesta): 75 % contra un tercio del pozo, 67 % contra medio, 50 % contra el pozo | [pokerskill](https://www.pokerskill.com/poker-glossary/minimum-defense-frequency/) |
| Faroles según el tamaño | Parte de faroles = apuesta / (pozo + 2 × apuesta): 25 % con medio pozo, 33 % con el pozo | [pokerskill](https://www.pokerskill.com/blog/optimal-bluffing-frequency-formula/) |
| Contra rivales débiles, explotar rinde más que el equilibrio | En las competencias de programas, el formato que suma fichas contra todos premia explotar a los débiles | [IEEE Spectrum](https://spectrum.ieee.org/champions-declared-in-ai-poker-tournament) |
| Modelar al rival | Estadísticas por rival: cuánto entra, cuánto sube, agresión, cuánto se retira; ajustar solo con muestra | [Univ. de Alberta](https://papersdb.cs.ualberta.ca/~papersdb/view_publication.php?pub_id=578) |

Sin verificar: los porcentajes de apertura por posición de Pluribus y su frecuencia exacta de faroles.

## 4. Cómo juegan los profesionales en un formato parecido

Mesas de 5-6 con 100 ciegas. Las cifras de apertura coinciden entre fuentes; las de empujar o retirarse y de
defensa de la ciega grande **no son fiables** (blogs, sin tabla de un programa de cálculo detrás).

| Tema | Regla | Fuente |
| --- | --- | --- |
| Manos que se abren (subiendo) | Primera posición 15-18 %, intermedia ~19 %, cutoff ~25 %, botón 30-43 %, ciega chica ~33 % | [pokerskill](https://www.pokerskill.com/charts/opening-ranges/) |
| Apuesta de continuación | 25-33 % del pozo y muy seguido (64-87 %) en mesas secas; más grande y menos seguido en mesas con proyectos | [DeucesCracked](https://www.deucescracked.com/blog/poker-continuation-bet-cbet-strategy-guide) |
| Jugarse el stack | Con pozo chico frente al stack (relación stack-pozo de 10-15) hacen falta dos pares o mejor; con relación menor de 4 alcanza un buen par | Inferencia, sin fuente directa |
| Supervivencia | Para pagar un all-in cuando quebrar cuesta más de lo que se gana se exige más equidad: factor / (factor + 1), 73-75 % con factor ~3. Pagar un all-in cuesta más que hacerlo | [ICMIZER](https://icmizer.com/en/blog/playing-short-stack-on-bubble) |
| Stack corto | Por debajo de 12-15 ciegas: empujar todo o retirarse | [pokerskill](https://www.pokerskill.com/poker-glossary/nash-push-fold/) (tabla sin respaldo sólido) |
| Mano a mano | El botón abre 70-85 % y la ciega grande defiende 55-65 % | [PokerNews](https://www.pokernews.com/strategy/heads-up-no-limit-hold-em-preflop-strategy-42299.htm) |
| Un solo ganador | Sin premios escalonados, cada ficha vale lo mismo: jugar por fichas esperadas, sin aversión al riesgo | Inferencia |
| Contra el que paga todo | Apostar valor más fino y más grande, casi sin faroles | [poker.academy](https://poker.academy/blog/post/crush-calling-stations-your-guide-to-success) |
| Contra el que se retira mucho | Robar ciegas y apostar más; creerle cuando apuesta fuerte | [pokerstrategy](https://www.pokerstrategy.com/news/content/The-adjustment-to-Nits-and-Calling-stations-is-the-same_128179/) |
| Contra el que va all-in seguido | Pagarle con más manos | Una sola mención, sin umbrales |

## 5. Cómo juega Lince hoy (medido)

1 610 manos de Lince v1.1 en 36 mesas reales de 5 contra los bots del kit (darwin, espia, estadistico, base,
novato, conservador), una ronda de 50 manos, con el registro de manos del bot.

| Posición | Manos | Entra | Sube preflop | Profesional (abre) |
| --- | ---: | ---: | ---: | ---: |
| Primera | 333 | 35 % | 33 % | 15-18 % |
| Intermedia | 220 | 40 % | 40 % | ~19-25 % |
| Botón | 385 | 57 % | 54 % | 30-43 % |
| Ciega chica | 342 | 89 % | 56 % | ~33 % |
| Ciega grande | 330 | 48 % | 41 % | (defiende, sin cifra fiable) |
| Todas | 1 610 | 55 % | 45 % | ~22-30 % |

| Medida | Lince | Referencia |
| --- | ---: | --- |
| Tamaño de la primera subida preflop | 2,0 ciegas (la mínima) | 2,5 |
| Apuesta cuando nadie apostó (flop, turn y river) | 27 % del pozo, siempre | 25-33 % en mesas secas; 66-100 % con proyectos o por valor |
| Apuesta de continuación en el flop | 69 % | 64-87 % |
| Ante una apuesta después del flop | Se retira 76 %, resube 24 %, **paga 0 %** | Seguir ~75 % contra un tercio del pozo |
| Llega al showdown | 20 % de las manos | |

Lectura:

- **Coincide con los programas**: casi nunca entra pagando (salvo en la ciega chica, igual que Pluribus),
  apuesta de continuación chica y muy frecuente, tiene en cuenta la posición (35 % en primera, 57 % en el botón).
- **Se aparta**: entra en el doble de manos; un solo tamaño de apuesta; y sobre todo **nunca paga después del
  flop**. Su regla es "retirarse salvo que la mano alcance para resubir".
- **Por qué le funciona**: los bots del kit se retiran mucho ante apuestas chicas (le regalan pozos) y cuando
  apuestan suelen tener mano (retirarse es correcto). Contra ellos gana el 89 % de las finales reales.
- **Por qué es un riesgo**: cualquier rival que apueste seguido le gana tres de cada cuatro pozos sin mostrar
  cartas. Es exactamente lo que hacen nuestros propios genomas entrenados, contra los que la v1.1 ganaba 1 de 20 finales.

## 6. Lo que se probó hoy

Sobre el genoma de la v1.1, en ligas simuladas de 30 inscritos (1 500 ligas por variante, las mismas para todas).

### Empujar o retirarse con stack corto

Tres genes nuevos (`corto_bb`, `umbral_empuje`, `empuje_vs_subida`): con ese stack o menos, preflop solo se
empuja todo o se retira. En 0 no hacen nada.

| Variante | kit: Δ llegar a la final | agresivo | mixto |
| --- | ---: | ---: | ---: |
| 8 ciegas, umbral 0,50 | 0,0 | **+2,0 [+1,1 a +2,9]** | +0,3 |
| 12 ciegas, umbral 0,50 | +0,5 | **+2,5 [+1,3 a +3,6]** | +0,1 |
| 16 ciegas, umbral 0,50 | +0,3 | **+3,1 [+1,8 a +4,3]** | −0,6 |
| Umbrales más bajos (0,30-0,40) | ~0 | ~0 | −0,5 a −1,4 |

Ayuda solo contra bots que van all-in seguido, y poco. Tiene sentido: aquí las ciegas no suben, así que casi
nunca se está corto; quien llega a 12 ciegas ya perdió casi todo.

### Apuesta de valor más grande con mano más fuerte

Un gen nuevo (`escala_valor`): cuánto del pozo se suma a la apuesta con la mejor mano posible.

| `escala_valor` | kit: U | agresivo: U | tight: U | mixto: U |
| --- | ---: | ---: | ---: | ---: |
| 0 (v1.1) | 43,0 | 36,6 | 33,8 | 40,8 |
| 0,25 | **45,0** | 36,4 | 34,7 | 40,8 |
| 0,50 | 44,4 | 35,6 | 33,4 | 40,3 |
| 1,00 | 42,2 | 30,0 | 26,6 | 35,3 |
| 1,50 | 37,9 | 25,8 | 21,8 | 31,3 |

Un poco más grande ayuda algo en el campo kit (termina 1.º más seguido); mucho más grande hace quebrar más
y baja todo. Los pozos grandes cuestan supervivencia.

## 7. Reentrenar la clasificatoria con los genes nuevos libres (no se promueve)

`entrenar-liga -modo clasificatoria -m2 0.65 -campos kit,agresivo,tight,mixto -n 25,30,35 -poblacion 64 -generaciones 80 -torneos 80`,
partiendo de la v1.1, con los seis genes nuevos libres (adaptación al rival, stack corto y escala de valor).
Resultado en `genoma_clasificatoria_g.json`. Comparación con la v1.1, 2 000 ligas de 30 inscritos por campo:

| Campo | Llega a la final: v1.1 → nuevo | Termina 1.º: v1.1 → nuevo | Gana la final sin tope | Premio esperado U |
| --- | ---: | ---: | ---: | ---: |
| kit | 86,3 → 89,5 % (+3,2) | 26,3 → 18,8 % (−7,6) | 89,5 → 92,0 % | 43,0 → 41,5 |
| agresivo | 80,8 → 77,3 % (−3,5) | 25,6 → 14,9 % (−10,7) | 74,2 → 79,6 % | 36,6 → 31,5 |
| tight | 76,0 → 76,1 % (0) | 11,9 → 9,2 % (−2,8) | 91,8 → 94,4 % | 33,9 → 31,6 |
| mixto | 82,8 → 84,8 % (+2,0) | 25,4 → 18,1 % (−7,3) | 87,0 → 91,7 % | 40,6 → 39,0 |
| fuerte | 27,0 → 31,6 % (+4,7) | 2,7 → 3,4 % | 3,9 → 12,3 % | 2,2 → 3,5 |

Qué eligió el entrenamiento con los genes nuevos:

- **Stack corto: lo apagó** (`corto_bb` = 2,5 ciegas, prácticamente nunca). Coincide con la prueba a mano.
- **Escala de valor: casi nada** (`escala_valor` = 0,11).
- **Adaptación al rival: sí la usa** (`adapta_subida` 0,43 y `adapta_pago` 0,41), también en la clasificatoria.
- Más aversión al riesgo (0,18 contra 0,14): juega más a sobrevivir.

Es **otro compromiso, no una mejora**: llega un poco más a la final en tres campos y gana más finales largas,
pero termina 1.º de la clasificatoria bastante menos y pierde contra el campo agresivo. En premio esperado
queda por debajo de la v1.1 en los cuatro campos tipo kit. Si solo importara el premio mayor y la final fuera
larga, sería algo mejor (la probabilidad de ser campeón sube 2-6 puntos); como la final ya se juega con un
genoma aparte, esa ventaja casi no cuenta. **La clasificatoria sigue con la v1.1.**

Conclusión práctica: con esta forma de política, lo que queda por ganar contra bots tipo kit es poco. Las
dos recetas de manual que parecían más obvias no aportan aquí, y el entrenamiento lo confirma por su cuenta.

## 8. En qué gastar los días que quedan (de más a menos valioso)

1. **Ver rivales reales.** Sigue siendo lo primero: dice si hay que cubrirse contra bots que apuestan seguido
   o si el campo se parece al kit. El bot ya graba las manos y `pruebas/rivales.py` las resume.
2. **Bajar la fuga de "retirarse demasiado" también en la clasificatoria**, sin perder contra el kit: la
   adaptación al rival hoy solo está activa en la final. Probar el genoma de la final en la clasificatoria ya
   se midió y pierde; habría que entrenar la clasificatoria con un campo que tenga algunos bots que apuestan seguido.
3. **Confirmar cuánto dura la final** (ritmo real): decide cuánta aversión al riesgo conviene ahí.
4. **Robustez y ensayo general** (jue 19): no se gana nada si el bot no conecta o se cae.

Lo que **no** conviene: copiar los rangos de un profesional. Están pensados para rivales que también juegan
bien; contra estos bots, el genoma entrenado gana más jugando el doble de manos.
