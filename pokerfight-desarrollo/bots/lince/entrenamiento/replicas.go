package entrenamiento

import (
	"math"
	"math/rand/v2"

	"github.com/aicspucp/pokerfight-arena-participantes/pkg/botsdk"
	"github.com/aicspucp/pokerfight-arena-participantes/pkg/poker"

	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/estrategia"
)

// Réplicas exactas de dos bots de ejemplo, con las mismas reglas sobre las
// mismas cartas. Las aproximaciones por fuerza de rivales.go sirven para
// variar estilos, pero un bot tight se distingue justamente por QUÉ manos
// juega: imitarlo con un umbral de fuerza le regala al genético un rival más
// fácil que el de verdad, y lo que aprende no se sostiene en la arena.

// politicaConCartas la implementan los rivales que deciden mirando las cartas
// y no solo la Situacion. El simulador se las pasa si la implementan.
type politicaConCartas interface {
	DecidirConCartas(s estrategia.Situacion, mano botsdk.Mano, mesa []botsdk.Carta, azar *rand.Rand) estrategia.Decision
}

// --- bots/base de la arena ---------------------------------------------------

// DecidirConCartas replica bots/base/main.go: Chen preflop y jugada hecha
// (pkg/poker) después del flop.
func (Base) DecidirConCartas(s estrategia.Situacion, mano botsdk.Mano, mesa []botsdk.Carta, _ *rand.Rand) estrategia.Decision {
	if len(mesa) < 3 {
		chen := puntuacionChen(mano)
		switch {
		case chen >= 9:
			return subir(3)
		case chen >= 5:
			if s.PorIgualar <= 0 {
				return pasar()
			}
			if s.PorIgualar <= 3 {
				return pagar()
			}
		}
		return retirarse()
	}

	switch categoria := poker.Mejor(mano, mesa).Categoria; {
	case categoria >= poker.Trio:
		if s.PuedeSubir {
			return subir(s.ApuestaActual + 0.75*(s.Pozo+s.PorIgualar))
		}
		return pagar()
	case categoria >= poker.Par:
		if s.PorIgualar <= 0 && s.PuedeSubir {
			return subir(s.ApuestaActual + 0.4*s.Pozo)
		}
		if s.PorIgualar > 0 && s.PorIgualar*2 <= s.Pozo {
			return pagar()
		}
	}
	return retirarse()
}

// puntuacionChen es la misma fórmula de bots/base.
func puntuacionChen(mano botsdk.Mano) float64 {
	alta, baja := mano[0].Rango, mano[1].Rango
	if baja > alta {
		alta, baja = baja, alta
	}
	var puntaje float64
	switch alta {
	case botsdk.As:
		puntaje = 10
	case botsdk.Rey:
		puntaje = 8
	case botsdk.Reina:
		puntaje = 7
	case botsdk.Jota:
		puntaje = 6
	default:
		puntaje = float64(alta) / 2
	}
	if alta == baja {
		puntaje = max(puntaje*2, 5)
	}
	if mano[0].Palo == mano[1].Palo {
		puntaje += 2
	}
	hueco := int(alta) - int(baja) - 1
	switch {
	case hueco <= 0:
	case hueco == 1:
		puntaje--
	case hueco == 2:
		puntaje -= 2
	case hueco == 3:
		puntaje -= 4
	default:
		puntaje -= 5
	}
	if hueco <= 1 && alta < botsdk.Reina && alta != baja {
		puntaje++
	}
	return puntaje
}

// --- Michi Novato (el bot fácil de Python) -------------------------------------

// DecidirConCartas replica estrategia.py de Michi Novato regla por regla.
func (Novato) DecidirConCartas(s estrategia.Situacion, mano botsdk.Mano, mesa []botsdk.Carta, _ *rand.Rand) estrategia.Decision {
	hablaUltimo := s.Posicion >= 1
	if len(mesa) < 3 {
		nadieSubio := s.ApuestaActual <= 1
		switch grupoPreflop(mano) {
		case "premium":
			if nadieSubio {
				return subir(3)
			}
			if s.ApuestaActual*3 >= s.Techo {
				return estrategia.Decision{Tipo: estrategia.TodoAdentro}
			}
			return subir(3 * s.ApuestaActual)
		case "fuerte":
			if nadieSubio {
				return subir(3)
			}
			if s.PorIgualar <= 4 {
				return pagar()
			}
		case "jugable":
			if nadieSubio && hablaUltimo {
				return subir(3)
			}
			if s.PorIgualar <= 1 {
				return pagar()
			}
		case "especulativa":
			if s.PorIgualar <= 1 && s.Rivales >= 2 {
				return pagar()
			}
		}
		return retirarse()
	}

	nivel, outs := leerJugada(mano, mesa)
	muchos := s.Rivales >= 3
	pozoFraccion := func(f float64) estrategia.Decision {
		return subir(s.ApuestaActual + (s.Pozo+s.PorIgualar)*f)
	}

	switch {
	case nivel == nivelMonstruo:
		if s.PorIgualar <= 0 {
			return pozoFraccion(0.75)
		}
		if s.PuedeSubir {
			return subir(s.ApuestaActual * 3)
		}
		return pagar()
	case nivel == nivelBuena:
		if s.PorIgualar <= 0 {
			if muchos {
				return pozoFraccion(0.33)
			}
			return pozoFraccion(0.5)
		}
		limite := 0.75
		if muchos {
			limite = 0.5
		}
		if s.PorIgualar <= s.Pozo*limite {
			return pagar()
		}
		return retirarse()
	case outs > 0:
		probabilidad := float64(outs * 2)
		if len(mesa) == 3 {
			probabilidad = float64(outs * 4)
		}
		if s.PorIgualar <= 0 {
			if hablaUltimo && s.Rivales == 1 {
				return pozoFraccion(0.5)
			}
			return pagar()
		}
		if probabilidad >= s.PorIgualar*100/(s.Pozo+s.PorIgualar) {
			return pagar()
		}
		return retirarse()
	case nivel == nivelMedia:
		if s.PorIgualar <= 0 {
			return pagar()
		}
		if s.PorIgualar <= s.Pozo/3 && !muchos {
			return pagar()
		}
		return retirarse()
	}
	if s.PorIgualar <= 0 && s.Calle == 3 && s.Rivales == 1 && hablaUltimo && s.Subidas == 0 {
		return pozoFraccion(0.5)
	}
	return retirarse()
}

func grupoPreflop(mano botsdk.Mano) string {
	alta, baja := max(mano[0].Rango, mano[1].Rango), min(mano[0].Rango, mano[1].Rango)
	mismoPalo := mano[0].Palo == mano[1].Palo
	pareja := alta == baja
	switch {
	case pareja && alta >= botsdk.Jota, alta == botsdk.As && baja == botsdk.Rey:
		return "premium"
	case pareja && alta >= botsdk.Ocho, alta == botsdk.As && baja >= botsdk.Jota, alta == botsdk.Rey && baja == botsdk.Reina && mismoPalo:
		return "fuerte"
	case pareja, alta == botsdk.As && (mismoPalo || baja >= botsdk.Diez), alta >= botsdk.Jota && baja >= botsdk.Diez:
		return "jugable"
	case mismoPalo && alta-baja == 1 && baja >= botsdk.Cinco:
		return "especulativa"
	}
	return "basura"
}

const (
	nivelNada = iota
	nivelMedia
	nivelBuena
	nivelMonstruo
)

// categoriaPorConteo es la categoria() de cartas.py: funciona con cualquier
// cantidad de cartas (también con las 3 del flop solas, que pkg/poker no
// evalúa porque no llegan a 5).
func categoriaPorConteo(cartas []botsdk.Carta) poker.Categoria {
	var rangos [15]int
	var palos [4]int
	for _, c := range cartas {
		rangos[c.Rango]++
		palos[c.Palo]++
	}
	paloColor := -1
	for p, n := range palos {
		if n >= 5 {
			paloColor = p
		}
	}
	if paloColor >= 0 {
		var delColor [15]bool
		for _, c := range cartas {
			if int(c.Palo) == paloColor {
				delColor[c.Rango] = true
			}
		}
		if hayEscalera(delColor) {
			return poker.EscaleraDeColor
		}
	}
	primero, segundo := 0, 0
	var presentes [15]bool
	for r := 2; r <= 14; r++ {
		n := rangos[r]
		presentes[r] = n > 0
		switch {
		case n > primero:
			primero, segundo = n, primero
		case n > segundo:
			segundo = n
		}
	}
	switch {
	case primero == 4:
		return poker.Poker
	case primero == 3 && segundo >= 2:
		return poker.FullHouse
	case paloColor >= 0:
		return poker.Color
	case hayEscalera(presentes):
		return poker.Escalera
	case primero == 3:
		return poker.Trio
	case primero == 2 && segundo == 2:
		return poker.DoblePar
	case primero == 2:
		return poker.Par
	}
	return poker.CartaAlta
}

func hayEscalera(presentes [15]bool) bool {
	seguidas := 0
	for r := 1; r <= 14; r++ {
		esta := presentes[r]
		if r == 1 {
			esta = presentes[14] // El as también vale 1.
		}
		if esta {
			seguidas++
			if seguidas >= 5 {
				return true
			}
		} else {
			seguidas = 0
		}
	}
	return false
}

// leerJugada es leer_jugada() de cartas.py: el nivel de la mano y los outs.
func leerJugada(mano botsdk.Mano, mesa []botsdk.Carta) (nivel, outs int) {
	todas := append(append(make([]botsdk.Carta, 0, 7), mano[:]...), mesa...)
	mia := categoriaPorConteo(todas)
	deLaMesa := categoriaPorConteo(mesa)

	masAlta := botsdk.Rango(0)
	enMesa := map[botsdk.Rango]bool{}
	for _, c := range mesa {
		masAlta = max(masAlta, c.Rango)
		enMesa[c.Rango] = true
	}
	parejaEnMano := mano[0].Rango == mano[1].Rango

	switch {
	case mia >= poker.Trio && mia > deLaMesa:
		nivel = nivelMonstruo
	case mia == poker.DoblePar && deLaMesa == poker.CartaAlta:
		nivel = nivelBuena
	case parejaEnMano && mano[0].Rango > masAlta:
		nivel = nivelBuena
	case mano[0].Rango == masAlta || mano[1].Rango == masAlta:
		nivel = nivelBuena
	case parejaEnMano || enMesa[mano[0].Rango] || enMesa[mano[1].Rango]:
		nivel = nivelMedia
	}

	if len(mesa) < 5 && nivel != nivelMonstruo {
		if proyectoDeColor(mano, mesa) {
			outs += 9
		}
		switch proyectoDeEscalera(mano, mesa) {
		case "abierta":
			outs += 8
		case "interna":
			outs += 4
		}
	}
	return nivel, min(outs, 15)
}

func proyectoDeColor(mano botsdk.Mano, mesa []botsdk.Carta) bool {
	for p := botsdk.Treboles; p <= botsdk.Picas; p++ {
		total, mias := 0, 0
		for _, c := range mesa {
			if c.Palo == p {
				total++
			}
		}
		for _, c := range mano {
			if c.Palo == p {
				total++
				mias++
			}
		}
		if total == 4 && mias >= 1 {
			return true
		}
	}
	return false
}

func proyectoDeEscalera(mano botsdk.Mano, mesa []botsdk.Carta) string {
	var rangos, mios [15]bool
	marcar := func(m *[15]bool, r botsdk.Rango) {
		m[r] = true
		if r == botsdk.As {
			m[1] = true
		}
	}
	for _, c := range mesa {
		marcar(&rangos, c.Rango)
	}
	for _, c := range mano {
		marcar(&rangos, c.Rango)
		marcar(&mios, c.Rango)
	}

	resultado := ""
	for bajo := 1; bajo <= 10; bajo++ {
		var presentes []int
		usaMia := false
		falta := -1
		for r := bajo; r < bajo+5; r++ {
			if rangos[r] {
				presentes = append(presentes, r)
				usaMia = usaMia || mios[r]
			} else {
				falta = r
			}
		}
		if len(presentes) != 4 || !usaMia {
			continue
		}
		if falta == bajo || falta == bajo+4 {
			if presentes[0]-1 >= 1 && presentes[3]+1 <= 14 {
				return "abierta"
			}
		}
		resultado = "interna"
	}
	return resultado
}

// --- Roca -------------------------------------------------------------------

// Roca juega muy pocas manos y, cuando apuesta, tiene. Es el rival que
// castiga a un bot que paga de más o que va all-in con manos medianas: contra
// una roca, la fuerza "contra una mano al azar" sobreestima mucho lo que vale
// la propia mano.
type Roca struct{}

func (Roca) Decidir(s estrategia.Situacion, _ *rand.Rand) estrategia.Decision {
	if s.Calle == 0 {
		switch {
		case s.Fuerza >= 0.72:
			return subir(max(3, 3*s.ApuestaActual))
		case s.Fuerza >= 0.64 && s.PorIgualar <= 4:
			return pagar()
		}
		return retirarse()
	}
	switch {
	case s.Fuerza >= 0.88:
		if s.PuedeSubir {
			return subir(s.ApuestaActual + s.Pozo + s.PorIgualar)
		}
		return pagar()
	case s.Fuerza >= 0.75:
		if s.PorIgualar <= 0 && s.PuedeSubir {
			return subir(0.5 * s.Pozo)
		}
		if s.PorIgualar <= s.Pozo {
			return pagar()
		}
	}
	return retirarse()
}

// --- Atajos ----------------------------------------------------------------------

func subir(total float64) estrategia.Decision {
	return estrategia.Decision{Tipo: estrategia.Subir, Total: total}
}
func pagar() estrategia.Decision     { return estrategia.Decision{Tipo: estrategia.Pagar} }
func pasar() estrategia.Decision     { return estrategia.Decision{Tipo: estrategia.Pasar} }
func retirarse() estrategia.Decision { return estrategia.Decision{Tipo: estrategia.Retirarse} }

// --- Réplicas de los bots fuertes del kit ------------------------------------
//
// Michi Espía y Bot Estadístico deciden con la equidad contra varias manos al
// azar, estimada por Monte Carlo. Aquí no se simula: la equidad contra k
// rivales se aproxima con fuerza^k, donde fuerza es la probabilidad
// precalculada de ganarle a UNA mano al azar. La fórmula ignora que las manos
// de los rivales comparten mazo y mesa, pero conserva lo que importa: cuánto
// cae el valor de una mano con cada rival extra.

func equidadContra(fuerza float64, rivales int) float64 {
	return math.Pow(fuerza, float64(max(rivales, 1)))
}

// apostarPozo es botsdk.Turno.ApostarPozo: llevar la apuesta a la actual más
// una fracción del pozo que quedaría tras pagar.
func apostarPozo(s estrategia.Situacion, fraccion float64) estrategia.Decision {
	return subir(s.ApuestaActual + fraccion*(s.Pozo+s.PorIgualar))
}

// Estadistico replica bots/estadistico/decision.go: sube si su equidad supera
// con margen la "parte justa" (1 entre los que quedan), paga si la equidad
// cubre el precio, y cerca de cada umbral sortea la decisión.
type Estadistico struct{}

func (Estadistico) Decidir(s estrategia.Situacion, azar *rand.Rand) estrategia.Decision {
	const (
		umbralSubir, recargoPorSubida, margenPagar = 0.25, 0.15, 0.04
		fraccionMinima, fraccionMaxima             = 0.5, 1.0
		probabilidadFarol, fraccionFarol           = 0.08, 0.5
	)
	rivales := max(s.Rivales, 1)
	eq := equidadContra(s.Fuerza, rivales)
	parteJusta := 1 / float64(rivales+1)
	ventaja := (eq - parteJusta) / (1 - parteJusta)
	umbral := umbralSubir + recargoPorSubida*float64(s.Subidas)

	if s.PuedeSubir && superaUmbral(ventaja, umbral, azar) {
		avance := 1.0
		if umbral < 1 {
			avance = min(max((ventaja-umbral)/(1-umbral), 0), 1)
		}
		return apostarPozo(s, fraccionMinima+(fraccionMaxima-fraccionMinima)*avance)
	}
	if s.PorIgualar <= 0 {
		if s.Calle >= 1 && rivales <= 1 && s.PuedeSubir && azar.Float64() < probabilidadFarol {
			return apostarPozo(s, fraccionFarol)
		}
		return pasar()
	}
	if superaUmbral(eq, s.PorIgualar/(s.Pozo+s.PorIgualar)+margenPagar, azar) {
		return pagar()
	}
	return retirarse()
}

// superaUmbral sortea la decisión en una franja de ±0,05 alrededor del umbral,
// como el bot original.
func superaUmbral(valor, umbral float64, azar *rand.Rand) bool {
	const ancho = 0.05
	switch {
	case valor >= umbral+ancho:
		return true
	case valor <= umbral-ancho:
		return false
	}
	return azar.Float64() < (valor-(umbral-ancho))/(2*ancho)
}

// Espia replica bots/espia/main.go con el modelo de rivales en sus valores de
// partida (rival desconocido: se retira la mitad de las veces, agresión
// típica). El Espía de verdad ajusta esos números a cada rival con las manos;
// en una ronda de 50 manos alcanza a moverlos poco.
type Espia struct{}

func (Espia) Decidir(s estrategia.Situacion, _ *rand.Rand) estrategia.Decision {
	rivales := max(s.Rivales, 1)
	if s.Calle == 0 {
		eq := s.Fuerza
		huboSubida := s.ApuestaActual > 1
		exigencia := 0.02 * float64(rivales-1)
		switch {
		case eq >= 0.70:
			return subir(max(3, 3*s.ApuestaActual))
		case eq >= 0.58+exigencia && !huboSubida:
			return subir(3)
		case eq >= 0.50+exigencia && s.PorIgualar <= 4:
			return pagar()
		}
		if s.PorIgualar <= 0 {
			return pasar()
		}
		return retirarse()
	}

	simulados := min(rivales, 3)
	eq := equidadContra(s.Fuerza, simulados)
	parteJusta := 1 / float64(simulados+1)
	if s.PorIgualar <= 0 {
		seRetiran := math.Pow(0.5, float64(rivales))
		switch {
		case eq >= parteJusta+0.22:
			if seRetiran < 0.30 {
				return apostarPozo(s, 0.9)
			}
			return apostarPozo(s, 0.6)
		case eq >= parteJusta+0.10 && seRetiran < 0.35:
			return apostarPozo(s, 0.5)
		}
		return pasar()
	}
	const creible = 0.10
	precio := s.PorIgualar / (s.Pozo + s.PorIgualar)
	switch {
	case eq >= parteJusta+0.35:
		return apostarPozo(s, 1)
	case eq-creible >= precio+0.03:
		return pagar()
	}
	return retirarse()
}
