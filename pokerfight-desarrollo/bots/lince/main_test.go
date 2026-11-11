package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/aicspucp/pokerfight-arena-participantes/pkg/botsdk"

	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/torneo"
)

func botDePrueba(salida io.Writer) *lince {
	return &lince{
		contexto: torneo.Nuevo(torneo.ConfigCompetencia(), ""),
		eventos:  log.New(salida, "", 0),
		detalle:  log.New(io.Discard, "", 0),
	}
}

// turnoPostflop arma un turno después del flop: ahí decidir necesita el
// generador de azar, que en estas pruebas falta a propósito.
func turnoPostflop(validas ...botsdk.TipoAccion) botsdk.Turno {
	return botsdk.Turno{
		IDJugador: "yo",
		Yo:        botsdk.JugadorPublico{ID: "yo", Saldo: 9000, Activo: true, EnMano: true, EnTorneo: true},
		Mano:      botsdk.Mano{{Rango: botsdk.As, Palo: botsdk.Picas}, {Rango: 13, Palo: botsdk.Picas}},
		Estado: botsdk.EstadoPublico{
			IDMano:       "mesa-1-mano-1",
			Etapa:        botsdk.Flop,
			CiegaGrande:  100,
			Pozo:         600,
			Comunitarias: []botsdk.Carta{{Rango: 2, Palo: botsdk.Treboles}, {Rango: 7, Palo: botsdk.Corazones}, {Rango: 9, Palo: botsdk.Picas}},
			Jugadores: []botsdk.JugadorPublico{
				{ID: "yo", Saldo: 9000, Activo: true, EnMano: true, EnTorneo: true},
				{ID: "rival", Saldo: 9000, Activo: true, EnMano: true, EnTorneo: true, PosicionSilla: 1},
			},
		},
		Validas: validas,
		Timeout: 5 * time.Second,
	}
}

// Un pánico al decidir no puede tirar el proceso: se juega la acción segura.
func TestUnPanicoAlDecidirJuegaLaAccionSegura(t *testing.T) {
	var salida bytes.Buffer
	b := botDePrueba(&salida) // Sin azar: decidir va a entrar en pánico.

	if a := b.Decidir(context.Background(), turnoPostflop(botsdk.Check, botsdk.Bet, botsdk.AllIn)); a.Tipo != botsdk.Check {
		t.Fatalf("con check disponible jugó %s", a.Tipo)
	}
	if a := b.Decidir(context.Background(), turnoPostflop(botsdk.Fold, botsdk.Call, botsdk.Raise, botsdk.AllIn)); a.Tipo != botsdk.Fold {
		t.Fatalf("sin check disponible jugó %s", a.Tipo)
	}
	if !strings.Contains(salida.String(), "PÁNICO") {
		t.Fatal("el pánico no quedó en el log")
	}
}

// Un pánico en el seguimiento del torneo tampoco puede tirar el proceso.
func TestUnPanicoEnElSeguimientoNoSeEscapa(t *testing.T) {
	var salida bytes.Buffer
	b := botDePrueba(&salida)
	b.contexto = nil // Cualquier uso del contexto entra en pánico.
	estado := turnoPostflop().Estado
	b.ManoInicio(estado, botsdk.Mano{})
	b.ManoFin(botsdk.ResultadoMano{})
	b.EstadoActualizado(estado)
	if got := strings.Count(salida.String(), "PÁNICO"); got != 3 {
		t.Fatalf("se esperaban 3 pánicos anotados, hay %d", got)
	}
}

func TestCapturaElIDDelLogDelSDK(t *testing.T) {
	var visto string
	c := capturaID{destino: io.Discard, alVerID: func(id string) { visto = id }}
	registro := log.New(c, "sdk: ", log.LstdFlags)

	registro.Printf("la mesa está saturada; reintento 1 de 3 en 200ms")
	if visto != "" {
		t.Fatalf("tomó un id de una línea que no lo trae: %q", visto)
	}
	// El formato exacto de botsdk.Correr.
	registro.Printf("conectado a %s como %s (silla %d)", "32.201.29.142:9000", "j-a1b2c3", 4)
	if visto != "j-a1b2c3" {
		t.Fatalf("id capturado %q; se esperaba j-a1b2c3", visto)
	}
}

func TestEsperaDeReconexion(t *testing.T) {
	// Fallos seguidos: 1, 2, 3, 4, 5, 5, 5 segundos.
	espera := time.Duration(0)
	var serie []time.Duration
	for i := 0; i < 7; i++ {
		espera = siguienteEspera(espera, 0)
		serie = append(serie, espera)
	}
	quiero := []time.Duration{1, 2, 3, 4, 5, 5, 5}
	for i := range quiero {
		if serie[i] != quiero[i]*time.Second {
			t.Fatalf("esperas %v; se esperaba 1,2,3,4,5,5,5 s", serie)
		}
	}
	// Tras una conexión que duró, vuelve al mínimo.
	if e := siguienteEspera(espera, 10*time.Minute); e != time.Second {
		t.Fatalf("tras una conexión larga la espera es %s; se esperaba 1s", e)
	}
}
