package entrenamiento

import (
	"slices"
	"testing"

	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/estrategia"
)

func bibliotecaDePrueba() Biblioteca {
	return NuevaBiblioteca(ConfigBiblioteca{Repartos: 400, Muestras: 20, Semilla: 7}, nil)
}

func campoDePrueba(n int) []estrategia.Politica {
	tipos := []estrategia.Politica{Aleatorio{}, Maniaco{}, Base{}, Novato{}, Roca{}, Conservador{}, Estacion{}, estrategia.GenomaInicial()}
	ps := make([]estrategia.Politica, n)
	for i := range ps {
		ps[i] = tipos[i%len(tipos)]
	}
	return ps
}

func TestTamanosMesas(t *testing.T) {
	casos := map[int][]int{
		0: nil, 1: nil, 2: {2}, 5: {5}, 6: {6},
		7: {4, 3}, 8: {4, 4}, 9: {5, 4}, 10: {5, 5}, 11: {6, 5}, 12: {6, 6},
		13: {5, 4, 4}, 14: {5, 5, 4}, 15: {5, 5, 5}, 16: {6, 5, 5},
		20: {5, 5, 5, 5}, 23: {6, 6, 6, 5}, 40: {5, 5, 5, 5, 5, 5, 5, 5},
	}
	for n, quiero := range casos {
		if got := TamanosMesas(n); !slices.Equal(got, quiero) {
			t.Errorf("TamanosMesas(%d) = %v, quiero %v", n, got, quiero)
		}
	}
	for n := 2; n <= 60; n++ {
		suma := 0
		for _, tam := range TamanosMesas(n) {
			suma += tam
			if tam < 2 || tam > MaxAsientos {
				t.Errorf("TamanosMesas(%d) tiene una mesa de %d", n, tam)
			}
		}
		if suma != n {
			t.Errorf("TamanosMesas(%d) sienta a %d", n, suma)
		}
	}
}

func TestLigaConservaFichasYClasificaSeis(t *testing.T) {
	bib := bibliotecaDePrueba()
	cfg := LigaDeLasBases()
	for _, n := range []int{4, 6, 7, 13, 20, 31} {
		for semilla := uint64(1); semilla <= 20; semilla++ {
			res := JugarLiga(cfg, campoDePrueba(n), bib, int(semilla)*37, semilla)

			var total int64
			vivos := 0
			for i, f := range res.Fichas {
				total += f
				if f < 0 {
					t.Fatalf("n=%d semilla=%d: fichas negativas", n, semilla)
				}
				if f > 0 {
					vivos++
				}
				if (f == 0) != (res.RondaQuiebra[i] > 0) {
					t.Fatalf("n=%d semilla=%d: jugador %d con %d fichas y ronda de quiebra %d", n, semilla, i, f, res.RondaQuiebra[i])
				}
			}
			if quiero := int64(n) * cfg.Mesa.Stack; total != quiero {
				t.Fatalf("n=%d semilla=%d: hay %d fichas, tendría que haber %d", n, semilla, total, quiero)
			}
			if vivos != res.VivosTrasRonda[cfg.Rondas-1] {
				t.Fatalf("n=%d semilla=%d: VivosTrasRonda no coincide con las fichas", n, semilla)
			}
			if len(res.Finalistas) != min(6, n) {
				t.Fatalf("n=%d semilla=%d: %d finalistas", n, semilla, len(res.Finalistas))
			}

			// Los puestos son una permutación de 1..n y respetan el orden de las bases.
			puestos := slices.Clone(res.Puesto)
			slices.Sort(puestos)
			for i, p := range puestos {
				if p != i+1 {
					t.Fatalf("n=%d semilla=%d: puestos %v", n, semilla, res.Puesto)
				}
			}
			for a := 0; a < n; a++ {
				for b := 0; b < n; b++ {
					if res.Puesto[a] >= res.Puesto[b] {
						continue
					}
					if res.Fichas[a] < res.Fichas[b] {
						t.Fatalf("n=%d semilla=%d: %d quedó delante de %d con menos fichas", n, semilla, a, b)
					}
					if res.Fichas[a] == 0 && res.RondaQuiebra[a] < res.RondaQuiebra[b] {
						t.Fatalf("n=%d semilla=%d: %d quedó delante de %d y cayó antes", n, semilla, a, b)
					}
				}
			}
			if !res.EsFinalista(res.Primero()) || res.Puesto[res.Primero()] != 1 {
				t.Fatalf("n=%d semilla=%d: el primero no tiene el puesto 1", n, semilla)
			}
		}
	}
}

func TestLigaEsReproducible(t *testing.T) {
	bib := bibliotecaDePrueba()
	cfg := LigaDeLasBases()
	a := JugarLiga(cfg, campoDePrueba(20), bib, 11, 99)
	b := JugarLiga(cfg, campoDePrueba(20), bib, 11, 99)
	if !slices.Equal(a.Fichas, b.Fichas) || !slices.Equal(a.Finalistas, b.Finalistas) {
		t.Fatal("la misma semilla dio dos ligas distintas")
	}
	c := JugarLiga(cfg, campoDePrueba(20), bib, 11, 100)
	if slices.Equal(a.Fichas, c.Fichas) {
		t.Fatal("otra semilla dio la misma liga")
	}
}

func TestFinal(t *testing.T) {
	bib := bibliotecaDePrueba()
	ganadores := map[int]int{}
	for semilla := uint64(1); semilla <= 60; semilla++ {
		// Con tope corto gana el que va adelante; sin tope tiene que quedar uno.
		for _, tope := range []int{150, 0} {
			cfg := FinalDeLasBases(tope)
			ganador, manos := JugarFinal(cfg, campoDePrueba(6), bib, int(semilla)*13, semilla)
			if ganador < 0 || ganador >= 6 {
				t.Fatalf("ganador %d fuera de rango", ganador)
			}
			if manos < 1 || manos > cfg.TopeManos {
				t.Fatalf("la final jugó %d manos con tope %d", manos, cfg.TopeManos)
			}
			ganadores[ganador]++
		}
		g1, m1 := JugarFinal(FinalDeLasBases(500), campoDePrueba(6), bib, 5, semilla)
		g2, m2 := JugarFinal(FinalDeLasBases(500), campoDePrueba(6), bib, 5, semilla)
		if g1 != g2 || m1 != m2 {
			t.Fatal("la misma semilla dio dos finales distintas")
		}
	}
	if len(ganadores) < 2 {
		t.Fatalf("siempre gana el mismo asiento: %v", ganadores)
	}
}

func TestCampoArmar(t *testing.T) {
	g := estrategia.GenomaInicial()
	for _, campo := range Campos(g, g) {
		for _, n := range []int{1, 5, 14, 19, 29, 39} {
			rivales := campo.Armar(n)
			if len(rivales) != n {
				t.Errorf("campo %s: Armar(%d) devolvió %d", campo.Nombre, n, len(rivales))
			}
		}
		// Con muchos rivales tienen que aparecer todos los tipos.
		vistos := map[string]bool{}
		for _, r := range campo.Armar(39) {
			vistos[r.Nombre] = true
		}
		if len(vistos) != len(campo.Tipos) {
			t.Errorf("campo %s: con 39 rivales aparecen %d de %d tipos", campo.Nombre, len(vistos), len(campo.Tipos))
		}
	}
}
