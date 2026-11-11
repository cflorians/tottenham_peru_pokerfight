# Fase 0 · Lince v0 (mar 10/11/2026)

Lince v0 = Gato Darwin con el genoma prudente + infraestructura (reconexión, `recover`, supervisor,
estado en disco, log). Código en `pokerfight-desarrollo/bots/lince/`, etiqueta git `v0`.

## Juega igual que Darwin prudente

`diff` de `bots/darwin/estrategia` contra `bots/lince/estrategia`: solo cambian las rutas de los imports.
La política es la misma; lo único distinto es el envoltorio.

## Mesas en condiciones de torneo

Arena `mesa-unica`: 5 jugadores (lince, darwin, espia, estadistico, base), 10 000 fichas, ciegas 50/100,
50 manos, plazo 5 s, sin reponer fichas.

| Bot en la silla "lince" | Mesas | Neto medio | Quiebras | Timeouts |
| --- | ---: | ---: | ---: | ---: |
| `bin/lince` (v0) | 30 | +2 336 | 5 (17 %) | 0 |
| `bin/darwin -genoma genoma_prudente.json` | 30 | +4 522 | 3 (10 %) | no medido |
| `bin/lince` (v0), binario final | 15 | no medido | 2 (13 %) | 0 |

Con 30 mesas esas diferencias son ruido: el código de decisión es el mismo, y el intervalo del 95 % de
"5 de 30" va de 7 % a 34 %. Lo que la prueba sí muestra: **45 mesas sin un solo timeout**.

## Las mesas de la arena no son reproducibles

Misma configuración, misma semilla y mismos bots, dos corridas:

| Semilla | Corrida 1 (neto de Lince) | Corrida 2 |
| --- | ---: | ---: |
| 1 | −102 | −9 141 |
| 2 | +6 875 | +6 811 |
| 3 | +16 067 | +6 742 |

Las cartas son las mismas, pero algún bot del kit decide con azar sin sembrar. Consecuencia: en la arena
real no se pueden comparar dos versiones "con las mismas cartas"; hacen falta más mesas. La comparación
pareada solo existe en el simulador rápido del entrenamiento.

## Robustez (`bots/lince/pruebas/robustez.py`)

Contra `bin/mesa.exe` real, con el supervisor. 21 de 21 comprobaciones:

1. Conecta y aguanta en silencio sin cerrar la conexión (probado con 40 s y con 10 minutos).
2. Juega una mesa de 5 de 50 manos; la anota como ronda 1, clasificatoria.
3. Cuando la mesa cierra, reconecta solo a la siguiente (esperas de 1, 2, 3 s) y la anota como ronda 2.
4. En una mesa de 6 con 10 000 fichas cada uno detecta la final y guarda la hora de inicio.
5. Con el proceso matado en la mano 27 de la final: el supervisor lo relanza en 1 s, el bot retoma el estado
   de disco, vuelve a su silla y sigue en ronda 3, fase final, con la misma hora de inicio y sin contarla como
   mesa nueva.

Sin pánicos, errores ni decisiones de más de 1 s en el log.

**Un error encontrado y corregido.** En la corrida de 10 minutos falló una vez el guardado del estado
("Access is denied"): en Windows no se puede reemplazar un archivo que otro proceso tiene abierto, y la
propia prueba lo estaba leyendo cada 20 ms. El bot siguió jugando y guardó bien en la mano siguiente, pero
el día del torneo lo mismo puede pasar con un antivirus o con alguien mirando el archivo. Ahora el guardado
reintenta hasta 5 veces (50 ms en total). Con la corrección: tres corridas seguidas, 21 de 21 y cero errores.

## Smart App Control

Bloqueó `lince.test.exe` (el ejecutable de `go test` del paquete principal) y lo dejó pasar al recompilar
con `-trimpath` o `-ldflags=-s`; después de editar un test, volvió a bloquear una variante que antes pasaba.
El veredicto depende del binario exacto. `bin/lince.exe` arrancó en las tres compilaciones de hoy.

## No probado todavía

- Silencio de varias horas.
- Corte de red real (se probó el cierre de la mesa y la muerte del proceso, no un cable desconectado).
- Servidor de la organización: token, latencia, forma de los `id_mano`.
