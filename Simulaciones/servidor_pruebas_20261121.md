# Práctica del 21 de noviembre: perfiles de rivales reales

- **Fecha y autor** · sáb 21/11/2026, 10:38-10:49, Carlos con Claude Code.
- **Versión** · `bin/lince.exe` compilado de `master` en `43bfb91` (Lince v1.2), copiado a `bin/lince-servidor.exe`.
- **Servidor** · `34.230.114.164:9000`, la dirección nueva que dio la organización el 21/11.
- **Comando** (desde `pokerfight-desarrollo/`, con el token en `POKERFIGHT_TOKEN` y con aviso previo a Hans):

  ```powershell
  powershell -ExecutionPolicy Bypass -File bots\lince\correr.ps1 -Addr 34.230.114.164:9000 -Fase clasificatoria -Nuevo -Silencioso -Datos <repo>\pokerfight-desarrollo\bots\lince\datos\practica-20261121 -Exe <repo>\pokerfight-desarrollo\bin\lince-servidor.exe
  ```

- **Muestra** · 853 manos en 18 mesas de 50 manos. **Todas las mesas fueron de 4 jugadores**, no de 5-6.
  Hubo 7 rivales distintos. Las cartas de los rivales solo se ven en el showdown (243 manos).
- **Análisis** · `bots/lince/pruebas/torneo_sim.py` y un script de perfiles fuera del repo, a partir de
  `manos-20261121.jsonl`. Las ciegas no vienen como
  acciones: se reconstruyeron a partir del botón. El neto por rival es aproximado.

## Lince

- **Neto: unas +63 000 fichas en 18 mesas** (+3 497 por mesa de media, error estándar 1 374; terminó 1.º en 7 de 18
  mesas y no quebró en ninguna). Una primera cuenta daba +93 912, pero el método de reparto por mano estaba
  inflado; esta sale de los saldos de cada mesa. En el servidor todos vuelven a 10 000 fichas en cada
  mesa. No hubo timeouts ni acciones inválidas, y el supervisor no tuvo que relanzar el bot.
- Lince entra en el 60 % de las manos y sube antes del flop en el 51 %.
- Después del flop apuesta casi siempre: el 64 % de las veces en el flop y el turn y el 78 % en el river.
  Lo hace con un tamaño casi fijo: mediana de 0,27 del pozo, que es el gen `tam_valor` = 0,2748.
- Ante una apuesta después del flop se retira el 48 % de las veces (66 de 138), paga el 40 % y sube el 12 %.
  Esos 66 retiros le costaron 15 434 fichas ya puestas. Cuando se vieron las cartas, la mayoría de los retiros
  eran correctos: en un primer corte de 550 manos, solo en 3 de 17 casos Lince tenía la mejor mano.

## Rivales

| Alias | Manos | Entra | Sube pre | Apuesta post (flop/turn/river) | Se retira ante apuesta | Tamaño | Neto de Lince contra él (aprox., solo para ordenar) | Tipo |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| j-e34cb4 | 553 | 40 % | 3 % | 34/27/34 % | 30 % | 0,66 | +3 806 | Pasivo antes del flop; apuesta grande con manos hechas |
| j-440f62 | 553 | 33 % | 5 % | 52/79/89 % | 36 % | 0,45 | +5 860 | Apuesta chico casi siempre después del flop, sobre todo con un par |
| j-261eec | 400 | 22 % | 6 % | 32/45/30 % | 59 % | 0,63 | +7 891 | Tight-pasivo; se le roba con apuestas chicas |
| j-6ebd85 | 300 | 85 % | 0 % | 0/0/0 % | 37 % | — | +33 954 | Paga todo y no apuesta nunca |
| j-8a2e43 | 300 | 42 % | 24 % | 71/40/20 % | 41 % | 1,49 | −1 343 | Loose-agresivo con apuestas grandes |
| j-3bb928 | 253 | 30 % | 16 % | 80/84/91 % | 81 % | 0,55 | −2 744 | Apuesta casi siempre y se retira ante resistencia |
| j-41be79 | 200 | 26 % | 5 % | 42/31/32 % | 11 % | 0,63 | +45 970 | Casi nunca se retira después del flop |

## Conclusiones

1. Lince le gana con claridad a este campo. Casi toda la ganancia sale de dos rivales que pagan demasiado
   (j-41be79 y j-6ebd85).
2. La fuga más clara no son los retiros, sino el **tamaño de la apuesta de valor**: 0,27 del pozo siempre,
   también con manos fuertes contra rivales que no se retiran. El gen `escala_valor` (hoy en 0) agranda la
   apuesta con la fuerza de la mano. **Se midió y no se activa.** En el simulador pareado (10 000 ligas por
   campo, 6 campos) baja la probabilidad de llegar a la final en todos los campos: con 0,3, entre 0,3 y 2,5
   puntos; con 0,6, entre 1,6 y 6 puntos; con 1,0, entre 4 y 13,5 puntos. Sube algo la de terminar 1.º, y el
   premio esperado sale nulo o negativo en el campo realista y en el agresivo. En la arena local (60 mesas por
   variante) gana más fichas por mesa con 0,6 y 1,0, pero dentro de unos 1,5 errores estándar. Apostar más
   grande sube la varianza, y en la clasificatoria eso cuesta clasificar.
3. Los únicos rivales contra los que pierde apuestan grande (j-8a2e43) o apuestan casi siempre y se retiran
   ante resistencia (j-3bb928). Las muestras son chicas: 250-300 manos por rival.
4. **Las mesas de práctica fueron de 4 jugadores.** No se sabe si el torneo también las armará de 4.

## Sesión larga: 62 mesas en formato de torneo (11:44-12:24)

Misma versión, mismo comando, con `-Datos ...\datos\practica-20261121b`. El objetivo era tener una muestra
parecida al torneo: varias clasificatorias de 10 rondas. **3 013 manos en 62 mesas de 4 jugadores** con 6 de los
7 rivales de la mañana. Análisis: `python bots/lince/pruebas/torneo_sim.py <manos.jsonl>`.

| Métrica | Sesión larga (62 mesas) | Todo el día (80 mesas) |
| --- | --- | --- |
| Neto medio por mesa | **+4 384** (error estándar 945) | +4 184 (error estándar 792) |
| Mediana por mesa | +2 154 | +2 373 |
| 1.º en la mesa | 35,5 % | 36,2 % |
| Puesto medio (de 4) | 1,87 | 1,93 |
| Quiebras | **1 de 62** (1,6 %) | 1 de 80 (1,2 %) |

**Como torneos de 10 rondas.** Se tomaron 6 bloques de 10 mesas seguidas: 5 terminan sin quiebra, con entre
+11 500 y +72 000 fichas; el otro quiebra en la ronda 5. Con un bootstrap de 20 000 torneos de 10 mesas al
azar, Lince **sobrevive las 10 rondas en el 85 %** (88 % con las 80 mesas del día) y, cuando sobrevive, junta
una mediana de **+44 700 fichas** (p10 +17 000, p90 +75 600). Es una aproximación: en la práctica las fichas
se reinician en cada mesa y las mesas son de 4, mientras que en el torneo se arrastran y las mesas son de 5-6.
Con una sola quiebra, el 85 % tiene mucha incertidumbre.

**La quiebra** (mesa `r350-m01`, mano 9): Lince tenía J♠ A♦ y en el turn la mesa era J♣ 4♦ 6♦ A♠, así que tenía
doble par alto. Fue all-in con 8 460 fichas, le pagó j-41be79 (el que casi nunca se retira) y el 5♣ del river
le dio la mano a él. Es mala suerte, no un error de la política.

**Robustez en condiciones reales.** A las 12:01 se cortó la conexión. Lince reconectó en el primer intento,
recuperó su silla y siguió jugando, sin que el supervisor tuviera que relanzarlo. El log no registró pánicos
ni decisiones lentas. Los timeouts no se pueden ver en el historial, que solo graba la acción aceptada.

**Rivales.** Lince le ganó a los 6 en la sesión larga, incluido j-3bb928, contra el que perdía en la mañana.
Se retiró ante una apuesta después del flop el 38 % de las veces, contra el 48 % de la mañana. Los perfiles
de los rivales se mantuvieron: j-6ebd85 paga todo sin apostar nunca, j-41be79 se retira solo el 10 %, j-3bb928
apuesta el 81 % de las veces y j-261eec se retira el 52 %.

**Conclusión.** Las 80 mesas confirman la v1.2 tal como se entregó: gana con claridad a este campo, termina 1.º
o 2.º en el 79 % de las mesas y quiebra poco. No hay ningún cambio que haga falta antes del torneo.
