# Práctica del 20 de noviembre con el candidato de entrega

- **Fecha y autor** · 20/11/2026, Hans con Codex.
- **Versión probada** · Ejecutable candidato de `6a5ac83` (SHA-256 `9bb6bb0b552bdab7336b799b9ee4d874142a523ef6bdaa718a0b5b008d6d8bf2`), en Windows x64.
- **Configuración** · Una conexión al servidor de práctica, con aviso al compañero y fase clasificatoria forzada. Se usó una carpeta de datos separada y se detuvo el proceso a los 90 segundos.

Desde `pokerfight-desarrollo/`, con el token del equipo en la variable de entorno y la carpeta de salida recién creada:

```powershell
& .\bin\practica-20261120-134832\lince-practica.exe `
    -addr ec2-34-230-114-164.compute-1.amazonaws.com:9000 `
    -fase clasificatoria -una-partida -silencioso -datos .\bin\practica-20261120-134832
```

## Resultados

- El servidor aceptó la conexión. Lince registró 179 manos en cuatro mesas, con 50, 50, 50 y 29 manos respectivamente.
- El estado guardado quedó en la ronda 4, en fase clasificatoria. El log no registró pánicos, errores ni avisos de decisión lenta.
- Se verificó que el proceso terminó. Los historiales y logs se conservaron en la carpeta local ignorada por Git.

La sesión verifica la conexión y el juego en el servidor de práctica de esa fecha. Quedan por confirmar la dirección y el horario del torneo, y por evaluar los perfiles del campo con una muestra más amplia.
