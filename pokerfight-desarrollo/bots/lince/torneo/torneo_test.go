package torneo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aicspucp/pokerfight-arena-participantes/pkg/botsdk"
)

const yo = "j-yo0000"

// reloj es un reloj de mentira que avanza a mano.
type reloj struct{ t time.Time }

func (r *reloj) ahora() time.Time              { return r.t }
func (r *reloj) avanzar(d time.Duration)       { r.t = r.t.Add(d) }
func nuevoReloj() *reloj                       { return &reloj{t: time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)} }
func conReloj(c *Contexto, r *reloj) *Contexto { c.Ahora = r.ahora; return c }

// mesa arma un estado público: stacks[id] son las fichas de cada sentado.
func mesa(idMano string, stacks map[string]int64) botsdk.EstadoPublico {
	e := botsdk.EstadoPublico{IDMano: idMano, CiegaChica: 50, CiegaGrande: 100}
	silla := 0
	for _, id := range ordenados(stacks) {
		e.Jugadores = append(e.Jugadores, botsdk.JugadorPublico{
			ID: id, Saldo: stacks[id], EnTorneo: stacks[id] > 0, EnMano: stacks[id] > 0, Activo: stacks[id] > 0, PosicionSilla: silla,
		})
		silla++
	}
	return e
}

func ordenados(m map[string]int64) []string {
	ids := make([]string, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	for i := range ids {
		for j := i + 1; j < len(ids); j++ {
			if ids[j] < ids[i] {
				ids[i], ids[j] = ids[j], ids[i]
			}
		}
	}
	return ids
}

func rivales(prefijo string, n int, fichas int64) map[string]int64 {
	m := map[string]int64{yo: fichas}
	for i := 0; i < n; i++ {
		m[fmt.Sprintf("j-%s%d", prefijo, i)] = fichas
	}
	return m
}

// jugarMesa juega `manos` manos en una mesa, sin que cambien las fichas.
func jugarMesa(t *testing.T, c *Contexto, r *reloj, idMesa string, stacks map[string]int64, manos int) {
	t.Helper()
	for n := 1; n <= manos; n++ {
		id := fmt.Sprintf("%s-mano-%d", idMesa, n)
		nueva, _ := c.ManoInicio(mesa(id, stacks))
		if nueva != (n == 1) {
			t.Fatalf("mesa %s mano %d: nuevaMesa = %v", idMesa, n, nueva)
		}
		r.avanzar(3 * time.Second)
		if err := c.ManoFin(botsdk.ResultadoMano{IDMano: id}); err != nil {
			t.Fatalf("ManoFin: %v", err)
		}
		r.avanzar(time.Second)
	}
}

func TestCuentaRondasYManos(t *testing.T) {
	r := nuevoReloj()
	c := conReloj(Nuevo(ConfigCompetencia(), ""), r)
	c.Identificar(yo)

	jugarMesa(t, c, r, "r1-m3", rivales("a", 4, 10000), 50)
	jugarMesa(t, c, r, "r2-m1", rivales("b", 4, 12000), 37)

	e := c.Estado()
	if e.Ronda != 2 || e.ManosRonda != 37 || e.ManosTotales != 87 {
		t.Fatalf("ronda %d, manos de la ronda %d, totales %d; se esperaba 2, 37, 87", e.Ronda, e.ManosRonda, e.ManosTotales)
	}
	if c.Fase() != Clasificatoria {
		t.Fatalf("fase %s; se esperaba clasificatoria", c.Fase())
	}
	if e.Rondas[0].StackInicio != 10000 || e.Rondas[0].StackFin != 12000 || e.Rondas[0].Manos != 50 {
		t.Fatalf("ronda 1 mal registrada: %+v", e.Rondas[0])
	}
	if a := e.Alias["j-a0"]; a.Mesas != 1 || a.UltimaRonda != 1 || a.Stack != 10000 {
		t.Fatalf("alias j-a0 mal registrado: %+v", a)
	}
}

// Con pocos bots vivos, dos rondas seguidas repiten los mismos rivales: la
// ronda nueva se tiene que notar igual.
func TestMismosRivalesDosRondas(t *testing.T) {
	stacks := rivales("a", 3, 9000)

	t.Run("por el id de mesa", func(t *testing.T) {
		r := nuevoReloj()
		c := conReloj(Nuevo(ConfigCompetencia(), ""), r)
		c.Identificar(yo)
		jugarMesa(t, c, r, "r4-m1", stacks, 20)
		jugarMesa(t, c, r, "r5-m1", stacks, 5)
		if c.Estado().Ronda != 2 {
			t.Fatalf("ronda %d; se esperaba 2", c.Estado().Ronda)
		}
	})

	t.Run("por el número de mano que vuelve atrás", func(t *testing.T) {
		r := nuevoReloj()
		c := conReloj(Nuevo(ConfigCompetencia(), ""), r)
		c.Identificar(yo)
		jugarMesa(t, c, r, "mesa-1", stacks, 20)
		jugarMesa(t, c, r, "mesa-1", stacks, 5)
		if c.Estado().Ronda != 2 {
			t.Fatalf("ronda %d; se esperaba 2", c.Estado().Ronda)
		}
	})

	t.Run("con id de mano opaco, por la cuenta de 50 manos", func(t *testing.T) {
		r := nuevoReloj()
		c := conReloj(Nuevo(ConfigCompetencia(), ""), r)
		c.Identificar(yo)
		for n := 0; n < 51; n++ {
			nueva, _ := c.ManoInicio(mesa(fmt.Sprintf("x%04d", n), stacks))
			if nueva != (n == 0 || n == 50) {
				t.Fatalf("mano %d: nuevaMesa = %v", n, nueva)
			}
			c.ManoFin(botsdk.ResultadoMano{})
			r.avanzar(time.Second)
		}
	})

	t.Run("con id de mano opaco, por la pausa", func(t *testing.T) {
		r := nuevoReloj()
		c := conReloj(Nuevo(ConfigCompetencia(), ""), r)
		c.Identificar(yo)
		c.ManoInicio(mesa("aaa", stacks))
		c.ManoFin(botsdk.ResultadoMano{})
		r.avanzar(5 * time.Second)
		if nueva, _ := c.ManoInicio(mesa("bbb", stacks)); nueva {
			t.Fatal("5 s de pausa no es una mesa nueva")
		}
		c.ManoFin(botsdk.ResultadoMano{})
		r.avanzar(2 * time.Minute)
		if nueva, _ := c.ManoInicio(mesa("ccc", stacks)); !nueva {
			t.Fatal("2 min de pausa con id opaco tiene que tomarse como mesa nueva")
		}
	})
}

// Una desconexión larga en la misma mesa no es una ronda nueva.
func TestPausaEnLaMismaMesaNoEsRondaNueva(t *testing.T) {
	r := nuevoReloj()
	c := conReloj(Nuevo(ConfigCompetencia(), ""), r)
	c.Identificar(yo)
	stacks := rivales("a", 4, 10000)
	jugarMesa(t, c, r, "r1-m1", stacks, 10)
	r.avanzar(3 * time.Minute)
	if nueva, motivo := c.ManoInicio(mesa("r1-m1-mano-14", stacks)); nueva {
		t.Fatalf("se tomó como mesa nueva: %s", motivo)
	}
}

func TestDetectaLaFinal(t *testing.T) {
	r := nuevoReloj()
	c := conReloj(Nuevo(ConfigCompetencia(), ""), r)
	c.Identificar(yo)

	// Ronda 1 en una mesa de 6: todos con 10 000 y NO es la final.
	jugarMesa(t, c, r, "r1-m1", rivales("a", 5, 10000), 50)
	if c.Fase() != Clasificatoria {
		t.Fatal("la primera mesa de 6 se tomó como final")
	}
	jugarMesa(t, c, r, "r2-m1", rivales("b", 4, 31000), 50)
	if c.Fase() != Clasificatoria {
		t.Fatal("la ronda 2 se tomó como final")
	}

	inicioFinal := r.ahora()
	nueva, motivo := c.ManoInicio(mesa("final-mano-1", rivales("f", 5, 10000)))
	if !nueva || c.Fase() != Final {
		t.Fatalf("no detectó la final (nueva=%v, fase=%s, motivo=%q)", nueva, c.Fase(), motivo)
	}
	if !strings.Contains(motivo, "FINAL") {
		t.Fatalf("el motivo no avisa de la final: %q", motivo)
	}
	e := c.Estado()
	if !e.InicioFinal.Equal(inicioFinal) {
		t.Fatalf("inicio de la final %v; se esperaba %v", e.InicioFinal, inicioFinal)
	}
	// Las fichas de la final no son con las que terminó la clasificatoria.
	if e.Rondas[1].StackFin == 10000 {
		t.Fatal("la última ronda de la clasificatoria quedó con las fichas de la final")
	}
}

func TestMesaDeCincoConStackInicialNoEsFinal(t *testing.T) {
	r := nuevoReloj()
	c := conReloj(Nuevo(ConfigCompetencia(), ""), r)
	c.Identificar(yo)
	jugarMesa(t, c, r, "r1-m1", rivales("a", 4, 10000), 50)
	// Servidor de pruebas: cada ronda todos vuelven a 10 000, en mesas de 5.
	jugarMesa(t, c, r, "r2-m1", rivales("b", 4, 10000), 50)
	if c.Fase() != Clasificatoria {
		t.Fatal("una mesa de 5 se tomó como final")
	}
}

func TestFaseForzada(t *testing.T) {
	cfg := ConfigCompetencia()
	cfg.FaseForzada = Clasificatoria
	r := nuevoReloj()
	c := conReloj(Nuevo(cfg, ""), r)
	c.Identificar(yo)
	jugarMesa(t, c, r, "r1-m1", rivales("a", 5, 10000), 50)
	jugarMesa(t, c, r, "r2-m1", rivales("b", 5, 10000), 50)
	if c.Fase() != Clasificatoria {
		t.Fatal("la fase forzada no se respetó")
	}
}

// Lo importante de todo el paquete: matar el proceso y relanzarlo no pierde la
// cuenta.
func TestRetomaElEstadoDeDisco(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "estado.json")
	r := nuevoReloj()
	c := conReloj(Nuevo(ConfigCompetencia(), ruta), r)
	c.Identificar(yo)
	jugarMesa(t, c, r, "r1-m1", rivales("a", 4, 10000), 50)
	jugarMesa(t, c, r, "r2-m4", rivales("b", 4, 8000), 12)

	// Proceso nuevo, unos segundos después.
	r.avanzar(20 * time.Second)
	c2 := reabrir(ruta, r)
	c2.Identificar(yo)
	if e := c2.Estado(); e.Ronda != 2 || e.ManosRonda != 12 || e.Stack != 8000 {
		t.Fatalf("retomó ronda %d, mano %d, stack %d; se esperaba 2, 12, 8000", e.Ronda, e.ManosRonda, e.Stack)
	}
	// Sigue en la misma mesa: no es ronda nueva.
	if nueva, motivo := c2.ManoInicio(mesa("r2-m4-mano-15", rivales("b", 4, 8000))); nueva {
		t.Fatalf("tras el reinicio tomó la misma mesa como nueva: %s", motivo)
	}
	if c2.Estado().Ronda != 2 {
		t.Fatalf("ronda %d tras el reinicio; se esperaba 2", c2.Estado().Ronda)
	}
}

// reabrir simula un proceso nuevo que lee el estado de disco.
func reabrir(ruta string, r *reloj) *Contexto {
	c := conReloj(Nuevo(ConfigCompetencia(), ruta), r)
	c.Retomar()
	return c
}

func TestEstadoDeOtroTorneoSeDescarta(t *testing.T) {
	t.Run("otro jugador", func(t *testing.T) {
		ruta := filepath.Join(t.TempDir(), "estado.json")
		r := nuevoReloj()
		c := conReloj(Nuevo(ConfigCompetencia(), ruta), r)
		c.Identificar("j-viejo")
		jugarMesa(t, c, r, "r1-m1", map[string]int64{"j-viejo": 10000, "j-a": 10000, "j-b": 10000}, 30)

		c2 := reabrir(ruta, r)
		c2.Identificar(yo)
		if e := c2.Estado(); e.Ronda != 0 || e.IDJugador != yo {
			t.Fatalf("no arrancó de cero: %+v", e)
		}
		if _, err := os.Stat(ruta); !os.IsNotExist(err) {
			t.Fatal("el estado del otro torneo sigue en su lugar")
		}
	})

	t.Run("vencido", func(t *testing.T) {
		ruta := filepath.Join(t.TempDir(), "estado.json")
		r := nuevoReloj()
		c := conReloj(Nuevo(ConfigCompetencia(), ruta), r)
		c.Identificar(yo)
		jugarMesa(t, c, r, "r1-m1", rivales("a", 4, 10000), 30)

		r.avanzar(7 * time.Hour)
		c2 := reabrir(ruta, r)
		if c2.Estado().Ronda != 0 {
			t.Fatal("retomó un estado de hace 7 horas")
		}
	})

	t.Run("ilegible", func(t *testing.T) {
		ruta := filepath.Join(t.TempDir(), "estado.json")
		if err := os.WriteFile(ruta, []byte("{esto no es json"), 0o644); err != nil {
			t.Fatal(err)
		}
		r := nuevoReloj()
		c := reabrir(ruta, r)
		if c.Estado().Ronda != 0 {
			t.Fatal("no arrancó de cero con un archivo ilegible")
		}
	})
}

func TestStacksAlCerrarLaMano(t *testing.T) {
	r := nuevoReloj()
	c := conReloj(Nuevo(ConfigCompetencia(), ""), r)
	c.Identificar(yo)
	c.ManoInicio(mesa("r1-m1-mano-1", map[string]int64{yo: 10000, "j-a": 10000, "j-b": 10000}))
	// j-a va all-in y pierde contra nosotros; j-b se retiró habiendo puesto 100.
	c.Observar(mesa("r1-m1-mano-1", map[string]int64{yo: 0, "j-a": 0, "j-b": 9900}))
	c.ManoFin(botsdk.ResultadoMano{Repartos: []botsdk.Reparto{{IDJugador: yo, Monto: 20100}}})

	e := c.Estado()
	if a := e.Alias["j-a"]; !a.Eliminado || a.Stack != 0 {
		t.Fatalf("j-a tendría que figurar eliminado: %+v", a)
	}
	if a := e.Alias["j-b"]; a.Eliminado || a.Stack != 9900 {
		t.Fatalf("j-b mal registrado: %+v", a)
	}
	if e.Rondas[0].StackFin != 20100 {
		t.Fatalf("nuestro stack estimado %d; se esperaba 20100", e.Rondas[0].StackFin)
	}
}

func TestPartirIDMano(t *testing.T) {
	casos := []struct {
		id     string
		mesa   string
		numero int
	}{
		{"mesa-1-mano-7", "mesa-1", 7},
		{"r3-m12-mano-50", "r3-m12", 50},
		{"mano-3", "", 0},
		{"abc123", "", 0},
		{"mesa-1-mano-x", "", 0},
		{"", "", 0},
	}
	for _, c := range casos {
		if mesa, n := partirIDMano(c.id); mesa != c.mesa || n != c.numero {
			t.Errorf("partirIDMano(%q) = (%q, %d); se esperaba (%q, %d)", c.id, mesa, n, c.mesa, c.numero)
		}
	}
}

// En Windows no se puede reemplazar un archivo que otro proceso tiene abierto.
// Guardar tiene que reintentar en vez de rendirse: la prueba abre el archivo
// unos milisegundos, como haría un antivirus.
func TestGuardarReintentaSiElArchivoEstaOcupado(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "estado.json")
	c := Nuevo(ConfigCompetencia(), ruta)
	c.Identificar(yo)
	if err := c.Guardar(); err != nil {
		t.Fatal(err)
	}

	ocupado, err := os.Open(ruta)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(12 * time.Millisecond)
		ocupado.Close()
	}()

	c.ManoInicio(mesa("r1-m1-mano-1", rivales("a", 4, 10000)))
	if err := c.Guardar(); err != nil {
		t.Fatalf("no guardó con el archivo ocupado un momento: %v", err)
	}
	c2 := Nuevo(ConfigCompetencia(), ruta)
	if !c2.Retomar() || c2.Estado().Ronda != 1 {
		t.Fatalf("lo guardado no es lo último: ronda %d", c2.Estado().Ronda)
	}
}
