package main

import (
	"fmt"
	"slices"
)

// Lógica pura del orquestador: armar el campo y ordenar la clasificación. Va
// aparte para poder probarla sin lanzar procesos.

// idLince es el id con el que la mesa conoce a nuestro bot en todas las mesas
// de la liga. Es lo que mantiene su identidad entre rondas.
const idLince = "lince"

// participante es un inscrito en la liga.
type participante struct {
	ID      string `json:"id"`
	Tipo    string `json:"tipo"` // "lince" o el tipo de bot del kit.
	EsLince bool   `json:"es_lince"`
}

// armarCampo arma los n participantes: Lince en el índice 0 y n-1 instancias
// de los tipos dados, repartidas de forma pareja y determinista (en el orden
// de la lista, de a una por tipo). Cada instancia tiene un id propio: el
// tipo y su número ("darwin-2").
func armarCampo(n int, tipos []string) []participante {
	if n < 1 {
		return nil
	}
	campo := []participante{{ID: idLince, Tipo: "lince", EsLince: true}}
	if len(tipos) == 0 {
		return campo
	}
	cuenta := map[string]int{}
	for i := 0; i < n-1; i++ {
		tipo := tipos[i%len(tipos)]
		cuenta[tipo]++
		campo = append(campo, participante{ID: fmt.Sprintf("%s-%d", tipo, cuenta[tipo]), Tipo: tipo})
	}
	return campo
}

// marcaQuiebra es cuándo se quedó sin fichas un participante.
type marcaQuiebra struct {
	Ronda      int   // Desde 1; 0 = nunca quebró.
	Mano       int   // Número de mano dentro de la mesa (desde 1).
	StackAntes int64 // Fichas con que empezó esa mano.
}

// ordenarClasificacion devuelve los índices de los participantes del mejor al
// peor, con el mismo criterio que entrenamiento.JugarLiga: más fichas primero;
// entre eliminados, mejor el que cayó más tarde (ronda y número de mano) y, en
// la misma mano, el que la empezó con más fichas; el empate restante lo rompe
// el sorteo (menor valor de sorteo primero).
func ordenarClasificacion(fichas []int64, quiebras []marcaQuiebra, sorteo []int) []int {
	orden := make([]int, len(fichas))
	for i := range orden {
		orden[i] = i
	}
	slices.SortFunc(orden, func(a, b int) int {
		qa, qb := quiebras[a], quiebras[b]
		switch {
		case fichas[a] != fichas[b]:
			return compararDesc(fichas[a], fichas[b])
		case qa.Ronda != qb.Ronda:
			return qb.Ronda - qa.Ronda
		case qa.Mano != qb.Mano:
			return qb.Mano - qa.Mano
		case qa.StackAntes != qb.StackAntes:
			return compararDesc(qa.StackAntes, qb.StackAntes)
		}
		return sorteo[a] - sorteo[b]
	})
	return orden
}

func compararDesc(a, b int64) int {
	switch {
	case a > b:
		return -1
	case a < b:
		return 1
	}
	return 0
}

// mezclar combina números en una semilla (splitmix64 encadenado). Sirve para
// derivar la semilla de cada ronda y de cada mesa de la semilla de la liga.
func mezclar(valores ...uint64) uint64 {
	h := uint64(0x9E3779B97F4A7C15)
	for _, v := range valores {
		h ^= v + 0x9E3779B97F4A7C15 + (h << 6) + (h >> 2)
		h ^= h >> 30
		h *= 0xBF58476D1CE4E5B9
		h ^= h >> 27
		h *= 0x94D049BB133111EB
		h ^= h >> 31
	}
	return h
}
