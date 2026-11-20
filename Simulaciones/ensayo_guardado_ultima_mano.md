# Guardado de la última mano al cerrar una liga local

- **Fecha y autor** · 20/11/2026, Hans con Codex.
- **Versión probada** · Lince candidato de `6a5ac83` (SHA-256 `9bb6bb0b552bdab7336b799b9ee4d874142a523ef6bdaa718a0b5b008d6d8bf2`), con `liga-real` compilado desde `f0ae920` más el ajuste de guardado.
- **Configuración** · Tres ligas locales de 12 bots con tres rondas y reloj de 5 s, más una liga de 30 bots con diez rondas y reloj de 6 min. Límite de 50 manos por mesa clasificatoria y 5 000 ms por decisión. Los rivales son los ejecutables del kit.

Desde `pokerfight-desarrollo/`, con la herramienta actualizada y el candidato de entrega como `bin/lince.exe` en la copia de prueba:

```powershell
.\bin\liga-real.exe -n 12 -rondas 3 -manos 50 -semilla 701 -ligas 3 -paralelo 2 `
    -puerto-lince 19525 -reloj-mesa 5s -timeout-ms 5000 -tope-final 120 `
    -tipos darwin,espia,estadistico,base,conservador,aleatorio -salida bin/validacion-guardado-reloj-tcp
.\bin\liga-real.exe -n 30 -rondas 10 -manos 50 -semilla 500 -ligas 1 -paralelo 2 `
    -puerto-lince 19526 -reloj-mesa 6m -timeout-ms 5000 -tope-final 500 `
    -tipos darwin,espia,estadistico,base,conservador,aleatorio -salida bin/validacion-guardado-liga-tcp
```

## Resultados

| Semilla | Cierres por reloj | Última mano de Lince en la mesa y en `estado.json` |
| ---: | ---: | --- |
| 701 | 0 | `L01-final-mano-109` |
| 702 | 3 | `L02-final-mano-120` |
| 703 | 1 | `L03-final-mano-120` |
| 500 | 0 | `L01-final-mano-500` |

- En las cuatro ligas coincidieron el último ID de mano liquidada de Lince y el del archivo definitivo. Se verificaron los archivos directamente después de detener los procesos. No quedaron archivos de estado temporales.
- En la semilla 703, Lince quedó eliminado en la segunda ronda y pasó a la final por el criterio de últimos eliminados. La herramienta esperó el guardado al detenerlo y al cerrar la final.
- En la semilla 500, Lince quedó como único superviviente de la clasificatoria en la quinta ronda. La liga conservó las 300 000 fichas hasta cerrar la clasificatoria y jugó la final.
- Las cuatro ligas registraron cero timeouts de Lince, pánicos y errores en su log. Pasaron la compilación, `go vet` y `go test -trimpath -count=1 ./bots/lince/...`. Las pruebas de la herramienta verifican que el archivo temporal no satisface el chequeo y que la espera se interrumpe si termina el proceso o se cancela el ensayo.

El ajuste se limita a la herramienta de pruebas. Se usó el mismo ejecutable de Lince que en la preparación de la entrega.
