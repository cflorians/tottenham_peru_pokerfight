package main

import (
	"slices"
	"testing"

	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/entrenamiento"
)

func TestArmarCampoParejoYDeterminista(t *testing.T) {
	tipos := []string{"darwin", "espia", "estadistico", "base", "conservador", "aleatorio", "novato"}
	campo := armarCampo(20, tipos)
	if len(campo) != 20 || !campo[0].EsLince || campo[0].ID != idLince {
		t.Fatalf("campo mal armado: %+v", campo[:1])
	}
	cuenta := map[string]int{}
	ids := map[string]bool{}
	for _, p := range campo[1:] {
		cuenta[p.Tipo]++
		if ids[p.ID] {
			t.Fatalf("id repetido %s", p.ID)
		}
		ids[p.ID] = true
	}
	// 19 entre 7 tipos: cinco de 3 y dos de 2.
	for i, tipo := range tipos {
		quiero := 2
		if i < 5 {
			quiero = 3
		}
		if cuenta[tipo] != quiero {
			t.Errorf("%s: %d instancias, quería %d", tipo, cuenta[tipo], quiero)
		}
	}
	if !ids["darwin-3"] || ids["novato-3"] {
		t.Errorf("numeración inesperada: %v", ids)
	}
	if !slices.Equal(campo, armarCampo(20, tipos)) {
		t.Error("el campo no es determinista")
	}
}

func TestOrdenarClasificacion(t *testing.T) {
	fichas := []int64{0, 25000, 0, 0, 25000, 50000}
	quiebras := []marcaQuiebra{
		{Ronda: 3, Mano: 10, StackAntes: 500},
		{},
		{Ronda: 5, Mano: 2, StackAntes: 100},
		{Ronda: 5, Mano: 2, StackAntes: 900}, // Misma mano que el 2: empezó con más.
		{},
		{},
	}
	sorteo := []int{0, 5, 1, 2, 3, 4} // Entre 1 y 4 (empate en fichas) va primero el 4.
	got := ordenarClasificacion(fichas, quiebras, sorteo)
	want := []int{5, 4, 1, 3, 2, 0}
	if !slices.Equal(got, want) {
		t.Fatalf("orden %v, quería %v", got, want)
	}
}

func TestMesasCubrenATodos(t *testing.T) {
	for n := 2; n <= 40; n++ {
		suma := 0
		for _, tam := range entrenamiento.TamanosMesas(n) {
			if tam < 2 || tam > 6 {
				t.Fatalf("n=%d: mesa de %d", n, tam)
			}
			suma += tam
		}
		if suma != n {
			t.Fatalf("n=%d: las mesas suman %d", n, suma)
		}
	}
}

func TestMezclarDistingue(t *testing.T) {
	if mezclar(1, 2) == mezclar(2, 1) || mezclar(1) == mezclar(2) || mezclar(7, 3) != mezclar(7, 3) {
		t.Fatal("mezclar no separa bien las semillas")
	}
}
