# Simulaciones

Resultados de pruebas y simulaciones del equipo. Los planes citan estos archivos como evidencia.

| Archivo | Qué tiene |
| --- | --- |
| [`pruebas_iniciales.md`](pruebas_iniciales.md) | 9 torneos mano a mano entre los bots del kit (`torneo-go.json`). No es el formato de la competencia: sirve solo de referencia |
| [`pruebas_formato_real.md`](pruebas_formato_real.md) | 90 mesas de 5-6 en condiciones de una ronda de la clasificatoria. De aquí sale que Darwin original quiebra en el 44 % de las mesas y el genoma prudente en el 10 % |
| [`genoma_prudente.json`](genoma_prudente.json) | Genoma de Darwin con 4 genes ajustados a mano; es el que usa la v0 |
| [`fase0_v0.md`](fase0_v0.md) | Resultados de Lince v0: mesas de torneo, no reproducibilidad de la arena, robustez |
| [`a1_b1_linea_base.md`](a1_b1_linea_base.md) | Línea base en el simulador de liga y de mesa final: M1, M1b y M2 de la v0, Darwin original y Novato en 4 campos y 4 tamaños, curva de vivos por ronda, quién llega a la final y tablas de valor. Novato supera a la v0 |
| [`a2_b2_v1.md`](a2_b2_v1.md) | **Lince v1 y v1.1**: genoma entrenado con el premio del torneo. Pruebas de promoción en el simulador, cientos de mesas y finales reales por versión (la v1.1 gana el 89 % de 80 finales), réplicas de Espía y Estadístico y paridad del simulador |
| [`genoma_final_b2.json`](genoma_final_b2.json) | Genoma entrenado solo para ganar finales (B2). No supera al de la v1 en la final: no se usa |
| [`servidor_pruebas.md`](servidor_pruebas.md) | Primera conexión al servidor de pruebas: token válido, forma de los `id_mano`, 3 manos por segundo, servidor inestable |
| [`a3_campo_fuerte.md`](a3_campo_fuerte.md) | Qué pasa si otros equipos entrenan como nosotros: la v1.1 gana la final el 5-13 % en un campo de bots entrenados. No hay un genoma mejor que todos. Diagnóstico de quiebras, genomas candidatos y el selector adaptativo, descartado. El bot no cambia |
| [`genoma_final_f2.json`](genoma_final_f2.json) | Genoma de final entrenado contra finalistas fuertes: les gana el 54 %, pero solo el 8 % a los del kit. No se usa |
| [`genoma_final_eq.json`](genoma_final_eq.json) | Genoma de final entrenado con objetivo equilibrado: parejo (59 % contra mixto, 63 % contra fuertes) pero mediocre. No se usa |
| [`b4_adaptacion_rival.md`](b4_adaptacion_rival.md) | **Lince v1.2**: el bot lee cuán seguido apuesta cada rival y la final se juega con un genoma adaptativo. En 20 finales reales contra bots entrenados gana 10 (la v1.1, 1); contra el kit, 82,5 % de 40 (v1.1: 92,5 %) |
| `genoma_final_ad*.json`, `genoma_clasificatoria_ad.json` | Genomas adaptativos de `b4`: el promovido (`ad25`) y los que no |
| [`analisis_como_jugar.md`](analisis_como_jugar.md) | **Cómo debería jugar el bot**: qué pide el formato, cómo juegan los mejores programas y profesionales (con fuentes), y cómo juega Lince medido en 1 610 manos reales. Lince entra en el doble de manos que un profesional y nunca paga después del flop. Empujar con stack corto y escalar la apuesta de valor, probados: casi no aportan. El bot no cambia |
| [`genoma_clasificatoria_g.json`](genoma_clasificatoria_g.json) | Clasificatoria reentrenada con los genes nuevos libres: llega algo más a la final y termina 1.º bastante menos. No se usa |

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
