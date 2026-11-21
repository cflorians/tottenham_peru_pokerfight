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
- **Análisis** · script propio fuera del repo, a partir de `manos-20261121.jsonl`. Las ciegas no vienen como
  acciones: se reconstruyeron a partir del botón. El neto por rival es aproximado.

## Lince

- **Neto: +93 912 fichas en 853 manos** (+110 por mano). En el servidor todos vuelven a 10 000 fichas en cada
  mesa. No hubo timeouts ni acciones inválidas, y el supervisor no tuvo que relanzar el bot.
- Lince entra en el 60 % de las manos y sube antes del flop en el 51 %.
- Después del flop apuesta casi siempre: el 64 % de las veces en el flop y el turn y el 78 % en el river.
  Lo hace con un tamaño casi fijo: mediana de 0,27 del pozo, que es el gen `tam_valor` = 0,2748.
- Ante una apuesta después del flop se retira el 48 % de las veces (66 de 138), paga el 40 % y sube el 12 %.
  Esos 66 retiros le costaron 15 434 fichas ya puestas. Cuando se vieron las cartas, la mayoría de los retiros
  eran correctos: en un primer corte de 550 manos, solo en 3 de 17 casos Lince tenía la mejor mano.

## Rivales

| Alias | Manos | Entra | Sube pre | Apuesta post (flop/turn/river) | Se retira ante apuesta | Tamaño | Neto de Lince contra él | Tipo |
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
