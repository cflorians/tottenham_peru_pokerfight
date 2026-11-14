// Lince: nuestro bot para PokerFight.
//
// Parte de Gato Darwin (bots/darwin): una política cuyos números salen de un
// algoritmo genético, con la fuerza de la mano estimada por Monte Carlo. Sobre
// eso agrega lo que hace falta para jugar el campeonato entero sin caerse:
//
//   - Reconexión: si se corta la conexión, vuelve a entrar con el mismo token
//     y retoma su silla y sus fichas (el SDK no lo hace solo).
//   - Nunca se cae por un error propio: ante un pánico al decidir, juega la
//     acción segura y lo deja anotado.
//   - Contexto de torneo guardado en disco (paquete torneo): ronda, manos,
//     fase y fichas de cada rival. Sobrevive a un reinicio del proceso.
//   - Log a archivo de cada decisión, cambio de mesa y reconexión.
//
// Uso:
//
//	lince -addr servidor:9000 -token <token>          # torneo: reconecta siempre
//	lince -addr localhost:9000 -token x -una-partida  # arena local: termina con la mesa
package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/aicspucp/pokerfight-arena-participantes/pkg/botsdk"

	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/estrategia"
	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/torneo"
)

// genomaIncluido es el genoma con el que juega el bot. Se incluye en el
// binario para que funcione desde cualquier directorio.
//
//go:embed genoma.json
var genomaIncluido []byte

// genomaFinalIncluido es el genoma con el que juega la mesa final: ahí no hay
// premio por sobrevivir, solo por ganar, y se juega distinto.
//
//go:embed genoma_final.json
var genomaFinalIncluido []byte

// tablaPreflop es la fuerza de las 169 manos iniciales, precalculada por el
// entrenamiento. Sin ella, cada proceso la recalcularía de cero.
//
//go:embed preflop.json
var tablaPreflop []byte

var (
	direccion   = flag.String("addr", "localhost:9000", "dirección de la mesa (host:puerto)")
	token       = flag.String("token", "", "token del equipo; si falta, se toma de la variable POKERFIGHT_TOKEN. En una mesa local, cualquier texto")
	rutaGenoma  = flag.String("genoma", "", "genoma.json a usar en lugar del incluido en el binario")
	rutaFinal   = flag.String("genoma-final", "", "genoma de la mesa final a usar en lugar del incluido; con -genoma y sin este flag, la final se juega con el de -genoma")
	semilla     = flag.Uint64("semilla", 7, "semilla de los faroles y del Monte Carlo")
	muestras    = flag.Int("muestras", 1500, "muestras de Monte Carlo por decisión después del flop")
	silencioso  = flag.Bool("silencioso", false, "no imprime cada decisión en la terminal (al archivo de log van igual)")
	dirDatos    = flag.String("datos", "lince-datos", "carpeta donde se guardan el estado del torneo y el log")
	sinArchivos = flag.Bool("sin-archivos", false, "no escribe estado ni log en disco (para torneos de prueba en la arena local)")
	unaPartida  = flag.Bool("una-partida", false, "termina cuando la mesa cierra la conexión, en vez de reconectar (para la arena local)")
	faseForzada = flag.String("fase", "auto", "fase del torneo: auto (se deduce), clasificatoria o final. En el servidor de pruebas, usar clasificatoria")
)

const nombreBot = "lince"

type lince struct {
	genoma   estrategia.Genoma // Clasificatoria.
	final    estrategia.Genoma // Mesa final.
	azar     *rand.Rand
	contexto *torneo.Contexto

	eventos *log.Logger // Lo que importa siempre: conexión, mesas, errores. Terminal y archivo.
	detalle *log.Logger // Cada decisión. Archivo, y terminal salvo -silencioso.
}

func (*lince) Nombre() string { return nombreBot }

// preparar carga el genoma, abre el log y retoma el estado del torneo. Se
// llama una sola vez, antes del bucle de conexión: no se implementa
// botsdk.Inicializable porque el SDK lo llamaría de nuevo en cada reconexión.
func (b *lince) preparar() (cerrar func(), err error) {
	cerrar = func() {}
	salidaEventos, salidaDetalle := io.Writer(os.Stderr), io.Writer(os.Stderr)
	if *silencioso {
		salidaDetalle = io.Discard
	}

	rutaEstado := ""
	if !*sinArchivos {
		if err := os.MkdirAll(*dirDatos, 0o755); err != nil {
			return cerrar, fmt.Errorf("no se pudo crear la carpeta de datos %s: %w", *dirDatos, err)
		}
		rutaLog := filepath.Join(*dirDatos, "lince-"+time.Now().Format("20060102")+".log")
		archivo, err := os.OpenFile(rutaLog, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return cerrar, fmt.Errorf("no se pudo abrir el log %s: %w", rutaLog, err)
		}
		cerrar = func() { archivo.Close() }
		salidaEventos = io.MultiWriter(os.Stderr, archivo)
		salidaDetalle = io.MultiWriter(salidaDetalle, archivo)
		rutaEstado = filepath.Join(*dirDatos, "estado.json")
	}
	b.eventos = log.New(salidaEventos, "", log.LstdFlags|log.Lmicroseconds)
	b.detalle = log.New(salidaDetalle, "", log.LstdFlags|log.Lmicroseconds)

	datos := genomaIncluido
	if *rutaGenoma != "" {
		leidos, err := os.ReadFile(*rutaGenoma)
		if err != nil {
			return cerrar, err
		}
		datos = leidos
	}
	genoma, archivoGenoma, err := estrategia.LeerGenoma(datos)
	if err != nil {
		return cerrar, err
	}
	b.genoma = genoma

	// El genoma de la final: el del flag, o el incluido. Si se pasó -genoma
	// sin -genoma-final (pruebas con un genoma suelto), se usa ese en las dos
	// fases, para medir un solo genoma por vez.
	b.final = genoma
	if *rutaFinal != "" || *rutaGenoma == "" {
		datosFinal := genomaFinalIncluido
		if *rutaFinal != "" {
			if datosFinal, err = os.ReadFile(*rutaFinal); err != nil {
				return cerrar, err
			}
		}
		if b.final, _, err = estrategia.LeerGenoma(datosFinal); err != nil {
			return cerrar, fmt.Errorf("genoma de la final: %w", err)
		}
	}

	var tabla map[string]float64
	if err := json.Unmarshal(tablaPreflop, &tabla); err != nil {
		return cerrar, fmt.Errorf("preflop.json ilegible: %w", err)
	}
	estrategia.CargarTablaPreflop(tabla)
	// Una fuente propia y sembrada en vez de la global de math/rand, que
	// cambia en cada proceso: así una partida del bot es reproducible.
	b.azar = rand.New(rand.NewPCG(*semilla, *semilla^0xDA8F1))

	cfg := torneo.ConfigCompetencia()
	switch *faseForzada {
	case "auto", "":
	case string(torneo.Clasificatoria):
		cfg.FaseForzada = torneo.Clasificatoria
	case string(torneo.Final):
		cfg.FaseForzada = torneo.Final
	default:
		return cerrar, fmt.Errorf("-fase %q no existe: usar auto, clasificatoria o final", *faseForzada)
	}
	b.contexto = torneo.Abrir(cfg, rutaEstado, b.eventos.Printf)

	b.eventos.Printf("Lince listo: genoma %q, fase %s, datos en %q", archivoGenoma.Bot, *faseForzada, rutaEstadoLegible(rutaEstado))
	return cerrar, nil
}

func rutaEstadoLegible(ruta string) string {
	if ruta == "" {
		return "(sin archivos)"
	}
	return filepath.Dir(ruta)
}

// --- Decisión ---------------------------------------------------------------

// Decidir implementa botsdk.Bot. Envuelve a decidir para que nada de lo que
// pase adentro pueda tirar abajo el proceso ni mandar una acción inválida:
// las dos cosas cuestan lo mismo que no responder, y un proceso caído además
// pierde las ciegas hasta que alguien lo relanza.
func (b *lince) Decidir(ctx context.Context, t botsdk.Turno) (accion botsdk.Accion) {
	inicio := time.Now()
	defer func() {
		if r := recover(); r != nil {
			accion = accionSegura(t)
			b.eventos.Printf("PÁNICO al decidir (%v): se juega %s\n%s", r, accion.Tipo, debug.Stack())
		}
		if !t.Puede(accion.Tipo) {
			invalida := accion
			accion = accionSegura(t)
			b.eventos.Printf("ERROR: la decisión %s no está entre las válidas %v: se juega %s", describir(invalida), t.Validas, accion.Tipo)
		}
		if tardo := time.Since(inicio); tardo > time.Second {
			b.eventos.Printf("AVISO: la decisión tardó %s (plazo %s)", tardo.Round(time.Millisecond), t.Timeout)
		}
	}()
	return b.decidir(ctx, t)
}

// accionSegura es la misma que juega la mesa por un bot que no responde:
// pasar si es gratis y, si no, retirarse.
func accionSegura(t botsdk.Turno) botsdk.Accion {
	if t.Puede(botsdk.Check) {
		return t.Check()
	}
	return t.Fold()
}

func (b *lince) decidir(ctx context.Context, t botsdk.Turno) botsdk.Accion {
	b.contexto.Identificar(t.IDJugador)
	b.contexto.Observar(t.Estado)

	// Se reserva como mucho un cuarto del plazo para pensar: el resto es
	// margen para la red. Pasarse significa que la mesa juega por nosotros.
	presupuesto := min(t.Timeout/4, 400*time.Millisecond)
	if presupuesto <= 0 {
		presupuesto = 200 * time.Millisecond
	}
	limite := time.Now().Add(presupuesto)
	if plazo, ok := ctx.Deadline(); ok && plazo.Before(limite) {
		limite = plazo
	}

	fuerza := estrategia.FuerzaConPlazo(t.Mano, t.Estado.Comunitarias, *muestras, limite, b.azar)
	situacion := estrategia.SituacionDesde(t, fuerza)
	// Dos modos, un solo bot: en la mesa final juega el genoma de la final.
	genoma := b.genoma
	if b.contexto.Fase() == torneo.Final {
		genoma = b.final
	}
	decision := genoma.Decidir(situacion, b.azar)
	accion := estrategia.AAccion(decision, t)

	b.detalle.Printf("[%-7s] %s %s | mesa %-14s | fuerza %3.0f%% (efectiva %3.0f%%) vs %d | %-10s | %s",
		t.Estado.Etapa, t.Mano[0], t.Mano[1], cartas(t.Estado.Comunitarias),
		fuerza*100, genoma.FuerzaEfectiva(situacion)*100, situacion.Rivales,
		describir(accion), decision.Motivo)
	return accion
}

// --- Seguimiento del torneo ---------------------------------------------------

// protegido corre f sin dejar que un pánico se escape: un error al llevar la
// cuenta del torneo no puede costar la conexión.
func (b *lince) protegido(que string, f func()) {
	defer func() {
		if r := recover(); r != nil {
			b.eventos.Printf("PÁNICO en %s (%v)\n%s", que, r, debug.Stack())
		}
	}()
	f()
}

// ManoInicio implementa botsdk.ObservaManos.
func (b *lince) ManoInicio(estado botsdk.EstadoPublico, _ botsdk.Mano) {
	b.protegido("ManoInicio", func() {
		nueva, motivo := b.contexto.ManoInicio(estado)
		if !nueva {
			return
		}
		e := b.contexto.Estado()
		b.eventos.Printf("MESA NUEVA · %s · ronda %d · %d jugadores con fichas · nuestro stack %d · motivo: %s",
			b.contexto.Fase(), e.Ronda, e.Rondas[len(e.Rondas)-1].Jugadores, e.Stack, motivo)
	})
}

// ManoFin implementa botsdk.ObservaManos.
func (b *lince) ManoFin(resultado botsdk.ResultadoMano) {
	b.protegido("ManoFin", func() {
		if err := b.contexto.ManoFin(resultado); err != nil {
			b.eventos.Printf("ERROR al guardar el estado: %v", err)
		}
	})
}

// EstadoActualizado implementa botsdk.ObservaEstado.
func (b *lince) EstadoActualizado(estado botsdk.EstadoPublico) {
	b.protegido("EstadoActualizado", func() { b.contexto.Observar(estado) })
}

// --- Conexión -----------------------------------------------------------------

// capturaID mira lo que registra el SDK para enterarse del id_jugador apenas
// llega la bienvenida. El SDK solo se lo pasa al bot en su primer turno, y el
// seguimiento del torneo lo necesita desde la primera mano_inicio.
type capturaID struct {
	destino io.Writer
	alVerID func(id string)
}

var patronConectado = regexp.MustCompile(`conectado a \S+ como (\S+) \(silla`)

func (c capturaID) Write(p []byte) (int, error) {
	if m := patronConectado.FindSubmatch(p); m != nil {
		c.alVerID(string(m[1]))
	}
	return c.destino.Write(p)
}

const (
	esperaMinima = time.Second
	esperaMaxima = 5 * time.Second
	// Una conexión que duró más que esto se considera buena: la próxima caída
	// vuelve a empezar con la espera mínima.
	conexionEstable = 30 * time.Second
)

// siguienteEspera es la espera antes del próximo intento de conexión, a
// partir de la anterior (0 la primera vez): crece de a un segundo hasta el
// máximo, y vuelve al mínimo después de una conexión que duró. No hay límite de intentos: en el torneo, rendirse es regalar las
// ciegas hasta el final.
func siguienteEspera(actual, duroLaConexion time.Duration) time.Duration {
	if duroLaConexion >= conexionEstable {
		return esperaMinima
	}
	return min(actual+time.Second, esperaMaxima)
}

func main() {
	flag.Parse()
	bot := &lince{}
	cerrar, err := bot.preparar()
	if err != nil {
		log.Fatalf("%s: %v", nombreBot, err)
	}
	defer cerrar()

	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	opciones := botsdk.Opciones{
		Direccion: *direccion,
		Token:     *token,
		// Pocos reintentos dentro del SDK: del resto se encarga el bucle de
		// abajo, con esperas más largas (el servidor limita los intentos por IP).
		Reintentos:      3,
		EsperaReintento: 300 * time.Millisecond,
		Log: log.New(capturaID{destino: bot.eventos.Writer(), alVerID: func(id string) {
			bot.protegido("Identificar", func() { bot.contexto.Identificar(id) })
		}}, "sdk: ", log.LstdFlags|log.Lmicroseconds),
	}
	if opciones.Token == "" {
		// El token del equipo se toma de la variable de entorno: así no queda
		// en la línea de comandos (ni en los logs ni en la lista de procesos).
		opciones.Token = os.Getenv("POKERFIGHT_TOKEN")
	}
	if opciones.Token == "" {
		opciones.Token = nombreBot
	}

	var espera time.Duration
	for intento := 1; ; intento++ {
		inicio := time.Now()
		err := botsdk.Correr(ctx, bot, opciones)
		if ctx.Err() != nil {
			bot.eventos.Printf("detenido a mano")
			return
		}
		if *unaPartida {
			if err != nil && !errors.Is(err, botsdk.ErrMesaCerro) {
				bot.eventos.Printf("ERROR: %v", err)
				cerrar()
				os.Exit(1)
			}
			return
		}

		duro := time.Since(inicio)
		espera = siguienteEspera(espera, duro)
		if errors.Is(err, botsdk.ErrMesaCerro) {
			bot.eventos.Printf("DESCONECTADO tras %s: la mesa cerró la conexión; intento %d de reconexión en %s", duro.Round(time.Second), intento, espera)
		} else {
			bot.eventos.Printf("SIN CONEXIÓN (%v); intento %d de reconexión en %s", err, intento, espera)
		}
		select {
		case <-ctx.Done():
			bot.eventos.Printf("detenido a mano")
			return
		case <-time.After(espera):
		}
	}
}

// --- Auxiliares ---------------------------------------------------------------

func cartas(cs []botsdk.Carta) string {
	if len(cs) == 0 {
		return "-"
	}
	texto := ""
	for i, c := range cs {
		if i > 0 {
			texto += " "
		}
		texto += c.String()
	}
	return texto
}

func describir(a botsdk.Accion) string {
	if a.Tipo == botsdk.Bet || a.Tipo == botsdk.Raise {
		return fmt.Sprintf("%s %d", a.Tipo, a.Monto)
	}
	return string(a.Tipo)
}
