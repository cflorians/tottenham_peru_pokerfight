# Plan maestro del bot · PokerFight 2026 (v3)

Fechas: sesión de dudas dom 15/11 · **entrega sáb 21/11** · torneo dom 22/11/2026.

Este archivo tiene lo común. La estrategia está en dos planes:

- **[Plan A · Clasificatoria](PLAN-CLASIFICATORIA.md)**: entrar entre los 6 y pelear el 1.º de la liga.
- **[Plan B · Mesa final](PLAN-FINAL.md)**: ganar el premio mayor.

## Estado

_Última actualización: dom 15/11/2026 (análisis de cómo debería jugar el bot; el bot sigue en la v1.2)._

- Hecho: análisis de las bases y de la guía, 90 mesas de prueba en el formato real
  (`Simulaciones/pruebas_formato_real.md`), prueba del genoma prudente (`Simulaciones/genoma_prudente.json`)
  y revisión del plan contra las bases y el código del kit.
- **Fase 0 hecha.** El bot está en `pokerfight-desarrollo/bots/lince/` (etiqueta git `v0`):
  Darwin con el genoma prudente más reconexión, `recover`, supervisor, estado en disco y log. Resultados en
  `Simulaciones/fase0_v0.md`. Quedan dos cosas de la Fase 0 que no son código: la prueba de silencio de
  varias horas y el correo con las preguntas.
- **Fase 1, nivel rápido hecho (vie 13/11, Carlos)**: liga de 10 rondas y mesa final en
  `bots/lince/entrenamiento/liga.go`, campos de rivales en `campos.go` y la herramienta
  `bots/lince/cmd/simular-liga` (M1, M1b, M2, quiebra por ronda, tablas de valor). Líneas base en
  `Simulaciones/a1_b1_linea_base.md`. **Lo que dicen**: Novato supera a la v0 en todo (U de 15,2 contra 7,2 en
  el campo mixto con 20 inscritos), la v0 juega la final peor que el azar, y casi nunca llegan 6 con fichas a
  la ronda 10: clasifican los últimos eliminados.
- **v1 promovida (sáb 14/11, Carlos)**: genoma entrenado con el premio del torneo (`cmd/entrenar-liga`), ahora
  en `genoma.json`; el de la v0 quedó en `genoma_v0.json`. Supera a la v0 en el simulador en los cuatro campos,
  en 300 mesas reales (neto +7 489 contra +5 136, misma quiebra) y sobre todo en la final: **gana el 78 % de
  60 finales reales** contra los bots del kit (v0: 30 %). Detalle en `Simulaciones/a2_b2_v1.md`.
- **v1.1 promovida (sáb 14/11, Carlos)**: la v1 reentrenada valorando llegar a la final con la probabilidad
  real de ganarla (`-m2 0.65` en vez de 0,20) y con Espía y Estadístico en los campos. En el simulador llega a
  la final 10-13 puntos más que la v1 en los cuatro campos (84 % en el mixto con 20 inscritos) y la quiebra de
  la ronda 1 baja de 15 % a 5 %. En la arena real: misma ganancia por ronda, quiebra 6 % contra 10 %, y 89 %
  de 80 finales ganadas (v1: 82 %). Etiquetas git `v1` y `v1.1`.
- **v2 (genoma aparte para la final): no se justifica por ahora.** El entrenado solo para finales no supera
  al de la v1 jugando finales. El bot ya cambia de genoma al detectar la final (`genoma_final.json`), pero hoy
  los dos archivos tienen el mismo genoma.
- **Simulador**: se agregaron réplicas de Espía y Estadístico y una prueba de paridad de una ronda contra la
  arena real (`simular-liga -paridad`), que coincide bien. `pruebas/mesas.py` corre mesas y finales reales en
  serie: es el nivel real liviano. Falta la paridad con fichas arrastradas (10 rondas en la arena real).
- **Servidor de pruebas**: el sáb 14/11 Lince se conectó con el token del equipo y jugó 30 manos; después el
  servidor cerró la conexión y dejó de responder (`Simulaciones/servidor_pruebas.md`). Los `id_mano` son
  `practica-r1-m01-mano-30` y el bot los lee bien. El ritmo fue de 3 manos por segundo: **la final se
  decidiría por eliminación, no por reloj**. Cómo conectarse y cómo guardar el token: `AGENTS.md`.
- **Registro de manos y perfil de rivales (sáb 14/11)**: el bot graba cada mano completa y
  `pruebas/rivales.py` resume cómo juega cada rival. La v1.1 quedó con el supervisor reintentando la conexión
  al servidor de pruebas, que sigue caído desde las 13:02. **Prioridad del equipo: lo que pase en el servidor
  de pruebas** (ver cómo juegan los otros equipos y recalibrar los campos del simulador).
- **Corregido**: si el bot volvía a mitad de una mano, la mesa no le repetía sus cartas y la decisión
  fallaba (jugaba la acción segura con un error en el log). Ahora guarda sus cartas en memoria y en disco y
  sigue la mano. Lo destapó la prueba de robustez.
- **Hallazgo (sáb 14/11, tarde)**: la v1.1 es explotable por bots entrenados como ella. En un campo donde la
  mitad son genomas nuestros, gana la final solo el 5-13 %; y el genoma que les gana a esos pierde contra el
  kit. No hay un genoma mejor que todos, y un selector que elige sobre la marcha según lo que gana no
  funciona. **El bot no cambia** hasta ver rivales reales (`Simulaciones/a3_campo_fuerte.md`).
- **v1.2 promovida (sáb 14/11, Carlos)**: el bot lee cuán seguido apuesta cada rival (dos genes nuevos de
  adaptación) y juega la **mesa final con un genoma adaptativo**; la clasificatoria sigue con el de la v1.1.
  Contra finalistas entrenados como nosotros gana 10 de 20 finales reales (la v1.1, 1 de 20); contra los del
  kit, 82,5 % de 40 (v1.1: 92,5 %). Detalle en `Simulaciones/b4_adaptacion_rival.md`. Etiqueta git `v1.2`.
- **Hay 30 equipos inscritos** (dato del sáb 14/11): pasan 6, el 20 %. Responde la pregunta 1.
- **Servidor de pruebas**: no volvió a responder desde las 13:02 del sáb 14/11. El bot ya no queda
  reintentando; hay que probar a mano de vez en cuando.
- **Análisis de cómo debería jugar (dom 15/11, Carlos)**: `Simulaciones/analisis_como_jugar.md` junta el
  formato, cómo juegan los mejores programas y profesionales (búsqueda en la web, con fuentes) y cómo juega
  Lince medido en 1 610 manos reales: entra en el doble de manos que un profesional, apuesta siempre el
  27 % del pozo y **nunca paga después del flop** (se retira el 76 %). Contra el kit eso gana; contra quien
  apuesta seguido es la fuga. Se probaron dos recetas de manual (empujar o retirarse con stack corto y
  apuesta de valor más grande con mano más fuerte), a mano y reentrenando: casi no aportan. Quedan como
  genes opcionales, en 0. **El bot no cambia.** El servidor de pruebas seguía caído a las 14:59.
- **Próximo paso**: ver rivales reales en el servidor de pruebas (decide si hay que cubrir también la
  clasificatoria contra quien apuesta seguido), confirmar cuánto dura la final, y el ensayo general del jue 19.
  De la política nueva (A3) ya se probó lo que parecía más barato y no rinde contra campos tipo kit. Falta avisar a la organización de las caídas del
  servidor, el correo con las preguntas y la prueba de silencio larga.
- Nombre del bot: **Lince** es un nombre de trabajo; se puede cambiar (es renombrar la carpeta y los imports).
- **Equipo de 2 personas** (una con Claude Code, la otra con Codex) sobre un repositorio privado de GitHub.
  Las reglas de trabajo en equipo están en [`AGENTS.md`](AGENTS.md); el reparto, en la sección 4.

## Reglas de trabajo

- **Solo se escribe código en `pokerfight-desarrollo/`.** La carpeta `pokerfight-arena-participantes/` es la
  copia de referencia del kit: **no se toca** (ni se edita, ni se compila, ni se corre nada dentro).
- Dentro de `pokerfight-desarrollo/`, el kit (`internal/`, `cmd/`, `pkg/`, `docs/`, los bots de ejemplo) tampoco
  se modifica: nuestro código va en `bots/<nuestro-bot>/` y en carpetas nuevas propias. Solo se puede editar
  `torneo-go.json`, para las pruebas locales.
- Los resultados de las simulaciones se guardan en `Simulaciones/`.
- **Control de versiones**: un solo repositorio git (esta carpeta), con remoto privado en GitHub. `master`
  siempre entregable, ramas cortas por tarea (`a/...`, `b/...`, `comun/...`), integración por pull request,
  un commit por cambio medido y una etiqueta por versión (`v0`, `v1`...). El repositorio no se hace público y
  el token nunca se sube: es secreto. El detalle está en [`AGENTS.md`](AGENTS.md).
- **Una sola conexión con el token del equipo a la vez**: avisar antes de conectar al servidor de pruebas.
- Al terminar una tarea o cambiar una decisión, actualizar las casillas y el "Estado" del plan que corresponda.

## 1. Premios y objetivo

| Premio | Cómo se decide | Monto | Plan |
| --- | --- | --- | --- |
| Campeón general | Ganar la mesa final | S/. 60 | A (llegar) + B (ganar) |
| El mejor de la liga | Terminar 1.º en la clasificatoria | S/. 30 | A |
| Bot más payaso | A criterio del jurado | Sorpresa | No se persigue |

> Premio esperado = 60 × M1 × M2 + 30 × M1b

- **M1** = probabilidad de llegar a la final. **M1b** = probabilidad de terminar 1.º en la clasificatoria.
  **M2** = probabilidad de ganar la final estando en ella.
- "Bot más payaso" premia jugadas raras, que cuestan fichas: no se sacrifica nada por él. Lo único gratis es
  el nombre y un README con personalidad.

## 2. Un solo bot, dos modos

Las bases permiten **un bot por equipo**: los dos planes viven en el mismo programa.

| | Modo clasificatoria (Plan A) | Modo final (Plan B) |
| --- | --- | --- |
| Qué se gana | Un cupo entre 6, y el 1.º de la liga | Solo cuenta ganar |
| Fichas | Se arrastran 10 rondas; sin fichas, eliminado | 10 000 para todos |
| Riesgo | Según la posición frente al corte | Neutral; agresivo contra rivales tight |
| Genoma | Entrenado con ligas de 10 rondas | Entrenado con mesas de 6 y un ganador |
| Cuándo | Por defecto | Al detectar la final (mesa de 6, todos con 10 000, rondas ya jugadas) |

Compartido por los dos modos: conexión y reconexión, estado en disco, política por genoma, equidad contra
rangos, perfiles de rivales por alias.

**Reparto del esfuerzo**: primero lo barato de cada plan, después lo caro. Reentrenar con otro fitness cuesta
una noche de máquina; cambiar la política cuesta días. Por eso el genoma de la final (v2) va antes que la
política nueva de la clasificatoria (v3).

## 3. Escalera de versiones

Cada versión es un bot completo y entregable. Una versión nueva reemplaza a la anterior **solo si la supera**;
si no, se entrega la anterior.

| Versión | Plan | Qué agrega | Lista para |
| --- | --- | --- | --- |
| **v0** | Común | Darwin + genoma prudente + infraestructura | mié 11 |
| **v1** | A | Genes actuales reentrenados con fitness de liga. **Hecha el sáb 14; hoy juega la v1.1** | vie 13 |
| **v2** | B | Genoma de la final y detección de la final. **Mecanismo hecho; el genoma aparte no supera al de la v1** | sáb 14 |
| **v3** | A | Política nueva: rangos, tablas de valor, juego según el stack | lun 16 |
| **v4** | B | Modo final completo: reloj, valor de la final, mesa corta | mar 17 |
| **v5** | A + B | Modelado de rivales. **Adelantado en parte: la v1.2 (sáb 14) lee la agresión de cada rival en la final** | mié 18 |

**Prueba de promoción**: mismas semillas para las dos versiones (solo en el simulador rápido: la arena real
no es reproducible); mejora de la métrica del plan (M1 y `U` en A,
M2 en B) con el intervalo del 95 % sin tocar el cero en el campo "mixto", sin empeorar de forma clara en
ningún otro campo y **sin bajar la métrica del otro plan**. Mínimos: 2 000 ligas o 3 000 finales en el
simulador rápido, y 300 mesas en la arena real.

**Regla de recorte**: si el lunes 16 a la noche v3 no está promovida, se salta v5 y el tiempo va a v4 y a la
robustez.

## 4. Calendario

| Días | Común | Plan A | Plan B |
| --- | --- | --- | --- |
| mar 10 - mié 11 | Fase 0: infraestructura → v0 | | |
| mié 11 - vie 13 | Fase 1: simulador de liga y de mesa final, paridad | A1 línea base | B1 línea base |
| vie 13 - sáb 14 | | A2 reentrenar → v1 | B2 genoma de la final → v2 |
| sáb 14 - lun 16 | dom 15: sesión de dudas | A3 política nueva → v3 | (entrenamientos de noche) |
| lun 16 - mar 17 | | A4 corte y 1.º de la liga | B3 modo final completo → v4 |
| mié 18 | | A5 modelado mínimo | B4 modelado en la final → v5 |
| jue 19 | Fase 5: ensayo general | | |
| vie 20 | **Congelar el código** | | |
| sáb 21 | Entrega | | |
| dom 22 | Torneo | | |

### Reparto del equipo (propuesta; confirmar entre los dos y anotar los nombres)

Somos 2: A y B avanzan **en paralelo** desde el viernes 13. Cada persona es dueña de un plan y mantiene su
"Estado"; el trabajo común se reparte así:

| Trabajo | Responsable |
| --- | --- |
| Plan A · Clasificatoria (A1-A5, versiones v1 y v3) | Persona 1: _por definir_ |
| Plan B · Mesa final (B1-B4, versiones v2 y v4) | Persona 2: _por definir_ |
| Fase 1 · Simulador, nivel rápido: liga de 10 rondas | Persona 1 |
| Fase 1 · Simulador, nivel rápido: mesa final (sobre la misma base) | Persona 2 |
| Fase 1 · Nivel real (orquestador) y prueba de paridad | Quien termine antes su parte del nivel rápido |
| Correo con las preguntas, prueba de silencio larga | _por definir_ |
| Fase 5 · Ensayo general, congelamiento y entrega | Los dos juntos |

- El simulador de la Fase 1 bloquea a los dos planes y toca los mismos archivos (`entrenamiento/`): acordar
  primero la interfaz común (cómo se arrastran las fichas y cómo se define el final de una mesa) y fusionar
  esa base antes de separar liga y mesa final.
- Las pruebas de promoción miden siempre las métricas de los dos planes: quien promueve una versión corre
  también la métrica del plan del otro.
- Si una sola persona queda disponible, el orden es el de la escalera de versiones.

## 5. Trabajo común

### Fase 0 · Infraestructura y v0 (mar 10 - mié 11)

- [x] Copiar `bots/darwin` a `bots/lince` y cambiar los imports. Rama git `lince`.
- [x] **Bucle de reconexión** alrededor de `botsdk.Correr` con el mismo token (el SDK no reconecta; el README
      de Darwin dice lo contrario y está mal). Espera creciente de 1 a 5 s, sin límite de reintentos, sin
      ráfagas (el servidor limita intentos por IP).
- [x] **`recover` en `Decidir`**: ante cualquier pánico, devolver la acción segura y dejarlo en el log.
- [x] **Supervisor** (`bots/lince/correr.ps1`): relanza el `.exe` si el proceso muere (una sola instancia).
- [x] **Estado en disco** (`bots/lince/torneo/`; se escribe al cerrar cada mano): ronda, manos de la ronda, fase, hora de inicio de
      la final, historial de stacks, stacks vistos por alias. Se lee al arrancar: un reinicio a mitad del
      torneo no puede dejar al bot creyendo que está en la ronda 1.
- [x] Log a archivo: decisiones, cambios de mesa, reconexiones.
- [x] Nunca enviar `abandono` (el bot no lo usa en ningún camino).
- [x] Verificar que **nuestro `.exe` no lo bloquee Smart App Control**: `bin/lince.exe` arranca. Pero el
      bloqueo cambia con cada compilación (ver "Lo aprendido" abajo).
- [x] v0 = todo lo anterior con `genoma_prudente.json`. Etiquetada.
- [x] Prueba automática de robustez (`bots/lince/pruebas/robustez.py`): silencio, reconexión entre mesas,
      conteo de rondas, detección de la final y proceso matado a mitad de la final.
- [ ] Prueba de silencio de varias horas: `python bots/lince/pruebas/robustez.py -silencio 14400` (se probó
      con 40 s y con 10 minutos).
- [ ] Mandar por correo las preguntas de la sección 6.

**Lo aprendido en la Fase 0** (cambia cosas de más adelante):

- **Smart App Control es una lotería por binario.** Bloqueó el ejecutable de los tests y lo dejó pasar al
  recompilar con otras opciones; el mismo código puede pasar o no. Regla: compilar la versión final, comprobar
  que arranca en la PC del torneo y **no volver a compilarla**. `go run` no es una salida segura: también
  genera un `.exe`.
- **Las mesas de la arena real no son reproducibles**: la misma semilla con los mismos bots da resultados
  distintos (algún bot del kit usa azar sin sembrar). En la arena real no hay comparaciones "con las mismas
  cartas": eso solo existe en el simulador rápido. Las pruebas de promoción en la arena necesitan más mesas.
- **En el servidor de pruebas hay que lanzar con `-Fase clasificatoria`**: ahí todos vuelven a 10 000 fichas en
  cada ronda y una mesa de 6 se confundiría con la final.
- **Al empezar el torneo, lanzar con `-Nuevo`** (estado limpio). Un estado viejo con rondas ya jugadas haría
  que una primera mesa de 6 parezca la final.
- La detección de ronda usa el id de mano del kit (`mesa-1-mano-7`). No sabemos qué forma tiene en el
  servidor del torneo: hay pistas de respaldo, pero conviene mirarlo apenas encienda el servidor de pruebas.

### Fase 1 · Simulador (mié 11 - vie 13)

La arena del kit no juega 10 rondas con fichas arrastradas. Herramienta propia en carpeta nueva, sin tocar el kit.

**Nivel rápido** (para entrenar), sobre el simulador de `entrenamiento/`:

- [x] Liga de N bots (`entrenamiento.JugarLiga`): sorteo de mesas de 5-6, 10 rondas de 50 manos, fichas
      arrastradas, eliminación, y el corte real: top 6 por fichas o, si quedan menos de 6, los últimos
      eliminados. Los supuestos donde las bases callan están comentados en `liga.go` (preguntas 2 a 4).
- [x] Mesa final (`entrenamiento.JugarFinal`): 6 jugadores, 10 000 fichas, hasta que quede uno o hasta un
      tope de manos.
- [x] Reporte por bot (`cmd/simular-liga`): M1, M1b, M2, quiebra por ronda, fichas finales, y del campo:
      vivos por ronda, fichas del 1.º y del 6.º, quién llega a la final, tablas de valor.
- [x] Comparación pareada de dos versiones: con varios bots en `-heroes`, `simular-liga` da la diferencia
      contra el primero liga por liga, con su intervalo.
- [x] Réplicas de Michi Espía y Bot Estadístico en el simulador (`replicas.go`), dentro de los campos kit y
      mixto. Falta la de Gato Bandido.

**Nivel real** (para validar), con los `.exe` y la mesa de verdad:

- [ ] Orquestador propio que usa `internal/mesa` desde nuestra carpeta (mismo módulo, no hace falta editar el
      kit): `ConfigPartida.StackPorJugador` ya permite sentar a cada bot con sus fichas. El límite de
      6 minutos no existe en la mesa local: lo pone el orquestador.
- [x] La final se prueba con el formato `mesa-unica` de la arena (`pruebas/mesas.py -final`).
- [x] **Prueba de paridad de una ronda** (`simular-liga -paridad` contra las mesas reales): neto y quiebra
      por bot coinciden en orden y magnitud (`Simulaciones/a2_b2_v1.md`, sección 5).
- [ ] Paridad con fichas arrastradas: 10 rondas en la arena real (necesita el orquestador).

Las líneas base de cada plan (A1 y B1) se corren apenas esté el simulador.

### Fase 5 · Ensayo general y congelamiento (jue 19 - vie 20)

- [ ] **Ensayo general**: una liga completa de 10 rondas **más la final** en la arena real, con el supervisor,
      un corte de red a mano y el proceso matado a mitad de una ronda y a mitad de la final. Revisar que el
      cambio de modo ocurra en el momento justo y el log entero.
- [ ] Resistencia: horas en silencio, mesas de 2 a 6, stacks de 500 a 60 000, rivales lentos (cerca de 5 s).
- [ ] **Servidor de pruebas** (cualquier día, apenas encienda; hoy figura como no disponible): probar el token,
      medir la latencia y el ritmo real, y **registrar cómo juegan los rivales de verdad** para recalibrar los
      campos de prueba de los dos planes.
- [ ] `git pull` diario del kit, recompilar y releer `docs/competencia.md`.
- [ ] README de una página (idea, técnicas, resultados; aclarar que parte de Gato Darwin).
- [ ] **Congelar el código el viernes 20.** El sábado solo se corrigen errores y se entrega.

## 6. Preguntas para la organización (por correo ya; el resto, el domingo 15)

| # | Pregunta | Afecta a |
| - | --- | --- |
| 1 | ¿Cuántos equipos hay inscritos? **Respondida: 30** (pasa el 20 %) | A |
| 2 | Si quedan 6 o menos bots con fichas antes de la ronda 10, ¿ya están todos clasificados? ¿Cómo se ordena a los eliminados en una misma ronda? ¿Y un empate en fichas en el corte? | A |
| 3 | ¿Cómo se arman las mesas cuando quedan pocos bots (7, 8, 9, o menos de 5)? | A |
| 4 | ¿Cómo se decide "el primero de la clasificatoria" si el líder en fichas quiebra después, o si quedan menos de 6 vivos? | A |
| 5 | ¿Hay pausa entre manos o entre rondas, o el ritmo lo ponen los bots? Al cumplirse los 6 minutos o la hora de la final, ¿se termina la mano en curso? | A y B |
| 6 | La final: ¿misma conexión o sesión nueva?, ¿se mantienen los alias de la clasificatoria?, ¿se avisa el inicio?, ¿la hora corre desde la primera mano? | B |
| 7 | ¿Se puede relanzar el bot (mismo código, otro flag) entre la clasificatoria y la final? | B |
| 8 | ¿Se avisa de algún modo el cambio de ronda, o hay que deducirlo? | A |
| 9 | ¿Los timeouts restan algo en el torneo, además de la acción segura? | A y B |
| 10 | ¿Hay alguna regla sobre cuánto tardar en responder dentro del plazo? (Por defecto respondemos rápido: no demorar a propósito para gastar el reloj.) | B |
| 11 | Confirmar que se puede guardar en disco lo aprendido de los rivales entre reinicios. | A y B |
| 12 | ¿Cuándo enciende el servidor de pruebas? | Común |
| 13 | ¿Qué forma tienen los `id_mano` en el torneo? ¿Cambian de prefijo en cada mesa? | Común |

## 7. Riesgos y plan B

| Riesgo | Mitigación |
| --- | --- |
| El día del torneo el bot no conecta (token nunca probado, servidor nuevo) | Probar el token con `conservador` apenas haya servidor; llegar con tiempo; tener a mano la dirección vigente |
| Windows bloquea el `.exe` | Compilar, comprobar que arranca y no recompilar; si lo bloquea, recompilar con otras opciones (`go run` no sirve: también genera un `.exe`); probar también en la máquina de respaldo |
| El proceso muere o la PC se suspende | Supervisor, estado en disco, suspensión y actualizaciones desactivadas |
| Se cae la red | Reconexión sin límite; red de respaldo (celular) lista; segunda máquina con el mismo binario, apagada (dos conexiones con el mismo token se pisan) |
| El bot se cree en la final estando en la clasificatoria (o al revés) | Estado en disco, flag `-fase`, y ante la duda modo clasificatoria |
| Un plan mejora y rompe al otro | La prueba de promoción mide siempre las métricas de los dos |
| El simulador rápido no se parece a la mesa real | Prueba de paridad antes de entrenar; validar cada versión en la arena real |
| Sobreajuste a los bots del kit | Cinco campos de prueba (el quinto, con genomas entrenados como el nuestro), coevolución, datos del servidor de pruebas. Riesgo abierto: la v1.1 pierde la final contra bots entrenados; hay genomas de cobertura listos (`Simulaciones/a3_campo_fuerte.md`) |
| Atraso | Escalera de versiones y regla de recorte: siempre hay algo entregable |
| Cambian las bases o el protocolo | `git pull` diario, recompilar, releer `docs/competencia.md` |

## 8. Entrega y día del torneo

**Entrega (sáb 21)**: código + README de una página. El jurado revisa el código. Lo que corre el domingo es
exactamente lo entregado.

**Torneo (dom 22)**:

- PC enchufada, sin suspensión ni actualizaciones, red por cable si se puede, respaldo listo.
- Bot lanzado antes de la hora con el supervisor, el token del equipo y `-Nuevo`:
  `powershell -ExecutionPolicy Bypass -File bots\lince\correr.ps1 -Addr HOST:PUERTO -Nuevo` (el token sale de
  la variable `POKERFIGHT_TOKEN`).
  **Una sola instancia.** Si hay que relanzar a mitad del torneo, **sin** `-Nuevo`.
- El `.exe` es el mismo que se probó: no recompilar ese día.
- Mirar el log en vivo, sin tocar nada: no se permite intervenir durante las partidas.

## 9. Próximo paso concreto

A3 (política nueva), empezando por lo que más duele hoy: los choques con bots que van all-in seguido. En paralelo: dejar el bot conectado al servidor de pruebas cuando esté estable,
avisar a la organización de las caídas, mandar el correo con las preguntas y la prueba de silencio larga.

## Historial del plan

- **v1 (original)**: un solo plan, centrado en sobrevivir la clasificatoria. Eliminado; queda en el historial
  de git (commit inicial, `PLAN-BOT.anterior.md`).
- **v2 (10/11)**: revisión de puntos débiles. El premio mayor se gana en la final, la meta de 10 % de quiebra
  por ronda no alcanza, las ciegas cuestan 15 000 fichas en 10 rondas, faltaba la regla de los últimos
  eliminados, el genoma prudente se retira con manos ganadoras, y no había versión entregable ni robustez
  hasta el final.
- **v3 (10/11)**: separación en Plan A y Plan B, premio esperado como objetivo único, y lo barato de cada
  plan primero.
