# Lince · Tottenham Perú

Lince es un bot autónomo de Texas Hold’em No-Limit para PokerFight 2026. Parte de **Gato Darwin**, un bot
de ejemplo del kit de AICS PUCP. Estima la fuerza de su mano con Monte Carlo a partir de sus cartas y las
comunitarias. Usa el estado público de la mesa para decidir entre las acciones válidas de cada turno.

## Estrategia

- **Clasificatoria.** Los parámetros de decisión se entrenaron en ligas simuladas de diez rondas con fichas
  arrastradas. El entrenamiento valora tanto llegar a la final como terminar primero en la liga.
- **Final.** Usa un genoma específico para esa mesa. Ajusta cuándo pagar o resubir según la frecuencia con
  que apuesta el último rival agresor, estimada a partir de las jugadas observadas por alias.

Lince guarda el estado del torneo y las cartas propias de la mano en curso. Si se corta la conexión,
intenta recuperarla. Un supervisor en Windows relanza el proceso si se detiene.

## Validación local

En 15 ligas de prueba con 30 participantes, Lince v1.2 llegó a la final 12 veces y fue campeón 10, frente
a bots del kit. La prueba de robustez pasó 38 comprobaciones, incluida la recuperación de la fase, la ronda
y el asiento tras reiniciar durante la clasificatoria y la final.

[Informe de las ligas locales](../../../Simulaciones/ensayo_liga_real.md) · [Guía técnica](README.md)

## Créditos

Carlos y Hans desarrollaron y validaron Lince con apoyo de **Claude Code** y **Codex**. Claude Code colaboró
en el simulador, el entrenamiento y la adaptación de la estrategia final. Codex colaboró en el registro de
fase y en las pruebas de reinicio y cierre por reloj.
