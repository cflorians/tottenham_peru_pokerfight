# Lince · Tottenham_Pucp

Bot de Texas Hold'em No-Limit para **PokerFight 2026** (AICS PUCP).

| | |
| --- | --- |
| Equipo | **Tottenham_Pucp** |
| Bot | **Lince**, versión **v1.2** (etiqueta git `v1.2`; el código entregado es el de `master`) |
| Lenguaje | Go (Go 1.26.10 o más nuevo), sin dependencias fuera del módulo del kit |
| Código del bot | [`pokerfight-desarrollo/bots/lince/`](pokerfight-desarrollo/bots/lince/) |

Lince decide solo, sin intervención humana. Estima la fuerza de su mano con **Monte Carlo** y la compara contra
umbrales que aprendió un **algoritmo genético** entrenado con lo que premia el torneo. Tiene dos modos en un
solo bot: uno prudente para la clasificatoria y otro para la mesa final que **se adapta a cada rival**. Está
preparado para jugar el torneo entero sin ayuda: reconecta, guarda su estado en disco y un supervisor lo
relanza si el proceso se cae.

## Cómo se creó

1. **Punto de partida: Gato Darwin**, bot de ejemplo del kit (`bots/darwin`). Su política tiene forma fija y
   sus números salen de un genoma que evoluciona un algoritmo genético. En nuestras primeras 90 mesas en el
   formato real, Darwin quebraba en el 44 % de ellas ([pruebas](Simulaciones/pruebas_formato_real.md)).
2. **v0.** Darwin con un genoma prudente (más aversión al riesgo, menos faroles) y toda la infraestructura del
   torneo: reconexión, supervisor, estado en disco y log. La quiebra bajó al 10 %.
3. **Simulador de liga propio** (`entrenamiento/liga.go`, `cmd/simular-liga`). Juega miles de torneos
   completos: 10 rondas con fichas arrastradas, eliminación, corte de los 6 y mesa final. Incluye réplicas de
   los bots del kit. Mide tres cosas: llegar a la final (M1), terminar 1.º de la liga (M1b) y ganar la final
   (M2). [Líneas base](Simulaciones/a1_b1_linea_base.md).
4. **v1 y v1.1: entrenamiento con aptitud de liga** (`cmd/entrenar-liga`). En vez de premiar fichas por mano,
   la aptitud es el premio esperado del torneo, 60 × M1 × M2 + 30 × M1b. La v1.1 llega a la final de 10 a 13
   puntos más seguido que la v1 en los cuatro campos simulados. En la arena real ganó el 89 % de 80 finales
   contra los bots del kit ([informe](Simulaciones/a2_b2_v1.md)).
5. **v1.2: adaptación al rival y genoma de la final.** Un genoma fijo no puede creerle a quien apuesta poco y
   pagarle a quien apuesta siempre. Se agregaron dos genes que usan cuán seguido apuesta cada rival, y se
   entrenó un genoma solo para la final, con la mitad de las finales contra genomas nuestros ya entrenados.
   Contra esos rivales fuertes gana 10 de 20 finales reales; la v1.1 ganaba 1 de 20
   ([informe](Simulaciones/b4_adaptacion_rival.md)).

Cada versión se promovió después de medirla contra la anterior en el simulador y en mesas reales del kit.

## Cómo juega

El código de la decisión está en `bots/lince/estrategia/` (`politica.go`, `fuerza.go`, `lectura.go`).

- **Fuerza de la mano.** Es la probabilidad de ganarle a *una* mano al azar. Antes del flop sale de una tabla de
  las 169 manos iniciales, precalculada con 4 000 simulaciones cada una e incluida en el binario. Después del
  flop se simulan hasta 1 500 repartos del resto del mazo. El cálculo usa el mismo evaluador que la mesa
  (`pkg/poker`) y se corta a los 400 ms, lejos del plazo de 5 s.
- **Fuerza efectiva.** Corrige esa probabilidad por la mesa:
  `fuerza^(1 + (rivales − 1) × multijugador) + bono_posición × (2 × posición − 1) − presión × subidas`.
  Este número se compara con los umbrales del genoma, que son distintos antes y después del flop.
- **Si nadie apostó.** Apuesta por valor si la fuerza efectiva supera el umbral de apuesta. Abre a unas 2 ciegas
  grandes antes del flop; después apuesta una fracción del pozo (0,27 en la clasificatoria). Si no, puede
  farolear: la probabilidad crece cuando habla último.
- **Ante una apuesta.** Sube si supera el umbral de subida. Paga si su fuerza supera las *pot odds*,
  `por_pagar / (pozo + por_pagar)`, multiplicadas por un margen. Si no, se retira. Nunca se retira cuando
  pasar es gratis.
- **Aversión al riesgo.** Cada umbral sube en proporción a la parte del stack que la jugada compromete.
  Jugarse todo exige más mano que jugarse una ciega, porque quebrar en la clasificatoria es quedar eliminado.
- **All-in.** Va all-in de entrada solo si la fuerza efectiva supera `umbral_allin + aversión_riesgo` (0,93 en
  la clasificatoria). Con el genoma de la final ese valor pasa de 1, así que allí solo llega al all-in subiendo.
  La política también tiene una regla de stack corto (empujar o retirarse); ningún genoma en uso la activa.
- **Lectura del rival.** Cuenta, por alias, las apuestas y subidas frente a los pagos de cada rival. Suma 8
  acciones imaginarias de un rival neutro para no creerle a una muestra chica. Con el rival que hizo la última
  apuesta de la calle, el genoma de la final baja su umbral para resubirle y abarata pagarle si apuesta muy
  seguido. Si apuesta poco, hace lo contrario. Los perfiles duran todo el torneo (`datos/rivales.json`).
- **Detección de fase.** La arena no avisa el cambio de ronda (`torneo/torneo.go`). Lince toma una mesa como
  nueva si cambia el prefijo del id de mano, si aparece un jugador nuevo o si la numeración vuelve atrás. La
  final es una mesa nueva de 6 jugadores, todos con 10 000 fichas, después de haber jugado al menos una mesa.
  Entonces cambia de `genoma.json` a `genoma_final.json`. Ante la duda juega como en la clasificatoria.

Los dos genomas están incluidos en el binario.

## Robustez

- **Reconexión** con el mismo token si se corta la conexión: espera de 1 a 5 s, sin límite de intentos.
- **Decisión protegida.** Un error al decidir juega la acción segura (pasar o retirarse) y queda en el log.
- **Estado en disco.** Al cerrar cada mano guarda la ronda, la fase, las fichas y las cartas de la mano en
  curso. Si vuelve a mitad de una mano, la sigue jugando.
- **Supervisor** (`correr.ps1`). Relanza el proceso si muere y no deja correr dos instancias con el mismo token.
- **Silencio entre rondas.** Pasó 38 de 38 comprobaciones tras 4 horas conectado sin jugar, con reinicios en
  la clasificatoria y en la final ([informe](Simulaciones/ensayo_silencio_cuatro_horas.md)).

## Resultados de validación

| Prueba | Resultado | Fuente |
| --- | --- | --- |
| Servidor de pruebas de la organización, sáb 21/11 | **+93 912 fichas en 853 manos** (+110 por mano), contra 7 rivales reales, sin timeouts ni acciones inválidas | [servidor_pruebas_20261121.md](Simulaciones/servidor_pruebas_20261121.md) |
| 15 torneos completos de 30 bots con los ejecutables reales del kit | Llegó a la final 12 veces y fue campeón 10. Contó bien las rondas y detectó la final en los 15 | [ensayo_liga_real.md](Simulaciones/ensayo_liga_real.md) |
| Finales reales contra genomas nuestros entrenados (v1.2) | Ganó 10 de 20; la v1.1, 1 de 20 | [b4_adaptacion_rival.md](Simulaciones/b4_adaptacion_rival.md) |
| Finales reales contra los bots del kit (v1.2) | 82,5 % de 40 | [b4_adaptacion_rival.md](Simulaciones/b4_adaptacion_rival.md) |
| Rondas sueltas de clasificatoria (v1.1, 210 mesas reales) | Neto medio +6 244, quiebra 6 % | [a2_b2_v1.md](Simulaciones/a2_b2_v1.md) |

En el servidor de pruebas todos vuelven a 10 000 fichas en cada mesa, así que el neto de la práctica no se
arrastra como en la clasificatoria.

## Estructura

```text
README.md                        este archivo
Simulaciones/                    informes de las pruebas citadas arriba
pokerfight-desarrollo/           módulo Go: copia del kit de la organización, sin cambios, más nuestro bot
  go.mod
  pkg/, internal/, cmd/, docs/   kit (SDK, evaluador, mesa y arena locales)
  bots/<ejemplos>/               bots de ejemplo del kit (rivales de las pruebas)
  bots/lince/                    NUESTRO BOT
    main.go                      conexión, reconexión, decisión protegida y log
    rivales.go, registro.go      lectura de rivales y registro de manos
    estrategia/                  fuerza por Monte Carlo, política y genoma
    torneo/                      ronda, fase y estado en disco
    genoma.json                  genoma de la clasificatoria (v1.1)
    genoma_final.json            genoma de la mesa final (v1.2)
    preflop.json                 tabla de fuerza de las manos iniciales
    entrenamiento/, cmd/         simulador de liga, entrenadores y ensayo del torneo
    rivales/, genoma_v0.json, genoma_darwin.json   genomas de referencia del simulador
    pruebas/                     mesas reales, perfil de rivales y prueba de robustez (Python 3)
    correr.ps1                   supervisor para Windows
    README.md                    guía técnica
  pruebas-lince/mesa5.json       mesa de 5 en condiciones de torneo
```

El kit se incluye porque el bot importa su SDK (`pkg/botsdk`, `pkg/poker`, `internal/protocolo`). Además,
las herramientas de prueba levantan su mesa y sus bots de ejemplo.

## Compilar y ejecutar

Desde `pokerfight-desarrollo/`:

```bash
go build -o bin/ ./bots/lince
go test -trimpath ./bots/lince/...
```

Para compilar también la arena y los bots de ejemplo, usar `go build -o bin/ ./cmd/... ./bots/...`. En
Windows hay que escribir `-o bin/` con la barra final para que Go agregue el `.exe`.

**Torneo.** El token del equipo se lee de la variable de entorno `POKERFIGHT_TOKEN`. Así no aparece en la línea
de comandos. La dirección `HOST:PUERTO` es la que anuncie la organización.

```powershell
powershell -ExecutionPolicy Bypass -File bots\lince\correr.ps1 -Addr HOST:PUERTO -Nuevo -Silencioso
```

- `-Nuevo` se usa una sola vez, al empezar el torneo. Si hay que relanzarlo a mitad, va sin `-Nuevo` para
  retomar la ronda guardada.
- En el servidor de pruebas se agrega `-Fase clasificatoria`, porque allí todos vuelven a 10 000 fichas en cada
  ronda.
- Sin PowerShell, sin supervisor: `./bin/lince -addr HOST:PUERTO -datos bots/lince/datos -silencioso`.

**Prueba local.** Una mesa de 5 en condiciones de torneo contra Darwin, Espía, Estadístico y Base:

```bash
go build -o bin/ ./cmd/... ./bots/...
./bin/arena correr -config pruebas-lince/mesa5.json
```

La guía técnica completa (simulador, entrenamiento, ensayos y todos los flags) está en el
[README de Lince](pokerfight-desarrollo/bots/lince/README.md).

## Créditos

- El equipo **Tottenham_Pucp** diseñó, entrenó y validó Lince.
- El código se escribió con apoyo de **Claude Code** (simulador de liga, entrenamiento, adaptación al rival,
  análisis del servidor de pruebas) y **Codex** (registro de fase y pruebas de reinicio, reloj y silencio).
- Kit de inicio, arena y bot base Gato Darwin: AICS PUCP.
