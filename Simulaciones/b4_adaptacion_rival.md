# B4 · Lince v1.2: adaptación al rival y genoma propio para la final

- **Fecha**: sáb 14/11/2026 (tarde) · **Quién**: Carlos (con Claude Code; dos tareas delegadas a subagentes) · **Rama**: `b/adaptacion-rival` sobre `72f4ff2`
- **Resultado**: **v1.2 promovida.** Clasificatoria con el genoma de la v1.1 (`genoma.json`, sin cambios) y
  **mesa final con un genoma adaptativo** (`genoma_final.json`). El bot ahora lee a cada rival.
- Sigue a `a3_campo_fuerte.md`, que mostró que la v1.1 pierde la final contra bots entrenados como ella.

## 1. Conclusiones

1. **Adaptarse al rival rompe el dilema que no rompían ni mezclar genes ni entrenar equilibrado.** Con dos genes
   nuevos que usan cuán seguido apuesta cada rival, un solo genoma de final gana el 55 % contra finalistas
   fuertes (la v1.1, el 5 %) y casi lo mismo que la v1.1 contra los del kit en finales largas.
2. **La arena real lo confirma, con rivales fuertes de verdad** (nuestros genomas entrenados corriendo como
   bots): la v1.1 gana 1 de 20 finales contra ellos y el adaptativo 10 de 20. El simulador decía 5 % y 55 %.
3. **Cuesta algo contra rivales tipo kit**: en 40 finales reales, 82,5 % contra 92,5 % de la v1.1 (los
   intervalos se solapan), y en el simulador cede unos 13 puntos si la final es corta (tope de 500 manos).
4. Con 30 equipos inscritos (pasan 6, el 20 %), la final va a tener a los mejores: se eligió cubrirse.

## 2. Qué se construyó

- **Lectura del rival** (`estrategia/lectura.go`): por cada rival se cuentan sus apuestas, subidas y all-in
  ("agresivas") y sus pagos ("pasivas"). La lectura es cuánto más agresivo que un rival neutro (45 %) es,
  suavizada con 8 acciones imaginarias para no creerle a una muestra chica. 0 = no se sabe.
- **`Situacion.LecturaRival`**: la lectura del rival que hizo la última apuesta o subida de la calle en curso.
- **Dos genes nuevos** (`estrategia/genoma.go`, `politica.go`), que valen 0 en los genomas anteriores (juegan
  igual que antes):
  - `adapta_subida`: cuánto baja el umbral para resubir contra un rival que apuesta seguido (y sube si apuesta poco).
  - `adapta_pago`: cuánto se abarata pagarle.
- **Simulador** (`entrenamiento/simulador.go`, `liga.go`): cada mesa lleva la cuenta de lo que hace cada
  jugador y se la muestra a las políticas. Se olvida al cambiar de mesa.
- **Bot real** (`rivales.go`, `main.go`; hecho por un subagente y revisado): cuenta las acciones de cada alias
  desde el historial de la mano, con el mismo criterio que el simulador, y lo guarda en `datos/rivales.json`
  para no perderlo en un reinicio. El bot real recuerda a los rivales todo el torneo; el simulador, solo la mesa.
  - Comprobado en el código del kit: el historial de acciones cubre toda la mano, las ciegas no figuran, y un
    all-in que es un pago corto no se distingue de uno que sube (cuenta como agresivo en los dos lados).
- **Entrenador** (`cmd/entrenar-liga -modo final -equilibrado [-peso-fuertes w]`): la mitad de las finales
  contra finalistas tipo kit y la otra mitad contra los genomas de `rivales/`. La aptitud es el peor de los
  dos resultados o, con `-peso-fuertes`, su promedio ponderado.
- **`pruebas/mesas.py`**: alineación F3, una final contra cuatro genomas nuestros (`v1`, `v1_1`, `final_b2`,
  `e`) corriendo como bots, más Novato.

## 3. Simulador: la frontera entre los dos tipos de rival

`simular-liga -n 20 -heroes actual -genoma-final <genoma>`: clasificatoria con la v1.1, final con el genoma
indicado, 2 000 ligas por campo. Probabilidad de ganar la final con tope de 150 / 500 manos / sin tope:

| Genoma en la final | kit | agresivo | tight | mixto | fuerte |
| --- | ---: | ---: | ---: | ---: | ---: |
| v1.1 | 39 / 78 / 89 | 33 / 66 / 71 | 33 / 76 / 92 | 40 / 78 / 88 | 13 / 12 / 5 |
| Adaptativo, peso 90/10 | 32 / 72 / 88 | 20 / 54 / 63 | 34 / 70 / 93 | 35 / 70 / 88 | 25 / 33 / 35 |
| **Adaptativo, peso 75/25 (promovido)** | 28 / 65 / 86 | 16 / 53 / 64 | 30 / 63 / 92 | 31 / 64 / 88 | 30 / 40 / 55 |
| Adaptativo equilibrado (el peor de los dos) | 27 / 53 / 72 | 17 / 39 / 52 | 31 / 53 / 85 | 30 / 54 / 78 | 41 / 61 / 76 |
| Equilibrado sin adaptación (`a3`) | 25 / 37 / 50 | 16 / 21 / 27 | 32 / 49 / 74 | 32 / 43 / 59 | 37 / 49 / 63 |

- Sin tope de manos, el 75/25 pierde 0-3 puntos contra kit, tight y mixto y 7 contra el campo agresivo, y
  gana 50 contra el fuerte.
- En finales cortas todos los adaptativos ceden: juegan con más aversión al riesgo (0,25 contra 0,14) y
  acumulan más despacio. Donde más pierden es contra bots que van all-in seguido.
- Promediando las tres duraciones, cada paso hacia la cobertura compensa si la probabilidad de una final con
  rivales fuertes pasa de ~21 %.
- Genes del promovido que más cambian respecto de la v1.1: `adapta_subida` 0,78, `adapta_pago` 0,18,
  `aversion_riesgo` 0,25, `umbral_subida_post` 0,73.

Un genoma de liga entrenado con adaptación (`genoma_clasificatoria_ad.json`, con `-con-final` y los cinco
campos) **no** mejora la clasificatoria: llega a la final 5-8 puntos menos que la v1.1 en kit, agresivo y
mixto, y solo gana en el campo fuerte. La clasificatoria sigue con la v1.1.

## 4. Arena real: 120 finales

`python bots/lince/pruebas/mesas.py -final -mesas 20 -semilla 7000 -paralelo 4 v11=bots/lince/genoma.json ad25=<genoma>`.
6 jugadores, 10 000 fichas, hasta 2 000 manos; gana quien termina con más fichas. Sin timeouts ni mesas fallidas.

| Final (Lince + ...) | v1.1 | Adaptativo 75/25 |
| --- | ---: | ---: |
| F1: novato, espia, estadistico, darwin, base | 95 % [76-99] | 80 % [58-92] |
| F2: novato, espia, estadistico, darwin, conservador | 90 % [70-97] | 85 % [64-95] |
| **F3: cuatro genomas nuestros entrenados + novato** | **5 % [1-24]** | **50 % [30-70]** |
| Las tres | 63 % [51-74] | 72 % [59-81] |

20 finales por celda es poca muestra: en F1 y F2 la diferencia no está demostrada; en F3, sí.

## 5. Lo que queda abierto

- **Cuánto dura la final de verdad.** Si es corta, la cobertura cuesta más. En el servidor de pruebas se
  jugaron 3 manos por segundo (final larga), pero con bots lentos puede ser otra cosa.
- **Los rivales fuertes son nuestros propios genomas**: el peor caso para nosotros, no una predicción.
- La clasificatoria no usa la adaptación (el genoma de la v1.1 tiene los genes nuevos en 0).
- No se probó la adaptación para el otro lado de la mesa: cuánto farolear según cuánto se retira cada rival.
- Genomas no promovidos, guardados aquí: `genoma_final_ad_equilibrado.json`, `genoma_final_ad10.json`,
  `genoma_final_ad25.json` (copia del promovido) y `genoma_clasificatoria_ad.json`.
