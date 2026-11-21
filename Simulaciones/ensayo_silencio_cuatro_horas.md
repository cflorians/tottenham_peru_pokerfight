# Cuatro horas de silencio y reanudación del juego

- **Fecha y autor** · 20/11/2026, Hans con Codex.
- **Versión probada** · Ejecutable candidato de `6a5ac83` (SHA-256 `9bb6bb0b552bdab7336b799b9ee4d874142a523ef6bdaa718a0b5b008d6d8bf2`), en Windows x64.
- **Configuración** · Supervisor `correr.ps1` y arena local con puerto 9117. El bot espera conectado 14 400 s antes de jugar dos mesas clasificatorias de cinco jugadores y una final de seis.

Desde `pokerfight-desarrollo/`, con el candidato como `bin/lince.exe` en la copia de prueba:

```powershell
python -B bots/lince/pruebas/robustez.py -silencio 14400
```

## Resultados

- **38 de 38 comprobaciones aprobadas**, con código de salida 0. Hasta iniciar la primera mesa, el log muestra un único arranque, una conexión aceptada y cero desconexiones o reintentos.
- Después de la espera, Lince jugó la primera mesa. Reconectó para la segunda y conservó el asiento, la ronda y la hora de inicio al relanzarlo en clasificatoria. La mesa aceptó una subida tras el reinicio.
- El supervisor también relanzó a Lince durante la final. Conservó el asiento, la fase y el inicio de la final, y la mesa aceptó un pago después de reconectar.
- El log quedó con cero pánicos, errores y avisos de decisión lenta. El puerto local quedó cerrado al terminar.

El reloj local muestra la conexión a las 13:42 y la primera mesa a las 22:53. La duración indicada corresponde al temporizador de 14 400 s del verificador. El registro de Lince mantiene el arranque y la conexión originales hasta la primera mesa.
