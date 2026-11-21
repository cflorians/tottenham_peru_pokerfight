# Cierre local con reloj de seis minutos

- **Fecha y autor** · 19/11/2026, Hans con Codex.
- **Versión probada** · fuente de `master` en `f0ae920` y ejecutable candidato de `6a5ac83` (SHA-256 `9bb6bb0b552bdab7336b799b9ee4d874142a523ef6bdaa718a0b5b008d6d8bf2`).
- **Configuración** · dos rondas, cinco jugadores, hasta 50 manos, ciegas 50/100 y cinco segundos por jugada. Cuatro rivales locales controlados retrasaron sus respuestas durante la primera ronda para alcanzar el reloj. En la segunda respondieron de inmediato.

El rival de prueba quedó en `bots/novato/bot.py` dentro de la copia local del ensayo. Para repetir desde la raíz del repositorio, usar una carpeta de salida nueva.

```powershell
$ligaReal = (Resolve-Path .\pokerfight-desarrollo\bin\liga-real.exe).Path
$carpetaCandidato = (Resolve-Path .\pokerfight-desarrollo\bin\entrega\6a5ac83).Path
$python = (Get-Command python).Source
Push-Location .\pokerfight-desarrollo\bin\ensayos-reloj\reloj-6min-6a5ac83-20261119-001644
& $ligaReal -n 5 -tipos novato -rondas 2 -manos 50 -semilla 701 -paralelo 1 `
    -reloj-mesa 6m -timeout-ms 5000 -sin-final -espera-bots 12s -espera-lince 15s `
    -puerto-lince 19510 -bin $carpetaCandidato -python $python -salida resultados-repeticion
Pop-Location
```

## Resultados

| Ronda | Manos | Duración de la mesa | Saldo final de Lince | Fichas totales |
| ---: | ---: | ---: | ---: | ---: |
| 1 | 6 | 372,014 s | 15 248 | 50 000 |
| 2 | 50 | 7,827 s | 11 250 | 50 000 |

El reloj se activó en la primera ronda y la mesa liquidó la mano en curso. Los cinco saldos coincidieron entre el cierre y el inicio de la ronda siguiente. Al acabar, el resultado, el historial y `estado.json` de Lince coincidieron en la última mano y su saldo. Hubo cero timeouts, pánicos y errores. El desfase de guardado apareció en otros ensayos, incluido el [corte de conexión](ensayo_corte_conexion_liga.md).
