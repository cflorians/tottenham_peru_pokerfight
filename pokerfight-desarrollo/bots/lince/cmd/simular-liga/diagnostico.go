package main

import (
	"fmt"
	"math/rand/v2"

	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/entrenamiento"
	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/estrategia"
)

// espiado envuelve a una política y recuerda su última decisión: cuando el
// héroe quiebra, esa decisión es la que lo dejó sin fichas.
type espiado struct {
	politica  estrategia.Politica
	situacion estrategia.Situacion
	decision  estrategia.Decision
}

func (e *espiado) Decidir(s estrategia.Situacion, azar *rand.Rand) estrategia.Decision {
	e.situacion = s
	e.decision = e.politica.Decidir(s, azar)
	return e.decision
}

type causa struct {
	casos              int
	fuerza, comprometi float64
	rivales            float64
}

// informarDiagnostico juega mesas sueltas de una ronda (todos con el stack
// inicial) con el héroe y rivales sorteados del campo, y cuenta con qué
// jugada se quedó sin fichas: es para saber qué arreglar, no para medir.
func informarDiagnostico(cfg entrenamiento.ConfigLiga, heroe heroe, campos []entrenamiento.Campo, bib entrenamiento.Biblioteca, mesas int, semilla uint64) {
	calles := []string{"preflop", "flop", "turn", "river"}
	fmt.Printf("\nDiagnóstico de %s: %d mesas de 5 por campo, una ronda de %d manos, todos con %d fichas.\n", heroe.nombre, mesas, cfg.ManosPorRonda, cfg.Mesa.Stack)
	fmt.Println("\"Última jugada\" es la decisión del héroe en la mano en que se quedó sin fichas. \"Fuerza\" es su probabilidad de ganarle a una mano al azar en ese momento; \"Stack en juego\" es la parte de sus fichas que esa jugada comprometía.")

	for _, campo := range campos {
		rivales := campo.Armar(19)
		causas := map[string]*causa{}
		var orden []string
		quiebras := 0
		for k := 0; k < mesas; k++ {
			s := semilla*104729 + uint64(k)
			azar := rand.New(rand.NewPCG(s, 0xD1A6))
			espia := &espiado{politica: heroe.politica}
			politicas := []estrategia.Politica{espia}
			for i := 0; i < 4; i++ {
				politicas = append(politicas, rivales[azar.IntN(len(rivales))].Politica)
			}
			// El héroe rota de silla.
			rot := k % len(politicas)
			politicas[0], politicas[rot] = politicas[rot], politicas[0]
			stacks := make([]int64, len(politicas))
			for i := range stacks {
				stacks[i] = cfg.Mesa.Stack
			}
			entrenamiento.JugarMesa(cfg.Mesa, politicas, stacks, bib, int(s*2654435761%uint64(len(bib))), cfg.ManosPorRonda, s,
				func(silla, _ int, _ int64) {
					if silla != rot {
						return
					}
					quiebras++
					sit, dec := espia.situacion, espia.decision
					var jugada string
					comprometido := 1.0
					switch dec.Tipo {
					case estrategia.TodoAdentro:
						jugada = "va all-in"
					case estrategia.Subir:
						comprometido = sit.Riesgo(dec.Total)
						if comprometido >= 0.6 {
							jugada = "sube con más del 60 % del stack"
						} else {
							jugada = "sube (menos del 60 % del stack)"
						}
					case estrategia.Pagar:
						comprometido = sit.Riesgo(sit.ApuestaActual)
						if comprometido >= 0.6 {
							jugada = "paga más del 60 % del stack"
						} else {
							jugada = "paga (menos del 60 % del stack)"
						}
					default:
						jugada = "pasa o se retira (ya estaba casi sin fichas)"
					}
					if dec.Motivo != "" && dec.Tipo != estrategia.Retirarse && dec.Tipo != estrategia.Pasar {
						jugada += " («" + dec.Motivo + "»)"
					}
					clave := jugada + " | " + calles[sit.Calle]
					c := causas[clave]
					if c == nil {
						c = &causa{}
						causas[clave] = c
						orden = append(orden, clave)
					}
					c.casos++
					c.fuerza += sit.Fuerza
					c.comprometi += comprometido
					c.rivales += float64(sit.Rivales)
				})
		}

		fmt.Printf("\n### Campo %s: quiebra en %s de las mesas\n\n", campo.Nombre, porcentaje(quiebras, mesas))
		fmt.Println("| Última jugada | Calle | % de las quiebras | Fuerza media | Stack en juego | Rivales en la mano |\n| --- | --- | ---: | ---: | ---: | ---: |")
		// De más a menos casos.
		for i := range orden {
			for j := i + 1; j < len(orden); j++ {
				if causas[orden[j]].casos > causas[orden[i]].casos {
					orden[i], orden[j] = orden[j], orden[i]
				}
			}
		}
		for _, clave := range orden {
			c := causas[clave]
			if c.casos*50 < quiebras { // Menos del 2 %: ruido.
				continue
			}
			n := float64(c.casos)
			var jugada, calle string
			fmt.Sscanf(clave, "%s", &jugada)
			for i := len(clave) - 1; i >= 0; i-- {
				if clave[i] == '|' {
					jugada, calle = clave[:i-1], clave[i+2:]
					break
				}
			}
			fmt.Printf("| %s | %s | %.0f %% | %.0f %% | %.0f %% | %.1f |\n", jugada, calle, 100*n/float64(quiebras), 100*c.fuerza/n, 100*c.comprometi/n, c.rivales/n)
		}
	}
}
