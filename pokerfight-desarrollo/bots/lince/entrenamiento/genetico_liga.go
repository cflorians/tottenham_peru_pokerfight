package entrenamiento

import (
	"math/rand/v2"
	"runtime"
	"slices"
	"sync"

	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/estrategia"
)

// Este archivo entrena con lo que de verdad se premia en el torneo, en vez de
// las ciegas ganadas por mano (bb/100) del genético original:
//
//   - Clasificatoria (plan A): el premio esperado U = 60 × m2 × P(llegar a la
//     final) + 30 × P(terminar 1.º), medido jugando ligas enteras.
//   - Mesa final (plan B): la probabilidad de ganar la final.
//
// El genético es el mismo (selección por torneo, cruce BLX-α, mutación
// gaussiana, elitismo); solo cambia la función de aptitud.

// AptitudFunc mide un genoma. `campeones` son los mejores de las generaciones
// anteriores (para la coevolución) y `semilla` es la misma para todos los
// genomas de una generación: todos juegan los mismos torneos.
type AptitudFunc func(g estrategia.Genoma, campeones []estrategia.Genoma, bib Biblioteca, semilla uint64) Evaluacion

// EvolucionarCon es Evolucionar con una función de aptitud cualquiera.
// `iniciales` entran en la primera generación (el resto es al azar): empezar
// desde genomas que ya funcionan ahorra muchas generaciones.
func EvolucionarCon(cfg ConfigGenetico, iniciales []estrategia.Genoma, bib Biblioteca, aptitud AptitudFunc, informar func(Generacion)) ([]Individuo, []estrategia.Genoma) {
	azar := rand.New(rand.NewPCG(cfg.Semilla, cfg.Semilla^0x6E6E))
	hilos := cfg.Hilos
	if hilos <= 0 {
		hilos = runtime.NumCPU()
	}

	poblacion := make([]Individuo, cfg.Poblacion)
	for i := range poblacion {
		switch {
		case i < len(iniciales):
			poblacion[i].Genoma = iniciales[i]
		case i < 3*len(iniciales):
			// Variantes cercanas de los iniciales: exploran alrededor de lo
			// que ya funciona.
			poblacion[i].Genoma = mutar(iniciales[i%len(iniciales)], 0.5, cfg.Sigma, azar)
		default:
			poblacion[i].Genoma = estrategia.Aleatorio(azar)
		}
	}

	var campeones []estrategia.Genoma
	for gen := 0; gen < cfg.Generaciones; gen++ {
		semilla := azar.Uint64()
		evaluarCon(poblacion, campeones, bib, aptitud, semilla, hilos)
		slices.SortStableFunc(poblacion, func(a, b Individuo) int {
			switch {
			case a.Evaluacion.Aptitud > b.Evaluacion.Aptitud:
				return -1
			case a.Evaluacion.Aptitud < b.Evaluacion.Aptitud:
				return 1
			}
			return 0
		})

		campeones = append(campeones, poblacion[0].Genoma)
		if informar != nil {
			informar(resumen(gen+1, poblacion))
		}
		if gen == cfg.Generaciones-1 {
			break
		}

		progreso := float64(gen) / float64(max(cfg.Generaciones-1, 1))
		sigma := cfg.Sigma * (1 - 0.75*progreso)

		siguiente := make([]Individuo, 0, cfg.Poblacion)
		for i := 0; i < cfg.Elite && i < len(poblacion); i++ {
			siguiente = append(siguiente, Individuo{Genoma: poblacion[i].Genoma})
		}
		for len(siguiente) < cfg.Poblacion {
			padre := torneo(poblacion, cfg.TamTorneo, azar)
			madre := torneo(poblacion, cfg.TamTorneo, azar)
			hijo := padre
			if azar.Float64() < cfg.ProbCruce {
				hijo = cruzar(padre, madre, azar)
			}
			hijo = mutar(hijo, cfg.ProbMutacion, sigma, azar)
			siguiente = append(siguiente, Individuo{Genoma: hijo})
		}
		poblacion = siguiente
	}
	return poblacion, campeones
}

func evaluarCon(poblacion []Individuo, campeones []estrategia.Genoma, bib Biblioteca, aptitud AptitudFunc, semilla uint64, hilos int) {
	trabajos := make(chan int)
	var espera sync.WaitGroup
	for w := 0; w < hilos; w++ {
		espera.Add(1)
		go func() {
			defer espera.Done()
			for i := range trabajos {
				poblacion[i].Evaluacion = aptitud(poblacion[i].Genoma, campeones, bib, semilla)
			}
		}()
	}
	for i := range poblacion {
		trabajos <- i
	}
	close(trabajos)
	espera.Wait()
}

// ElegirCampeonCon reevalúa a los candidatos con torneos que el genético nunca
// vio (otra biblioteca, otra semilla y sin coevolución) y devuelve el mejor
// con su evaluación, y las de todos en el mismo orden que los candidatos.
func ElegirCampeonCon(candidatos []estrategia.Genoma, validacion Biblioteca, aptitud AptitudFunc, semilla uint64) (int, []Evaluacion) {
	individuos := make([]Individuo, len(candidatos))
	for i, g := range candidatos {
		individuos[i].Genoma = g
	}
	evaluarCon(individuos, nil, validacion, aptitud, semilla, runtime.NumCPU())
	mejor := 0
	resultados := make([]Evaluacion, len(individuos))
	for i, ind := range individuos {
		resultados[i] = ind.Evaluacion
		if ind.Evaluacion.Aptitud > individuos[mejor].Evaluacion.Aptitud {
			mejor = i
		}
	}
	return mejor, resultados
}

// --- Aptitud de clasificatoria ---------------------------------------------------

// AptitudLiga mide a un genoma jugando ligas completas contra varios campos.
type AptitudLiga struct {
	Liga    ConfigLiga
	Campos  []Campo
	Tamanos []int // Cantidades de inscritos; se alternan entre las ligas.
	Ligas   int   // Ligas por campo en cada evaluación.
	// M2 es la probabilidad de ganar la final estando en ella: el peso de
	// "llegar a la final" en el premio esperado. Solo se usa si TopesFinal
	// está vacío.
	M2 float64
	// TopesFinal, si no está vacío, hace que cada liga en que el genoma
	// clasifica juegue además su mesa final (con los finalistas de esa liga y
	// uno de estos topes de manos, alternados): el premio de "llegar" deja de
	// ser 60 × M2 y pasa a ser 60 si gana la final y 0 si no. Es el objetivo
	// real del torneo, y castiga a un genoma que clasifica bien pero pierde
	// la final contra los rivales que llegan con él.
	TopesFinal []int
	// Forma suma hasta este valor según el puesto en la liga (todo al 1.º,
	// nada al último). "Clasificó o no" es un dato de sí o no y con pocas
	// ligas es puro ruido; el puesto dice además cuán cerca estuvo, y le da
	// al genético una pendiente por donde subir. En la validación va en 0.
	Forma float64
	// Coevolucion es cuántos rivales de cada liga se reemplazan por campeones
	// de generaciones anteriores.
	Coevolucion int
}

// Evaluar implementa AptitudFunc. Aptitud es el premio esperado en soles (más
// la forma); PorEscenario trae, por campo, U, M1 y M1b en ese orden, y al
// final U, M1 y M1b del conjunto. Con TopesFinal, U es el premio realmente
// cobrado (60 por final ganada y 30 por terminar 1.º).
func (a AptitudLiga) Evaluar(g estrategia.Genoma, campeones []estrategia.Genoma, bib Biblioteca, semilla uint64) Evaluacion {
	ev := Evaluacion{PorEscenario: make([]float64, 0, 3*len(a.Campos)+3)}
	var finalTotal, primeroTotal, formaTotal, premioTotal float64
	politicas := make([]estrategia.Politica, 0, 64)
	finalistas := make([]estrategia.Politica, 0, 8)
	// valorFinal es lo que vale haber clasificado en la liga k.
	valorFinal := func(res ResultadoLiga, k int, s uint64) float64 {
		if len(a.TopesFinal) == 0 {
			return 60 * a.M2
		}
		finalistas = finalistas[:0]
		yo := 0
		for i, f := range res.Finalistas {
			finalistas = append(finalistas, politicas[f])
			if f == 0 {
				yo = i
			}
		}
		tope := a.TopesFinal[k%len(a.TopesFinal)]
		if ganador, _ := JugarFinal(FinalDeLasBases(tope), finalistas, bib, int((s>>7)%uint64(len(bib))), s^0xF1A1); ganador == yo {
			return 60
		}
		return 0
	}

	for c, campo := range a.Campos {
		var final, primero, premio float64
		for k := 0; k < a.Ligas; k++ {
			n := a.Tamanos[k%len(a.Tamanos)]
			politicas = append(politicas[:0], g)
			for _, r := range campo.Armar(n - 1) {
				politicas = append(politicas, r.Politica)
			}
			for j := 0; j < a.Coevolucion && j < len(campeones); j++ {
				// Posiciones repartidas por el campo, para no reemplazar
				// siempre al mismo tipo de rival.
				politicas[1+(3+7*j)%(n-1)] = campeones[len(campeones)-1-j]
			}
			s := mezclar(semilla, uint64(c), uint64(k))
			res := JugarLiga(a.Liga, politicas, bib, int(s%uint64(len(bib))), s)
			if res.EsFinalista(0) {
				final++
				premio += valorFinal(res, k, s)
			}
			if res.Primero() == 0 {
				primero++
				premio += 30
			}
			formaTotal += 1 - float64(res.Puesto[0]-1)/float64(n-1)
			ev.ManosJugadas += a.Liga.Rondas * a.Liga.ManosPorRonda
		}
		ligas := float64(a.Ligas)
		ev.PorEscenario = append(ev.PorEscenario, premio/ligas, final/ligas, primero/ligas)
		finalTotal += final
		primeroTotal += primero
		premioTotal += premio
	}
	total := float64(a.Ligas * len(a.Campos))
	u := premioTotal / total
	ev.PorEscenario = append(ev.PorEscenario, u, finalTotal/total, primeroTotal/total)
	ev.Aptitud = u + a.Forma*formaTotal/total
	return ev
}

// --- Aptitud de mesa final ---------------------------------------------------------

// AptitudFinal mide a un genoma jugando mesas finales.
type AptitudFinal struct {
	// Rivales es de dónde salen los otros 5 finalistas: los tipos de bot que
	// llegan a la final en las ligas simuladas, con su frecuencia como peso.
	Rivales Campo
	// Fuertes, si tiene tipos, es un segundo grupo de rivales: las finales se
	// reparten por mitades entre los dos grupos y la aptitud es el PEOR de los
	// dos porcentajes. Así el genético no puede ganar aptitud aplastando a un
	// grupo y regalándose al otro: tiene que servir contra los dos.
	Fuertes Campo
	// PesoFuertes, entre 0 y 1, cambia la regla anterior: en vez del peor de
	// los dos porcentajes, la aptitud es su promedio con este peso para el
	// grupo Fuertes. Sirve para pedir "casi tan bueno como siempre contra los
	// rivales comunes, y lo mejor posible contra los fuertes". 0 = el peor.
	PesoFuertes float64
	Topes       []int // Topes de manos; se alternan entre las finales.
	Finales     int   // Finales por evaluación.
	// Coevolucion es cuántos de los 5 rivales se reemplazan por campeones de
	// generaciones anteriores.
	Coevolucion int
}

// Evaluar implementa AptitudFunc. Aptitud es el porcentaje de finales ganadas
// (M2); PorEscenario trae el porcentaje por tope, en el orden de Topes, y con
// Fuertes, al final, el porcentaje contra cada grupo.
func (a AptitudFinal) Evaluar(g estrategia.Genoma, campeones []estrategia.Genoma, bib Biblioteca, semilla uint64) Evaluacion {
	ganadas := make([]float64, len(a.Topes))
	jugadas := make([]float64, len(a.Topes))
	grupos := []Campo{a.Rivales}
	if len(a.Fuertes.Tipos) > 0 {
		grupos = append(grupos, a.Fuertes)
	}
	ganadasGrupo := make([]float64, len(grupos))
	jugadasGrupo := make([]float64, len(grupos))
	politicas := make([]estrategia.Politica, 6)
	politicas[0] = g

	ev := Evaluacion{}
	for k := 0; k < a.Finales; k++ {
		s := mezclar(semilla, 0xF1, uint64(k))
		azar := rand.New(rand.NewPCG(s, 0x52495641))
		// Los topes rotan cada final y el grupo cada vuelta completa de topes:
		// cada grupo juega con todos los topes.
		g := (k / len(a.Topes)) % len(grupos)
		grupo := grupos[g]
		var pesoTotal float64
		for _, p := range grupo.Pesos {
			pesoTotal += p
		}
		for i := 1; i < 6; i++ {
			tirada := azar.Float64() * pesoTotal
			elegido := len(grupo.Tipos) - 1
			for t, p := range grupo.Pesos {
				if tirada < p {
					elegido = t
					break
				}
				tirada -= p
			}
			politicas[i] = grupo.Tipos[elegido].Politica
		}
		for j := 0; j < a.Coevolucion && j < len(campeones); j++ {
			politicas[1+j] = campeones[len(campeones)-1-j]
		}
		t := k % len(a.Topes)
		ganador, manos := JugarFinal(FinalDeLasBases(a.Topes[t]), politicas, bib, int(s%uint64(len(bib))), s)
		jugadas[t]++
		jugadasGrupo[g]++
		if ganador == 0 {
			ganadas[t]++
			ganadasGrupo[g]++
		}
		ev.ManosJugadas += manos
	}

	var total float64
	for t := range a.Topes {
		ev.PorEscenario = append(ev.PorEscenario, 100*ganadas[t]/max(jugadas[t], 1))
		total += ganadas[t]
	}
	ev.Aptitud = 100 * total / float64(a.Finales)
	if len(grupos) > 1 {
		ev.Aptitud = 100
		for g := range grupos {
			m2 := 100 * ganadasGrupo[g] / max(jugadasGrupo[g], 1)
			ev.PorEscenario = append(ev.PorEscenario, m2)
			ev.Aptitud = min(ev.Aptitud, m2)
		}
		if a.PesoFuertes > 0 {
			n := len(ev.PorEscenario)
			ev.Aptitud = (1-a.PesoFuertes)*ev.PorEscenario[n-2] + a.PesoFuertes*ev.PorEscenario[n-1]
		}
	}
	return ev
}

// Finalistas es el campo del que salen los rivales de la mesa final: los
// tipos de bot que llegan a la final en las ligas simuladas, con el peso de su
// frecuencia (Simulaciones/a1_b1_linea_base.md, campo mixto con 20 inscritos),
// más el propio bot de clasificatoria, que puede ser cualquiera de los otros
// equipos que copie una idea parecida.
//
// `fuertes` son genomas nuestros ya entrenados: a la final llegan los mejores,
// así que si algún equipo trae un bot tan bueno como el nuestro, va a estar
// ahí. Entre todos pesan como el 40 % de los rivales.
func Finalistas(darwin, prudente, clasificatoria estrategia.Genoma, fuertes ...estrategia.Genoma) Campo {
	c := Campo{
		Nombre: "finalistas",
		Tipos: []TipoBot{
			{"novato", Novato{}}, {"roca", Roca{}}, {"prudente", prudente}, {"darwin", darwin}, {"base", Base{}},
			{"aleatorio", Aleatorio{}}, {"conservador", Conservador{}}, {"estacion", Estacion{}}, {"maniaco", Maniaco{}},
			{"clasificatoria", clasificatoria},
		},
		Pesos: []float64{1.32, 1.38, 1.14, 0.71, 0.46, 0.15, 0.13, 0.08, 0.04, 0.60},
	}
	for i, g := range fuertes {
		c.Tipos = append(c.Tipos, TipoBot{"entrenado-" + string(rune('a'+i)), g})
		c.Pesos = append(c.Pesos, 4/float64(len(fuertes)))
	}
	return c
}
