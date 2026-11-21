# Corte de conexión local durante una liga completa

- **Fecha y autor** · 19/11/2026, Hans con Codex.
- **Versión probada** · ejecutable candidato de `6a5ac83` (SHA-256 `9bb6bb0b552bdab7336b799b9ee4d874142a523ef6bdaa718a0b5b008d6d8bf2`).
- **Configuración** · 30 bots, diez rondas de hasta 50 manos y final de seis jugadores. Un intermediario TCP local cortó solo la conexión de Lince durante la primera ronda. El proceso de Lince siguió vivo y usó su reconexión propia.

El arnés del corte y el verificador `resumir_corte.py` quedaron en la copia local `pokerfight-desarrollo/bin/ensayo-corte-red-6a5ac83/pokerfight-desarrollo`. Comando desde esa copia con una carpeta de salida nueva para cada repetición.

```powershell
.\bin\liga-real.exe -n 30 -rondas 10 -manos 50 -semilla 500 -ligas 1 -paralelo 2 `
    -puerto-lince 19520 -reloj-mesa 6m -timeout-ms 5000 -tope-final 2000 -salida resultados-corte/semilla-500
```

## Resultados

- El corte ocurrió en la mano 5. Lince volvió con el mismo identificador y en la misma silla en 2,005 s. La mesa confirmó una apuesta de Lince en la mano 14.
- Se completaron las diez rondas y la final. Las clasificatorias conservaron las 300 000 fichas y los saldos de arrastre coincidieron. Lince terminó segundo de la clasificatoria con 71 708 fichas. En la final de este ensayo ganó tras 424 manos.
- Lince registró 12 timeouts en la ronda del corte y cero desde la ronda 2 hasta el final de la liga. El log quedó con cero pánicos y errores.
- El estado conservó la fase final y la ronda 11. El historial y la mesa llegaron a la mano 424, con 60 000 fichas para Lince. Al apagar el ensayo, `estado.json` quedó en la mano 423 con 59 992 fichas. El chequeo del orquestador compara los saldos de inicio y pasa aunque el último guardado quede rezagado.

El corte fue en loopback y la liga usó el ejecutable directamente. Para comprobar el supervisor durante una liga completa se requiere un ensayo con `correr.ps1`.
