# Cierre por reloj y arrastre de fichas en liga local

- **Fecha y autor** · 18/11/2026, Hans con Codex.
- **Código probado** · `63906a2`, rama `hans/pokerfight`.
- **Configuración** · dos clasificatorias de tres rondas, diez bots por liga, semillas 701 y 702. Cada bot comienza con 10 000 fichas. Ciegas 50/100, hasta 50 manos por mesa, plazo de 5 000 ms por jugada y fichas arrastradas entre rondas.
- **Reloj de mesa** · cinco segundos para provocar el cierre por tiempo en la arena local.

Comando ejecutado desde `pokerfight-desarrollo` en PowerShell. Para repetir el ensayo, elige una carpeta de salida nueva.

```powershell
.\bin\liga-real.exe -n 10 -rondas 3 -manos 50 -semilla 701 -ligas 2 `
    -tipos darwin,espia,estadistico,base,conservador -paralelo 2 `
    -reloj-mesa 5s -sin-final -espera-bots 12s -espera-lince 15s `
    -salida bin/ensayos-reloj/liga-5s-701
```

## Resultados

| Semilla | Mesas cerradas por reloj | Fichas de Lince al cerrar R1 y empezar R2 | Fichas de Lince al cerrar R2 y empezar R3 | Timeouts de Lince |
| ---: | ---: | ---: | ---: | ---: |
| 701 | 3 de 4 | 11 701 | 9 282 | 0 |
| 702 | 3 de 4 | 15 977 | 28 642 | 0 |

- Seis mesas cerraron por reloj tras liquidar la mano en curso, con entre 34 y 48 manos jugadas. Las otras dos completaron 50 manos.
- Las ocho mesas conservaron sus fichas y cada liga mantuvo las 100 000 fichas en las tres rondas.
- Coincidieron los 18 saldos que se arrastraron de una ronda a otra. Cuatro correspondieron a Lince, tres de ellos tras un cierre por reloj. Sus saldos de inicio en `estado.json` coincidieron con los de la mesa.

## Estado al apagar el ensayo

En la semilla 701, la mesa cerró la tercera ronda en la mano 34 con 26 188 fichas para Lince. El historial de Lince incluyó esa mano y `estado.json.tmp` contenía ese saldo. El `estado.json` definitivo quedó en la mano 33 con 26 389 fichas. La evidencia apunta a que `liga-real` detuvo el proceso entre la escritura del archivo temporal y su renombre. El resultado de la mesa y los saldos entregados en las transiciones coincidieron con el informe de la liga.

El ensayo cubrió el cierre entre manos del orquestador local con reloj reducido. Sigue pendiente la prueba con el reloj de seis minutos.
