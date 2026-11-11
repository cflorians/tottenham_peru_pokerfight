# Simulaciones

Resultados de pruebas y simulaciones del equipo. Los planes citan estos archivos como evidencia.

| Archivo | Qué tiene |
| --- | --- |
| [`pruebas_iniciales.md`](pruebas_iniciales.md) | 9 torneos mano a mano entre los bots del kit (`torneo-go.json`). No es el formato de la competencia: sirve solo de referencia |
| [`pruebas_formato_real.md`](pruebas_formato_real.md) | 90 mesas de 5-6 en condiciones de una ronda de la clasificatoria. De aquí sale que Darwin original quiebra en el 44 % de las mesas y el genoma prudente en el 10 % |
| [`genoma_prudente.json`](genoma_prudente.json) | Genoma de Darwin con 4 genes ajustados a mano; es el que usa la v0 |
| [`fase0_v0.md`](fase0_v0.md) | Resultados de Lince v0: mesas de torneo, no reproducibilidad de la arena, robustez |

## Cómo agregar un resultado

- **Un archivo nuevo por experimento** (no agregar a uno existente): así dos personas nunca chocan al
  fusionar. Nombre: `<plan>_<tema>.md`, por ejemplo `a1_linea_base_liga.md` o `b2_genoma_final.md`.
- Arriba de todo: fecha, quién lo corrió, commit (`git rev-parse --short HEAD`), comando y configuración
  usados, y cantidad de mesas o ligas.
- Dar las métricas con su intervalo de confianza del 95 % (M1, M1b, M2; ver `PLAN-BOT.md`).
- Condiciones de competencia: 10 000 fichas, ciegas 50/100, 50 manos, `timeout_ms` 5000,
  `reponer_stack: false`, mesas de 5-6.
- La arena real **no es reproducible** con la misma semilla: las comparaciones pareadas solo valen en el
  simulador rápido.
- No subir historiales ni logs crudos (`resultados/` está ignorado por git): solo el resumen en markdown y
  los artefactos chicos que hagan falta (un genoma, una tabla).
- Agregar una fila a la tabla de arriba y citar el archivo desde el plan correspondiente.
