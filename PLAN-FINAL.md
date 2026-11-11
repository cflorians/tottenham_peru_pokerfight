# Plan B · Mesa final: ganar el premio mayor

Parte del plan maestro [`PLAN-BOT.md`](PLAN-BOT.md). El otro plan es [`PLAN-CLASIFICATORIA.md`](PLAN-CLASIFICATORIA.md).
Lo común (infraestructura, simulador, calendario, preguntas, riesgos) está en el maestro.

## Estado

_Última actualización: mar 10/11/2026._ Nada empezado de este plan. La Fase 0 está hecha (v0 en `bots/lince`); depende del simulador (maestro).

## 1. Objetivo

**Ganar la mesa final** ("Campeón general", S/. 60). El 2.º puesto vale lo mismo que el 6.º: nada.

Métrica (intervalo de confianza del 95 %):

- **M2** = % de mesas finales simuladas que gana. El azar puro da **16,7 %**.
- Se mide con tres duraciones (150 manos, 500 manos y sin tope), porque no sabemos el ritmo real.

Este plan solo rinde si el Plan A nos lleva a la final, pero su valor también alimenta al Plan A: `m2` es el
peso de "llegar a la final" en el premio esperado.

## 2. Reglas que mandan en esta fase

| Regla | Consecuencia |
| --- | --- |
| 6 bots, **10 000 fichas cada uno**, ciegas 50/100 | Las fichas de la clasificatoria no cuentan. Todos empiezan iguales |
| Gana el **último en pie** | Un solo ganador: no hay premio por durar. La prudencia de la clasificatoria aquí es un defecto |
| ...o el que tenga **más fichas tras 1 hora** | Hay reloj. Con bots rápidos, 1 h son miles de manos y se define por eliminación; con bots lentos son ~100-200 manos y gana el líder. **Medir ritmo y tiempo** |
| Ciegas fijas | 25 fichas por mano de a 6, 50 de a 3, 75 mano a mano. De a 6 son 10 000 fichas cada 400 manos: esperar cartas es perder |
| Los rivales son los 6 mejores de la clasificatoria | Son los que sobrevivieron: probablemente tight. Y son los mismos 5 durante toda la final |
| Alias estables "durante todo el torneo" | Si valen también en la final (a confirmar), llegamos con perfiles ya aprendidos |

## 3. Estrategia

1. **Fichas = probabilidad de ganar.** Con un solo ganador y jugadores de nivel parecido, la probabilidad de
   ganar es aproximadamente la fracción de fichas que se tiene. Se decide por valor esperado en fichas, sin
   aversión al riesgo. Matiz, según lo que mida M2: si somos mejores que la mesa, evitar los volados parejos
   por todo el stack al principio (la ventaja se cobra jugando muchas manos); si somos peores, buscarlos.
2. **Robar a los que siguen jugando a sobrevivir.** Varios finalistas van a traer la lógica prudente con la
   que clasificaron y no la van a cambiar. Contra ellos: subir seguido, robar ciegas y soltar la mano cuando
   se resisten. Se detecta en pocas manos por cuánto se retiran ante una subida.
3. **Cuidarse del maníaco con suerte.** También puede llegar un bot agresivo que acumuló a los golpes. Contra
   él: no farolear, pagar más liviano con manos hechas.
4. **Mesa corta y mano a mano.** La final se achica hasta quedar dos. Los rangos se abren mucho con menos
   jugadores. Darwin original ya es fuerte mano a mano (1.º en 8 de las 9 pruebas iniciales): es la referencia
   para ese tramo.
5. **Reloj.** Estimar las manos que quedan = tiempo restante × ritmo medido. En el tramo final:
   - si vamos **primeros**, bajar la varianza (no pagar volados, seguir robando barato);
   - si vamos **atrás**, subirla: hace falta pasar al líder en pocas manos, y terminar 2.º no vale nada.
6. **Modelado completo de rivales.** Es donde más rinde: 5 rivales fijos durante cientos de manos, y
   posiblemente ya vistos en la clasificatoria.

## 4. Piezas

1. **Genoma de la final**, distinto del de clasificatoria, entrenado para ganar mesas de 6.
2. **Valor de la final**: probabilidad de ganar según fracción de fichas, jugadores vivos y manos restantes
   estimadas. Tabla sacada del simulador. Sin reloj es casi lineal; cerca de la hora deja de serlo.
3. **Reloj y ritmo**: hora de la primera mano de la final, manos por minuto, manos restantes estimadas.
   Persistido en disco: un reinicio no puede perder la hora de inicio.
4. **Detección de la final**: mesa de 6, todos con 10 000 fichas, y rondas ya jugadas según el estado en
   disco. Flag `-fase` para forzarla. Ante la duda, modo clasificatoria (equivocarse al revés, jugar la
   clasificatoria sin prudencia, cuesta la eliminación).
5. **Perfiles de rivales** heredados de la clasificatoria y afinados durante la final.

## 5. Tareas

### B1 · Simulador de mesa final y línea base (mié 11 - vie 13, junto con el simulador de liga)

- [ ] Mesa final en el simulador rápido: 6 jugadores, 10 000 fichas, hasta que quede uno o hasta un tope de
      manos (150, 500, sin tope); al tope gana el que tiene más fichas.
- [ ] Nivel real sin construir nada: la arena del kit ya tiene el formato `mesa-unica` (hasta que quede uno).
      Configuración de 6 con 10 000 fichas y ciegas 50/100.
- [ ] **Rivales de la final**: los tipos de bot que llegan a la final en las ligas simuladas del Plan A, con
      su frecuencia, por campo.
- [ ] M2 de Darwin original, del genoma prudente y de Novato contra esos rivales. Guardar en `Simulaciones/`.
- [ ] Tabla del valor de la final (probabilidad de ganar según fichas, vivos y manos restantes).

### B2 · Genoma de la final (noche del vie 13 - sáb 14) → versión v2

Es barato: el entrenador ya existe, solo cambia el fitness.

- [ ] Fitness = M2 (ganó o no ganó), promediando las tres duraciones.
- [ ] Rivales: los finalistas simulados, nuestro bot de clasificatoria y campeones de generaciones anteriores
      (coevolución).
- [ ] Detección de la final en su forma simple y cambio de genoma al detectarla.
- [ ] Validar con `mesa-unica` en la arena real (mínimo 300 finales) y promover.

### B3 · Modo final completo (lun 16 - mar 17) → versión v4

- [ ] Reloj y ritmo, persistidos.
- [ ] Decisiones grandes con el valor de la final (reemplaza a `aversion_riesgo` y `umbral_allin`).
- [ ] Tramo final por reloj: probar en el simulador que el líder cuida y el que va atrás arriesga.
- [ ] Mesa corta: comprobar el juego de a 3 y mano a mano contra Darwin original; si el genoma de la final
      pierde ahí, genes aparte para menos de 4 jugadores.
- [ ] Detección robusta: reinicio a mitad de la final, final que empieza tras un silencio largo, final en
      conexión nueva.

### B4 · Modelado en la final (mié 18) · recortable

- [ ] Cargar los perfiles de la clasificatoria al empezar la final.
- [ ] Ajuste rápido por rival: cuánto se retira ante una subida (para robar), cuánto paga (para valor),
      frecuencia de all-in (para pagarle).
- [ ] Rango del rival según su perfil en la equidad contra rangos (pieza del Plan A).

## 6. Metas

- M2 claramente por encima de 16,7 % contra los finalistas simulados, en las tres duraciones.
- M2 del genoma de la final por encima del genoma de clasificatoria jugando la misma final (si no lo supera,
  no se justifica el cambio de genoma).
- Cero fallas de detección en el ensayo general (maestro, Fase 5).
