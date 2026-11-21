# Frecuencia del pago corto mano a mano

- **Fecha y autor** · 19/11/2026, Hans con Codex.
- **Código probado** · `f0ae920`, política vigente con observación adicional en una copia del simulador.
- **Configuración** · 2 000 ligas por campo, 30 participantes por liga, semilla 20261119, 2 048 repartos, 128 muestras por calle y tope efectivo de 5 000 manos en la final.

Se contaron los turnos postflop en los que dos jugadores recibieron cartas, ambos seguían activos y a Lince le faltaban fichas para cubrir la apuesta. El diagnóstico calculó si el precio efectivo habría cambiado una retirada por un pago corto. Las ocho comparaciones de control coincidieron con el simulador base y se jugaron las decisiones vigentes.

Desde el módulo de la copia local `pokerfight-desarrollo/bin/diagnostico-pago-corto-f0ae920/pokerfight-desarrollo`, con la biblioteca de evaluación que quedó en esa carpeta de diagnóstico.

```powershell
$env:GOMAXPROCS = '2'
..\diagnostico-pago-corto.exe -n 30 -ligas 2000 -semilla 20261119 -repartos 2048 `
    -muestras 128 -tope-final 0 -campos 'kit,agresivo,mixto,fuerte' -cache ..\biblioteca-pago-corto.gob
```

| Campo | Clasificatorias con un posible cambio | Finales con un posible cambio |
| --- | ---: | ---: |
| Kit | 7/2 000 (0,35 %) | 8/1 686 (0,47 %) |
| Agresivo | 134/2 000 (6,70 %) | 82/1 609 (5,10 %) |
| Mixto | 6/2 000 (0,30 %) | 4/1 612 (0,25 %) |
| Fuerte | 0/2 000 (0 %) | 7/534 (1,31 %) |

El campo agresivo concentró 258 de los 294 cambios potenciales. Se midió la frecuencia del caso. El efecto en clasificación, campeonatos y premio esperado requiere una comparación con la variante aplicada. La política vigente sigue siendo la candidata de entrega.
