package entrenamiento

import (
	"math/rand/v2"
	"slices"

	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/estrategia"
)

// Este archivo simula el torneo entero, no una mesa suelta: la clasificatoria
// (liga de 10 rondas con fichas arrastradas) y la mesa final. Es el "nivel
// rápido" de la Fase 1 del plan: usa la biblioteca de repartos y JugarMano, así
// que una liga de 20 bots se juega en milisegundos.
//
// Lo que las bases no dicen (preguntas 2 a 4 del plan) está resuelto con un
// supuesto, marcado con SUPUESTO en el comentario de cada función.

// JugarMesa juega hasta `manos` manos con eliminación: las fichas no se
// reponen y quien se queda sin fichas deja de jugar. Termina antes si queda un
// solo jugador con fichas. Devuelve cuántas manos se jugaron.
//
// stacks[i] son las fichas de la silla i: se modifica en el lugar, y es lo que
// permite arrastrar las fichas de una ronda a la siguiente. Las sillas que
// llegan con 0 fichas no juegan.
//
// alQuebrar, si no es nil, se llama cuando una silla se queda sin fichas, con
// el número de mano (desde 0) y las fichas con que empezó esa mano.
//
// El botón avanza como en la mesa de la arena: una silla por mano, salteando
// a los eliminados.
func JugarMesa(mesa ConfigMesa, politicas []estrategia.Politica, stacks []int64, bib Biblioteca, desde, manos int, semilla uint64, alQuebrar func(silla, mano int, stackAntes int64)) int {
	n := len(politicas)
	var (
		enMesa   = make([]estrategia.Politica, 0, n)
		enMano   = make([]int64, 0, n)
		vivos    = make([]int, 0, n)
		asientos = make([]int, 0, n)
	)
	// Lo que cada uno hace en esta mesa queda a la vista de todos: es la
	// "memoria" de los bots que se adaptan al rival. Se pierde al cambiar de
	// mesa, como si en cada ronda los rivales fueran desconocidos (el bot real
	// además recuerda a los que ya vio: tiene algo más de información).
	perfiles := make([]estrategia.Perfil, n)
	enMesaPerfiles := make([]*estrategia.Perfil, 0, n)

	boton, jugadas := 0, 0
	for h := 0; h < manos; h++ {
		enMesa, enMano, vivos, asientos = enMesa[:0], enMano[:0], vivos[:0], asientos[:0]
		enMesaPerfiles = enMesaPerfiles[:0]
		botonMano := -1
		for i := 0; i < n; i++ {
			if s := (boton + i) % n; stacks[s] > 0 {
				botonMano = s
				break
			}
		}
		for s := 0; s < n; s++ {
			if stacks[s] <= 0 {
				continue
			}
			if s == botonMano {
				botonMano = len(vivos)
			}
			vivos = append(vivos, s)
			enMesa = append(enMesa, politicas[s])
			enMano = append(enMano, stacks[s])
			asientos = append(asientos, len(asientos))
			enMesaPerfiles = append(enMesaPerfiles, &perfiles[s])
		}
		if len(vivos) < 2 {
			break
		}

		neto := jugarMano(mesa, &bib[(desde+h)%len(bib)], enMesa, asientos, enMano, botonMano, mezclar(semilla, uint64(h)), enMesaPerfiles)
		for i, s := range vivos {
			stacks[s] += neto[i]
			if stacks[s] <= 0 && alQuebrar != nil {
				alQuebrar(s, h, enMano[i])
			}
		}
		jugadas++
		boton = (vivos[botonMano] + 1) % n
	}
	return jugadas
}

// --- Clasificatoria ----------------------------------------------------------

// ConfigLiga es el formato de la clasificatoria.
type ConfigLiga struct {
	Mesa          ConfigMesa // Ciegas y fichas iniciales.
	Rondas        int
	ManosPorRonda int
	Clasifican    int // Cuántos pasan a la final.
}

// LigaDeLasBases es la clasificatoria de las bases: 10 000 fichas, ciegas
// 50/100, 10 rondas de 50 manos, pasan 6.
func LigaDeLasBases() ConfigLiga {
	return ConfigLiga{
		Mesa:          ConfigMesa{CiegaChica: 50, CiegaGrande: 100, Stack: 10000},
		Rondas:        10,
		ManosPorRonda: 50,
		Clasifican:    6,
	}
}

// ResultadoLiga es cómo terminó cada participante. Los índices son los del
// slice de políticas que recibió JugarLiga.
type ResultadoLiga struct {
	Fichas []int64 // Al terminar la última ronda (0 = eliminado).
	// RondaQuiebra es la ronda (desde 1) en que se quedó sin fichas; 0 si
	// terminó la liga con fichas.
	RondaQuiebra []int
	// Puesto va de 1 a N: primero los que terminaron con fichas, de más a
	// menos, y después los eliminados, del último en caer al primero.
	Puesto []int
	// Finalistas son los Clasifican primeros puestos, en orden.
	Finalistas []int
	// VivosTrasRonda[r] es cuántos quedaban con fichas al terminar la ronda r+1.
	VivosTrasRonda []int
	// FichasAlEmpezar[r][i] son las fichas de i al empezar la ronda r+1. Sirve
	// para las tablas de valor (probabilidad de clasificar según ronda y stack).
	FichasAlEmpezar [][]int64
}

// Primero es el índice del que terminó 1.º en la clasificatoria.
func (r ResultadoLiga) Primero() int { return r.Finalistas[0] }

// EsFinalista dice si el participante i pasó a la final.
func (r ResultadoLiga) EsFinalista(i int) bool { return r.Puesto[i] <= len(r.Finalistas) }

// JugarLiga juega una clasificatoria completa.
//
// En cada ronda se sortean las mesas entre los que siguen con fichas, cada
// mesa juega ManosPorRonda manos con eliminación y las fichas se arrastran a
// la ronda siguiente. Al final pasan los Clasifican con más fichas; si quedan
// menos con fichas, se completa con los últimos eliminados (bases).
//
// Mismos `desde` y `semilla` dan la misma liga: mismos sorteos de mesas y
// mismas cartas en cada mesa. Así dos genomas se pueden comparar en la misma
// liga (hasta donde sus decisiones no cambien quién sigue vivo).
//
// SUPUESTOS (preguntas 2 a 4 a la organización):
//   - Entre eliminados, queda mejor el que cayó más tarde (ronda y número de
//     mano); en la misma mano, el que la empezó con más fichas.
//   - El empate en fichas se rompe por sorteo.
//   - La liga se juega entera aunque queden Clasifican o menos con fichas. No
//     cambia quién clasifica (los que siguen vivos ya son los últimos en caer),
//     solo las fichas con que terminan.
//   - Las mesas se arman como dice TamanosMesas.
func JugarLiga(cfg ConfigLiga, politicas []estrategia.Politica, bib Biblioteca, desde int, semilla uint64) ResultadoLiga {
	n := len(politicas)
	res := ResultadoLiga{
		Fichas:          make([]int64, n),
		RondaQuiebra:    make([]int, n),
		Puesto:          make([]int, n),
		VivosTrasRonda:  make([]int, cfg.Rondas),
		FichasAlEmpezar: make([][]int64, cfg.Rondas),
	}
	for i := range res.Fichas {
		res.Fichas[i] = cfg.Mesa.Stack
	}
	// Cuándo cayó cada uno, para ordenar a los eliminados.
	manoQuiebra := make([]int, n)
	stackQuiebra := make([]int64, n)

	// Cada mesa de cada ronda tiene su tramo propio de la biblioteca, fijo
	// según (ronda, mesa): no depende de cuántos sigan vivos.
	maxMesas := n/5 + 2

	vivos := make([]int, 0, n)
	for ronda := 0; ronda < cfg.Rondas; ronda++ {
		res.FichasAlEmpezar[ronda] = slices.Clone(res.Fichas)
		vivos = vivos[:0]
		for i, f := range res.Fichas {
			if f > 0 {
				vivos = append(vivos, i)
			}
		}
		if len(vivos) >= 2 {
			azar := rand.New(rand.NewPCG(mezclar(semilla, uint64(ronda)), 0x4C494741))
			azar.Shuffle(len(vivos), func(a, b int) { vivos[a], vivos[b] = vivos[b], vivos[a] })

			inicio := 0
			for m, tam := range TamanosMesas(len(vivos)) {
				sentados := vivos[inicio : inicio+tam]
				inicio += tam

				enMesa := make([]estrategia.Politica, tam)
				stacks := make([]int64, tam)
				for s, i := range sentados {
					enMesa[s] = politicas[i]
					stacks[s] = res.Fichas[i]
				}
				tramo := desde + (ronda*maxMesas+m)*cfg.ManosPorRonda
				JugarMesa(cfg.Mesa, enMesa, stacks, bib, tramo, cfg.ManosPorRonda, mezclar(semilla, uint64(ronda), uint64(m)),
					func(silla, mano int, stackAntes int64) {
						i := sentados[silla]
						res.RondaQuiebra[i] = ronda + 1
						manoQuiebra[i] = mano
						stackQuiebra[i] = stackAntes
					})
				for s, i := range sentados {
					res.Fichas[i] = stacks[s]
				}
			}
		}
		for _, f := range res.Fichas {
			if f > 0 {
				res.VivosTrasRonda[ronda]++
			}
		}
	}

	// Clasificación final. El sorteo de desempate se fija antes de ordenar.
	azar := rand.New(rand.NewPCG(mezclar(semilla, 0xD15E), 0x4C494741))
	sorteo := azar.Perm(n)
	orden := make([]int, n)
	for i := range orden {
		orden[i] = i
	}
	slices.SortFunc(orden, func(a, b int) int {
		switch {
		case res.Fichas[a] != res.Fichas[b]:
			return int(sign(res.Fichas[b] - res.Fichas[a]))
		case res.RondaQuiebra[a] != res.RondaQuiebra[b]:
			return res.RondaQuiebra[b] - res.RondaQuiebra[a]
		case manoQuiebra[a] != manoQuiebra[b]:
			return manoQuiebra[b] - manoQuiebra[a]
		case stackQuiebra[a] != stackQuiebra[b]:
			return int(sign(stackQuiebra[b] - stackQuiebra[a]))
		}
		return sorteo[a] - sorteo[b]
	})
	for puesto, i := range orden {
		res.Puesto[i] = puesto + 1
	}
	res.Finalistas = slices.Clone(orden[:min(cfg.Clasifican, n)])
	return res
}

func sign(v int64) int64 {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}

// TamanosMesas reparte n jugadores en mesas de 5, con algunas de 6 (bases).
//
// SUPUESTO (pregunta 3): cuando no alcanza con mesas de 5 y 6 (7, 8, 9, 13 o
// 14 jugadores) se arma una mesa más y se reparten lo más parejo posible
// (mesas de 4, o de 3 y 4 con 7 jugadores). Con 6 o menos, una sola mesa.
func TamanosMesas(n int) []int {
	if n <= 6 {
		if n < 2 {
			return nil
		}
		return []int{n}
	}
	mesas := n / 5
	if n-5*mesas > mesas { // Sobran más de los que entran como sexto jugador.
		mesas++
	}
	tamanos := make([]int, mesas)
	for i := range tamanos {
		tamanos[i] = n / mesas
		if i < n%mesas {
			tamanos[i]++
		}
	}
	return tamanos
}

// --- Mesa final --------------------------------------------------------------

// ConfigFinal es el formato de la mesa final.
type ConfigFinal struct {
	Mesa ConfigMesa // Todos empiezan con Mesa.Stack: las fichas de la liga no cuentan.
	// TopeManos representa la hora de reloj de las bases: cuántas manos entran
	// en una hora depende del ritmo real de los bots, que no conocemos. Al
	// llegar al tope gana el que tiene más fichas.
	TopeManos int
}

// TopeSinReloj es el tope que se usa como "sin límite de manos": alcanza para
// que casi todas las finales terminen con un solo jugador en pie, y corta las
// que no terminarían nunca (dos bots que solo se pasan las ciegas).
const TopeSinReloj = 5000

// FinalDeLasBases es la mesa final de las bases con el tope de manos dado
// (0 = TopeSinReloj).
func FinalDeLasBases(topeManos int) ConfigFinal {
	if topeManos <= 0 {
		topeManos = TopeSinReloj
	}
	return ConfigFinal{Mesa: ConfigMesa{CiegaChica: 50, CiegaGrande: 100, Stack: 10000}, TopeManos: topeManos}
}

// JugarFinal juega una mesa final y devuelve el índice del ganador (el último
// en pie o, al llegar al tope, el que tiene más fichas) y las manos jugadas.
// Las sillas se sortean. SUPUESTO: el empate en fichas al llegar al tope se
// rompe por sorteo.
func JugarFinal(cfg ConfigFinal, politicas []estrategia.Politica, bib Biblioteca, desde int, semilla uint64) (ganador, manos int) {
	n := len(politicas)
	azar := rand.New(rand.NewPCG(mezclar(semilla, 0xF1A1), 0x4C494741))
	sillas := azar.Perm(n) // sillas[s] = qué participante se sienta en la silla s.

	enMesa := make([]estrategia.Politica, n)
	stacks := make([]int64, n)
	for s, i := range sillas {
		enMesa[s] = politicas[i]
		stacks[s] = cfg.Mesa.Stack
	}
	manos = JugarMesa(cfg.Mesa, enMesa, stacks, bib, desde, cfg.TopeManos, semilla, nil)

	mejor := 0
	for s := 1; s < n; s++ {
		// Con empate gana la silla más baja, y las sillas ya salieron de un sorteo.
		if stacks[s] > stacks[mejor] {
			mejor = s
		}
	}
	return sillas[mejor], manos
}
