# Servidor de pruebas · primera conexión

- **Fecha**: sáb 14/11/2026, 13:02 · **Quién**: Carlos (con Claude Code) · **Commit**: rama `a/v1-fitness-liga` sobre `30ac88c`
- **Bot**: Lince v0 (`bin/lince.exe`, genoma prudente), con el supervisor.
- **Comando** (desde `pokerfight-desarrollo/`, con el token en la variable `POKERFIGHT_TOKEN`):
  `powershell -ExecutionPolicy Bypass -File bots\lince\correr.ps1 -Addr 32.201.29.142:9000 -Fase clasificatoria -Nuevo -Silencioso`
- **Muestra**: una sola mesa, 30 manos. Es muy poco: sirve para ver el formato, no para medir el juego.

Cómo conectarse (y cómo guardar el token) está en `AGENTS.md`, sección "Token y servidor de la organización".

## Qué se vio

| Tema | Resultado |
| --- | --- |
| Token | **Válido**: el servidor nos aceptó y nos dio el alias `j-c79d9f` |
| Disponibilidad | **A ratos.** Vie 13/11: sin respuesta desde dos redes. Sáb 14/11 12:57: abierto. 13:02:34 conectamos; 13:02:49 el servidor cerró la conexión y desde entonces no responde (reintentos fallidos a las 13:03 y 13:04) |
| Mesa | 6 jugadores, todos con 10 000 fichas: como dice la guía |
| Forma del `id_mano` | `practica-r1-m01-mano-30`: prefijo de mesa `practica-r1-m01` (ronda y mesa) y número de mano al final. **El bot lo leyó bien**: detectó la mesa nueva y contó las manos (responde la pregunta 13 del plan, al menos para el servidor de pruebas) |
| Ritmo | **30 manos en 10 segundos: 3 manos por segundo.** Una ronda de 50 manos dura ~17 s; el límite de 6 minutos no aprieta |
| Tiempo de respuesta | Nuestras decisiones salen cada 0,1-0,3 s, sin ningún timeout |
| Resultado | Stack de 10 000 a 12 277 en 30 manos (no dice nada: es una mesa) |
| Reconexión | Al cerrarse la conexión el bot reintentó solo, con esperas crecientes, sin caerse |

## Qué cambia en el plan

1. **La final se va a decidir por eliminación, no por reloj.** A 3 manos por segundo, una hora son ~10 000
   manos: con ciegas fijas de 50/100 y 60 000 fichas en la mesa, alguien queda solo mucho antes. De las tres
   duraciones del plan B (150, 500 y sin tope), la que importa es **sin tope**. El "tramo final por reloj"
   (B3) pierde prioridad. A confirmar cuando el servidor esté estable: el ritmo puede bajar con bots lentos.
2. **La detección de ronda por `id_mano` funciona** con el formato real. Falta ver el cambio de una ronda a
   la siguiente (el servidor cortó antes) y si en el torneo el prefijo es parecido.
3. **El servidor cierra la conexión**, o se cae: no se puede contar con que "una conexión dura todo el
   torneo". La reconexión con el mismo token es imprescindible, y ya funciona.

## Pendiente

- Dejar el bot conectado un rato largo cuando el servidor esté estable: ver varias rondas seguidas, el cambio
  de mesa, la latencia y cómo juegan los rivales de verdad (para recalibrar los campos del simulador).
- Avisar a la organización de las caídas, con la hora (13:02:49 del 14/11) y nuestro alias (`j-c79d9f`).
