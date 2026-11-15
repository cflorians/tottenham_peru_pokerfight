package entrenamiento

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

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
//
// `fuertes` son genomas nuestros ya entrenados (los de bots/lince/rivales). Si
// hay alguno se agrega el campo "fuerte": otros equipos pueden llegar con bots
// tan buenos como el nuestro, y un genoma que solo se midió contra los bots
// del kit puede ser fácil de explotar por uno que juegue como él.
func Campos(darwin, prudente estrategia.Genoma, fuertes ...estrategia.Genoma) []Campo {
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
	campos := []Campo{
		// Los equipos entregan los bots del kit casi sin cambios.
		{"kit", []TipoBot{aleatorio, conservador, base, novato, tDarwin, espia, estadistico}, []float64{1, 1, 1.5, 1.5, 2, 2, 2}},
		// Muchos bots que suben y van all-in: se eliminan entre ellos rápido.
		{"agresivo", []TipoBot{maniaco, aleatorio, tDarwin, estacion, base}, []float64{3, 2, 3, 1, 1}},
		// Muchos bots que esperan cartas: llegan casi todos a la ronda 10.
		{"tight", []TipoBot{roca, novato, base, conservador, tPrudente}, []float64{3, 2, 2, 1, 2}},
		// Un poco de todo: es el campo de las pruebas de promoción.
		{"mixto", []TipoBot{aleatorio, conservador, base, novato, roca, maniaco, estacion, tDarwin, tPrudente, espia, estadistico}, []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1.5, 1.5}},
	}
	if len(fuertes) > 0 {
		// La mitad del campo son genomas entrenados como el nuestro; el
		// resto, los mejores bots del kit.
		fuerte := Campo{Nombre: "fuerte", Tipos: []TipoBot{novato, espia, estadistico, roca}, Pesos: []float64{1, 1, 1, 1}}
		for i, g := range fuertes {
			fuerte.Tipos = append(fuerte.Tipos, TipoBot{"entrenado-" + string(rune('a'+i)), g})
			fuerte.Pesos = append(fuerte.Pesos, 4/float64(len(fuertes)))
		}
		campos = append(campos, fuerte)

		// El campo mixto con unos pocos bots entrenados (uno de cada siete):
		// lo que parece más probable entre 30 equipos. La mayoría juega como
		// el kit y dos o tres apuestan chico y muy seguido.
		mixto := campos[3]
		realista := Campo{Nombre: "realista", Tipos: append([]TipoBot{}, mixto.Tipos...), Pesos: append([]float64{}, mixto.Pesos...)}
		for i, g := range fuertes {
			realista.Tipos = append(realista.Tipos, TipoBot{"entrenado-" + string(rune('a'+i)), g})
			realista.Pesos = append(realista.Pesos, 2/float64(len(fuertes)))
		}
		campos = append(campos, realista)
	}
	return campos
}

// LeerRivales lee todos los genomas .json de una carpeta, en orden alfabético.
// Una carpeta que no existe o está vacía no es un error: no hay campo fuerte.
func LeerRivales(carpeta string) ([]estrategia.Genoma, error) {
	rutas, err := filepath.Glob(filepath.Join(carpeta, "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(rutas)
	var genomas []estrategia.Genoma
	for _, ruta := range rutas {
		datos, err := os.ReadFile(ruta)
		if err != nil {
			return nil, err
		}
		g, _, err := estrategia.LeerGenoma(datos)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", ruta, err)
		}
		genomas = append(genomas, g)
	}
	return genomas, nil
}
