package entrenamiento

import (
	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/estrategia"
)

// Un campo es una hipótesis sobre contra quién se juega la clasificatoria: qué
// estilos de bot hay inscritos y en qué proporción. Como no sabemos cómo van a
// jugar los demás equipos, cada versión del bot se mide en varios campos y
// tiene que servir en todos (plan A, tarea A1).

// TipoBot es un estilo de rival con nombre, para los reportes.
type TipoBot struct {
	Nombre   string
	Politica estrategia.Politica
}

// Campo es una mezcla de tipos de bot con sus pesos relativos.
type Campo struct {
	Nombre string
	Tipos  []TipoBot
	Pesos  []float64
}

// Armar devuelve n rivales del campo respetando las proporciones (método del
// resto mayor: sin sorteo, el mismo n da siempre la misma composición).
func (c Campo) Armar(n int) []TipoBot {
	var total float64
	for _, p := range c.Pesos {
		total += p
	}
	cuantos := make([]int, len(c.Tipos))
	restos := make([]float64, len(c.Tipos))
	asignados := 0
	for i, p := range c.Pesos {
		exacto := p / total * float64(n)
		cuantos[i] = int(exacto)
		restos[i] = exacto - float64(cuantos[i])
		asignados += cuantos[i]
	}
	for ; asignados < n; asignados++ {
		mejor := 0
		for i := range restos {
			if restos[i] > restos[mejor] {
				mejor = i
			}
		}
		cuantos[mejor]++
		restos[mejor] = -1
	}
	rivales := make([]TipoBot, 0, n)
	for i, k := range cuantos {
		for ; k > 0; k-- {
			rivales = append(rivales, c.Tipos[i])
		}
	}
	return rivales
}

// Campos arma los cuatro campos de prueba del plan. `darwin` es el genoma
// original de Gato Darwin y `prudente` el de la v0.
//
// Límite conocido: el simulador rápido no tiene réplica de Gato Bandido, y las
// de Michi Espía y Bot Estadístico son aproximadas (ver replicas.go).
//
// Historia: hasta el 14/11 los campos "kit" y "mixto" no tenían a Espía ni a
// Estadístico, y un genoma entrenado así ganaba mucho en el simulador pero
// poco contra esos dos bots en la arena real. Las cifras de
// Simulaciones/a1_b1_linea_base.md son de los campos viejos.
func Campos(darwin, prudente estrategia.Genoma) []Campo {
	var (
		aleatorio   = TipoBot{"aleatorio", Aleatorio{}}
		conservador = TipoBot{"conservador", Conservador{}}
		base        = TipoBot{"base", Base{}}
		novato      = TipoBot{"novato", Novato{}}
		roca        = TipoBot{"roca", Roca{}}
		maniaco     = TipoBot{"maniaco", Maniaco{}}
		estacion    = TipoBot{"estacion", Estacion{}}
		tDarwin     = TipoBot{"darwin", darwin}
		tPrudente   = TipoBot{"prudente", prudente}
		espia       = TipoBot{"espia", Espia{}}
		estadistico = TipoBot{"estadistico", Estadistico{}}
	)
	return []Campo{
		// Los equipos entregan los bots del kit casi sin cambios.
		{"kit", []TipoBot{aleatorio, conservador, base, novato, tDarwin, espia, estadistico}, []float64{1, 1, 1.5, 1.5, 2, 2, 2}},
		// Muchos bots que suben y van all-in: se eliminan entre ellos rápido.
		{"agresivo", []TipoBot{maniaco, aleatorio, tDarwin, estacion, base}, []float64{3, 2, 3, 1, 1}},
		// Muchos bots que esperan cartas: llegan casi todos a la ronda 10.
		{"tight", []TipoBot{roca, novato, base, conservador, tPrudente}, []float64{3, 2, 2, 1, 2}},
		// Un poco de todo: es el campo de las pruebas de promoción.
		{"mixto", []TipoBot{aleatorio, conservador, base, novato, roca, maniaco, estacion, tDarwin, tPrudente, espia, estadistico}, []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1.5, 1.5}},
	}
}
