package main

import (
	"bytes"
	"context"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aicspucp/pokerfight-arena-participantes/pkg/botsdk"

	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/estrategia"
)

func accion(id string, etapa botsdk.Etapa, tipo botsdk.TipoAccion) botsdk.AccionRegistrada {
	return botsdk.AccionRegistrada{IDJugador: id, Etapa: etapa, Accion: botsdk.Accion{Tipo: tipo}}
}

func estadoCon(idMano string, acciones ...botsdk.AccionRegistrada) botsdk.EstadoPublico {
	return botsdk.EstadoPublico{IDMano: idMano, Etapa: botsdk.PreFlop, HistorialAcciones: acciones}
}

func perfilDe(r *lecturaRivales, id string) estrategia.Perfil {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p := r.perfiles[id]; p != nil {
		return *p
	}
	return estrategia.Perfil{}
}

// El mismo estado (o uno más viejo) visto muchas veces no se cuenta dos veces.
func TestLecturaRivalesNoCuentaDosVeces(t *testing.T) {
	r := nuevaLecturaRivales("")
	a1 := accion("a", botsdk.PreFlop, botsdk.Raise)
	a2 := accion("b", botsdk.PreFlop, botsdk.Call)
	a3 := accion("a", botsdk.PreFlop, botsdk.Fold)
	a4 := accion("b", botsdk.Flop, botsdk.Bet)

	r.observar(estadoCon("m-1", a1), "yo")
	r.observar(estadoCon("m-1", a1), "yo")
	r.observar(estadoCon("m-1", a1, a2), "yo")
	r.observar(estadoCon("m-1", a1), "yo") // Un estado viejo que llega tarde.
	r.observar(estadoCon("m-1", a1, a2, a3, a4), "yo")
	r.observar(estadoCon("m-1", a1, a2, a3, a4), "yo")

	if p := perfilDe(r, "a"); p != (estrategia.Perfil{Agresivas: 1}) {
		t.Fatalf("perfil de a %+v; se esperaba 1 agresiva (el fold no cuenta)", p)
	}
	if p := perfilDe(r, "b"); p != (estrategia.Perfil{Agresivas: 1, Pasivas: 1}) {
		t.Fatalf("perfil de b %+v; se esperaba 1 agresiva y 1 pasiva", p)
	}
}

// Al cambiar de mano, el historial vuelve a empezar desde cero.
func TestLecturaRivalesCambiaDeMano(t *testing.T) {
	r := nuevaLecturaRivales("")
	r.observar(estadoCon("m-1", accion("a", botsdk.PreFlop, botsdk.Raise), accion("b", botsdk.PreFlop, botsdk.Call)), "yo")
	// La mano siguiente, con una sola acción: más corta que la anterior, pero nueva.
	r.observar(estadoCon("m-2", accion("a", botsdk.PreFlop, botsdk.AllIn)), "yo")
	if p := perfilDe(r, "a"); p.Agresivas != 2 {
		t.Fatalf("perfil de a %+v; se esperaban 2 agresivas", p)
	}
	// Mano nueva sin acciones (mano_inicio: las ciegas no son acciones).
	r.observar(estadoCon("m-3"), "yo")
	r.observar(estadoCon("m-3", accion("b", botsdk.PreFlop, botsdk.Call)), "yo")
	if p := perfilDe(r, "b"); p.Pasivas != 2 {
		t.Fatalf("perfil de b %+v; se esperaban 2 pasivas", p)
	}
}

// Las acciones propias no se anotan, aunque el id llegue en una llamada anterior.
func TestLecturaRivalesIgnoraLasPropias(t *testing.T) {
	r := nuevaLecturaRivales("")
	r.observar(estadoCon("m-1", accion("yo", botsdk.PreFlop, botsdk.Raise)), "yo")
	r.observar(estadoCon("m-1", accion("yo", botsdk.PreFlop, botsdk.Raise), accion("yo", botsdk.PreFlop, botsdk.Call)), "")
	if n := r.cantidad(); n != 0 {
		t.Fatalf("se anotaron %d perfiles con solo acciones propias", n)
	}
}

func TestLecturaDeUnDesconocidoEsCero(t *testing.T) {
	r := nuevaLecturaRivales("")
	if l := r.lectura(""); l != 0 {
		t.Fatalf("lectura de \"\" = %v", l)
	}
	if l := r.lectura("nadie"); l != 0 {
		t.Fatalf("lectura de un desconocido = %v", l)
	}
	r.observar(estadoCon("m-1", accion("a", botsdk.PreFlop, botsdk.Raise)), "yo")
	if l := r.lectura("a"); l <= 0 {
		t.Fatalf("lectura de un rival que subió = %v; se esperaba > 0", l)
	}
}

// Guardar y leer conserva los perfiles y la cuenta de la mano en curso.
func TestLecturaRivalesGuardaYLee(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "rivales.json")
	r := nuevaLecturaRivales(ruta)
	e := estadoCon("m-1", accion("a", botsdk.PreFlop, botsdk.Raise), accion("b", botsdk.PreFlop, botsdk.Call))
	r.observar(e, "yo")
	for i := 0; i < manosEntreGuardados-1; i++ {
		if err := r.manoFin(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(ruta); !os.IsNotExist(err) {
		t.Fatalf("guardó antes de tiempo (err %v)", err)
	}
	if err := r.manoFin(); err != nil {
		t.Fatal(err)
	}

	otro := nuevaLecturaRivales(ruta)
	if err := otro.leer(); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"a", "b"} {
		if perfilDe(otro, id) != perfilDe(r, id) {
			t.Fatalf("perfil de %s: guardado %+v, leído %+v", id, perfilDe(r, id), perfilDe(otro, id))
		}
	}
	// Retoma la cuenta: el mismo estado no se vuelve a contar tras el reinicio.
	otro.observar(e, "yo")
	if p := perfilDe(otro, "a"); p.Agresivas != 1 {
		t.Fatalf("tras retomar contó de nuevo: %+v", p)
	}

	// Sin ruta no se escribe nada, y leer un archivo que no existe no es error.
	if err := nuevaLecturaRivales("").guardar(); err != nil {
		t.Fatal(err)
	}
	if err := nuevaLecturaRivales(filepath.Join(t.TempDir(), "no-existe.json")).leer(); err != nil {
		t.Fatal(err)
	}
}

// decidir llena LecturaRival con el perfil del último agresor y lo anota en el
// detalle; funciona con un lince sin seguidor (como con -sin-archivos en pruebas).
func TestDecidirUsaLaLecturaDelUltimoAgresor(t *testing.T) {
	var salida, detalle bytes.Buffer
	b := botDePrueba(&salida)
	b.detalle.SetOutput(&detalle)
	b.azar = rand.New(rand.NewPCG(1, 2))

	turno := turnoPostflop(botsdk.Fold, botsdk.Call, botsdk.Raise, botsdk.AllIn)
	turno.Estado.HistorialAcciones = []botsdk.AccionRegistrada{
		accion("rival", botsdk.PreFlop, botsdk.Raise),
		accion("yo", botsdk.PreFlop, botsdk.Call),
		accion("rival", botsdk.Flop, botsdk.Bet),
	}
	b.Decidir(context.Background(), turno)
	if strings.Contains(salida.String(), "PÁNICO") {
		t.Fatalf("pánico: %s", salida.String())
	}
	if !strings.Contains(detalle.String(), "lectura +") {
		t.Fatalf("el detalle no trae la lectura: %s", detalle.String())
	}
	if p := perfilDe(b.rivales, "rival"); p.Agresivas != 2 || p.Pasivas != 0 {
		t.Fatalf("perfil del rival %+v; se esperaban 2 agresivas", p)
	}
}

func TestTextoLectura(t *testing.T) {
	if s := textoLectura(0); s != "" {
		t.Fatalf("lectura 0 escribe %q", s)
	}
	if s := textoLectura(0.3125); s != " · lectura +0,31" {
		t.Fatalf("texto %q", s)
	}
	if s := textoLectura(-0.2); s != " · lectura -0,20" {
		t.Fatalf("texto %q", s)
	}
}
