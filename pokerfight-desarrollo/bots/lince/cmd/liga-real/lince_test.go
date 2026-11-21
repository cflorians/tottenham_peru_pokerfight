package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/torneo"
	"github.com/aicspucp/pokerfight-arena-participantes/internal/mesa"
	"github.com/aicspucp/pokerfight-arena-participantes/internal/protocolo"
)

func escribirEstadoPrueba(t *testing.T, ruta string, estado torneo.Estado) {
	t.Helper()
	datos, err := json.Marshal(estado)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ruta, datos, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestChequeoDistingueManoDefinitivaDeTemporal(t *testing.T) {
	const manoAnterior = "L01-R03-M1-mano-33"
	const manoFinal = "L01-R03-M1-mano-34"
	dir := t.TempDir()
	ruta := filepath.Join(dir, "estado.json")
	estado := torneo.Estado{
		IDJugador: idLince, Ronda: 1, Fase: torneo.Clasificatoria,
		UltimaMano: manoAnterior, UltimaManoFin: time.Unix(1, 0),
		Rondas: []torneo.Ronda{{Fase: torneo.Clasificatoria, StackInicio: 10000, Jugadores: 5}},
	}
	escribirEstadoPrueba(t, ruta, estado)
	estado.UltimaMano = manoFinal
	escribirEstadoPrueba(t, ruta+".tmp", estado)
	jugadas := []mesaDeLince{{Jugadores: 5, FichasLince: 10000, UltimaMano: manoFinal}}

	chequeo := chequear(dir, jugadas)
	if chequeo.Coincide || chequeo.UltimaManoGuardada != manoAnterior || chequeo.UltimaManoEsperada != manoFinal {
		t.Fatalf("el temporal ocultó el desfase: %+v", chequeo)
	}
	if err := os.Rename(ruta+".tmp", ruta); err != nil {
		t.Fatal(err)
	}
	chequeo = chequear(dir, jugadas)
	if !chequeo.Coincide || chequeo.UltimaManoGuardada != manoFinal {
		t.Fatalf("el cierre definitivo no se reconoció: %+v", chequeo)
	}
}

func TestEsperarGuardadoHastaPublicacion(t *testing.T) {
	const manoFinal = "L01-R03-M1-mano-34"
	dir := t.TempDir()
	ruta := filepath.Join(dir, "estado.json")
	escribirEstadoPrueba(t, ruta+".tmp", torneo.Estado{
		IDJugador: idLince, UltimaMano: manoFinal, UltimaManoFin: time.Unix(1, 0),
	})
	proceso := &procesoLince{terminado: make(chan struct{})}
	ctx, cancelar := context.WithCancel(context.Background())
	defer cancelar()
	resultado := make(chan bool, 1)
	go func() { resultado <- proceso.esperarGuardado(ctx, dir, manoFinal) }()
	select {
	case <-resultado:
		t.Fatal("la espera terminó con el estado todavía en el archivo temporal")
	case <-time.After(50 * time.Millisecond):
	}

	if err := os.Rename(ruta+".tmp", ruta); err != nil {
		t.Fatal(err)
	}
	select {
	case guardada := <-resultado:
		if !guardada {
			t.Fatal("la espera no vio la publicación de la última mano")
		}
	case <-time.After(time.Second):
		t.Fatal("la espera no terminó al publicarse el estado")
	}
}

func TestEsperarGuardadoSeInterrumpe(t *testing.T) {
	for _, caso := range []string{"contexto cancelado", "proceso terminado"} {
		t.Run(caso, func(t *testing.T) {
			proceso := &procesoLince{terminado: make(chan struct{})}
			ctx, cancelar := context.WithCancel(context.Background())
			defer cancelar()
			if caso == "contexto cancelado" {
				cancelar()
			} else {
				close(proceso.terminado)
			}
			dir := t.TempDir()
			resultado := make(chan bool, 1)
			go func() { resultado <- proceso.esperarGuardado(ctx, dir, "mano-1") }()
			select {
			case guardada := <-resultado:
				if guardada {
					t.Fatal("la espera dio por guardada una mano que falta")
				}
			case <-time.After(time.Second):
				t.Fatal("la espera siguió tras la interrupción")
			}
		})
	}
}

func TestObservadorConservaUltimaManoDeLinceEliminado(t *testing.T) {
	const manoEliminacion = "L01-R03-M1-mano-34"
	observador := observadorMesa{
		saldos: map[string]int64{idLince: 120}, quiebras: map[string]marcaQuiebra{},
	}
	estado := protocolo.EstadoPublico{Jugadores: []protocolo.JugadorPublico{{ID: idLince, EnMano: true}}}
	observador.ManoTerminada(mesa.ManoJugada{IDMano: manoEliminacion, Estado: estado})
	estado.Jugadores[0].EnMano = false
	observador.ManoTerminada(mesa.ManoJugada{IDMano: "L01-R03-M1-mano-35", Estado: estado})
	if observador.ultimaManoLince != manoEliminacion {
		t.Fatalf("mano de Lince %q, se esperaba %q", observador.ultimaManoLince, manoEliminacion)
	}
}
