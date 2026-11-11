# Plan del bot · PokerFight 2026

Fechas: sesión de dudas dom 15/11 · **entrega sáb 21/11** · torneo dom 22/11/2026.

## Estado

_Última actualización: mar 10/11/2026._

- Hecho: análisis de las bases y de la guía, 90 mesas de prueba en el formato real
  (`Simulaciones/pruebas_formato_real.md`) y prueba del genoma prudente (`Simulaciones/genoma_prudente.json`).
- Todavía no hay código del bot. **Próximo paso: Fase 0** (copiar Darwin a `pokerfight-desarrollo/bots/<nuestro-bot>/`).
- Nombre del bot: por definir.

## Reglas de trabajo

- **Solo se escribe código en `pokerfight-desarrollo/`.** La carpeta `pokerfight-arena-participantes/` es la
  copia de referencia del kit: **no se toca** (ni se edita, ni se compila, ni se corre nada dentro).
- Dentro de `pokerfight-desarrollo/`, el kit (`internal/`, `cmd/`, `pkg/`, `docs/`, los bots de ejemplo) tampoco
  se modifica: nuestro código va en `bots/<nuestro-bot>/` y en carpetas nuevas propias. Solo se puede editar
  `torneo-go.json`, para las pruebas locales.
- Los resultados de las simulaciones se guardan en `Simulaciones/`.

## 1. Qué premian realmente las reglas

| Regla (bases y `docs/competencia.md`) | Consecuencia para el bot |
| --- | --- |
| Mesas de **5** (algunas de 6), no heads-up | La estrategia tiene que estar pensada para 4-5 rivales. Una mano que es buena contra uno es mediocre contra cuatro |
| 10 000 fichas, ciegas **fijas 50/100**, 50 manos o 6 min por mesa | Juego profundo (100 ciegas grandes) y con ciegas baratas: aguantar cuesta poco, ir all-in con manos medianas no hace falta |
| **10 rondas** y las fichas se arrastran de una ronda a la siguiente | El stack de cada mesa varía (puedes tener 3 000 o 30 000 fichas). El bot tiene que adaptarse a stacks cortos y profundos |
| **Quedarse sin fichas = eliminado** para siempre | **Sobrevivir es lo primero.** Un bot que quiebra en el 40 % de las mesas casi seguro no llega a la ronda 10 |
| Pasan los **6 con más fichas** (si quedan menos de 6 con fichas, los últimos eliminados) | Gana el que acumula sin arriesgar la eliminación, no el que gana más mesas |
| Final: 6 bots, 10 000 fichas cada uno, último en pie o más fichas tras 1 h | Es otro juego (mesa larga, sin arrastre de fichas): conviene un modo final con otros parámetros |
| Alias aleatorios pero **estables** durante todo el torneo | El modelado de rivales sirve: lo que aprendes de `j-a1b2c3` en la ronda 2 vale en la ronda 7 |
| 5 s por jugada; acción inválida = timeout; la conexión dura todo el torneo y queda en silencio entre rondas | Responder en <1 s, solo acciones válidas, **reconectar solo**, no cerrar el socket por estar en silencio |
| Premio aparte para el **primero de la clasificatoria** | Fichas al final de la ronda 10 = métrica a optimizar |

## 2. Lo que dicen las simulaciones

Tus 9 pruebas (`Simulaciones/pruebas_iniciales.md`) son heads-up, 200 ciegas grandes, con fichas
repuestas y con `aleatorio` regalando 90 000-210 000 fichas por torneo. Ahí Darwin gana
claramente (1.º en 8 de 9). Pero ese no es el formato del torneo.

Corrí 90 mesas en el formato real (detalle en `Simulaciones/pruebas_formato_real.md`):

- **Darwin quiebra en el 44 % de las mesas** de 5. Gana muchas mesas (31/90), pero cuando pierde, pierde todo.
- **Novato** (reglas simples y tight en Python) tiene el mejor neto medio y quiebra solo el 10 %.
- Con **4 genes de Darwin cambiados a mano** (más aversión al riesgo, umbral de all-in más alto, menos faroles),
  la quiebra baja de 44 % a **10 %** y el neto medio casi no cambia.

Conclusión: la técnica de Darwin sirve, pero hay que **reentrenarlo para el formato real, con la supervivencia como objetivo**.

## 3. Decisión: con qué bot empezar

**Partir de Gato Darwin (Go)**, copiado a `pokerfight-desarrollo/bots/<nuestro-bot>/`, y sumarle el modelo de
rivales de Michi Espía.

Por qué Darwin:

1. Es el más fuerte de los ejemplos y el que más mesas de 5 gana.
2. Trae lo más valioso del kit: un **entrenador genético** con un simulador que copia las reglas
   de la mesa (pozos laterales, subida mínima, all-in corto) y que ya sabe jugar mini-torneos de
   5-6 jugadores con eliminación. Reentrenar cuesta minutos, no días.
3. Su estrategia es un genoma JSON legible: se puede ajustar a mano y medir.
4. Piensa en ≤400 ms: no hay riesgo de timeouts.
5. Está en Go con el SDK oficial, así que el protocolo ya está resuelto.

Lo que tomamos de los otros:

- **Espía** (`modelo.go`): perfiles bayesianos por rival (cuánto entra, cuánto agrede, cuánto se retira).
- **Novato**: es la referencia de "disciplina". Nuestro bot debería quebrar tan poco como él.
- **Estadístico, Espía y Darwin original**: rivales de entrenamiento. Es probable que muchos equipos
  presenten variantes de estos bots.

Descartados como base: Novato (reglas a mano, sin forma de entrenarlo), Estadístico y Espía (quiebran
~35-40 % y no traen entrenador), Bandido (débil, y además Windows bloquea su `.exe`).

## 4. Plan de desarrollo

### Fase 0 · Infraestructura (mar 10 - mié 11)

- [ ] Copiar `bots/darwin` a `bots/<nuestro-bot>` y cambiar los imports. No tocar nada fuera de esa carpeta
      (lo pide la guía: el servidor real no usa nuestra copia del kit).
- [ ] **Bucle de reconexión** alrededor de `botsdk.Correr` con el mismo token (el SDK no reconecta, y el
      README de Darwin dice lo contrario: está mal). Espera creciente de 1 a 5 s y sin límite de reintentos.
- [ ] Log a archivo (decisiones + cambios de mesa) para revisar después de cada sesión.
- [ ] Verificar que **nuestro `.exe` no lo bloquee Smart App Control** (le pasó a `bandido.exe`). Si pasa:
      correr con `go run .`, o desactivar Smart App Control en la PC del torneo.

### Fase 1 · Simulador de la clasificatoria completa (mié 11 - jue 12)

La arena del kit no reproduce 10 rondas con fichas arrastradas. Construir **nuestra propia herramienta**
(en nuestra carpeta, sin tocar el kit) que:

- sortee mesas de 5-6 entre ~15-30 bots (copias de los ejemplos + nuestro bot), con 10 rondas, fichas
  arrastradas y eliminación;
- reporte por bot: % de supervivencia a 10 rondas, fichas finales medias, % de top 6 y % de 1.º.

Dos niveles: uno **rápido**, con el simulador de `bots/darwin/entrenamiento` (miles de ligas por minuto, para
entrenar), y uno **real**, con los `.exe` y la mesa de verdad (más lento, para validar).

Métrica principal: **% de veces que entra al top 6**. Secundaria: fichas finales medias.

### Fase 2 · Política "sobrevivir primero" (jue 12 - sáb 14)

1. **Equidad contra rangos, no contra una mano al azar.** Es el defecto principal de Darwin: cuando alguien
   sube fuerte o va all-in, su rango no es aleatorio. Simular contra el top X % de manos según la acción
   (y, desde la fase 4, según el perfil de ese rival).
2. **Regla de stack-off**: comprometer más de ~30 % del stack solo con fuerza muy alta contra el rango del rival.
3. **Juego según el stack** (en ciegas grandes): <15 BB push/fold, 15-40 BB juego reducido, >100 BB presionar
   a los stacks cortos.
4. **Modo protección**: si vamos holgados en fichas en las últimas rondas, jugar más tight. La ronda se
   detecta cuando cambia el conjunto de rivales en `estado.jugadores`.

### Fase 3 · Reentrenar el genoma para el formato real (sáb 14 - lun 16)

- Nuevos genes para lo de la fase 2 (umbral de stack-off, tamaño del rango del rival, umbrales por stack).
- **Fitness** = resultado en ligas simuladas de 10 rondas (p. ej. media de log(fichas finales), con fuerte
  penalización por eliminación), no bb/100 con stacks repuestos.
- Rivales: réplicas de Darwin original, Espía, Estadístico y Novato, más maníaco, "estación" (paga todo) y roca,
  más coevolución (versiones anteriores del propio bot).
- Validar en la arena real con mesas de 5: **objetivo ≤10 % de quiebra y neto medio > Novato**.

### Sesión de dudas (dom 15) · preguntas para llevar

- ¿Se avisa de algún modo el cambio de ronda o el inicio de la final, o hay que deducirlo?
- La final, ¿es la misma conexión o una sesión nueva?
- Con el límite de 6 minutos, ¿la mano en curso se termina?
- Confirmar que se puede guardar en disco lo aprendido de los rivales (alias) entre reinicios del bot.

### Fase 4 · Modelado de rivales (lun 16 - mié 18)

- Integrar los perfiles de Espía por alias, que **persisten entre rondas** y se guardan en disco por si el bot se reinicia.
- Usar las cartas mostradas en `mano_fin.mostradas` para ajustar el rango con el que apuesta cada rival.
- Explotar: más valor y ningún farol contra quien paga todo, robar ciegas a quien se retira mucho, pagar más
  liviano contra maníacos.
- Probar en la arena real con rivales "raros" (aleatorio, un maníaco a mano) para ver que no se rompe.

### Fase 5 · Modo final y robustez (jue 19 - vie 20)

- Parámetros de la mesa final (6 jugadores, 10 000 fichas, último en pie o más fichas tras 1 h), con flag
  `-modo final` y/o detección automática.
- Pruebas de resistencia: horas conectado en silencio, corte de red a mano, matar el proceso y relanzarlo,
  mesas de 5 y de 6, stacks de 2 000 a 40 000.
- **Servidor de pruebas** de la organización apenas esté encendido (por ahora figura como no disponible).
- **Congelar el código el viernes 20.** El sábado solo se corrigen errores.

### Entrega (sáb 21)

- Código + README de una página (idea, técnicas, resultados). El jurado revisa el código.

### Día del torneo (dom 22)

- PC enchufada, sin suspensión ni actualizaciones, red por cable si se puede.
- Bot lanzado antes de la hora con el token del equipo y el bucle de reconexión. **Una sola instancia**
  (una segunda conexión con el mismo token reemplaza a la primera).
- Mirar el log en vivo, sin tocar nada.

## 5. Primer paso concreto

Copiar Darwin a `bots/<nuestro-bot>`, agregar la reconexión y el log, y armar el simulador de liga de 10 rondas
(Fases 0 y 1). Sin ese simulador no hay forma de saber si un cambio mejora el resultado en el torneo real.
