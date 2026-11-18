// Package torneo lleva la cuenta de en qué punto del campeonato está el bot:
// qué ronda es, cuántas manos van, si ya empezó la mesa final y qué fichas
// tenía cada rival la última vez que se lo vio.
//
// La arena no avisa nada de esto. En el campeonato la conexión dura todo el
// torneo y cada ronda empieza con un mano_inicio como cualquier otro: lo único
// que cambia son las fichas y los rivales. Así que la ronda se deduce.
//
// Todo queda guardado en disco al cerrar cada mano. Si el proceso se cae y el
// supervisor lo relanza a mitad del torneo, el bot tiene que seguir sabiendo
// que está en la ronda 7 y no creer que empieza de cero.
package torneo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aicspucp/pokerfight-arena-participantes/pkg/botsdk"
)

// Fase es la etapa del campeonato.
type Fase string

const (
	// Clasificatoria: 10 rondas con fichas arrastradas. Es la fase por
	// defecto, y la que se asume ante cualquier duda: jugar la final con la
	// prudencia de la clasificatoria cuesta poco; jugar la clasificatoria sin
	// ella puede costar la eliminación.
	Clasificatoria Fase = "clasificatoria"
	// Final: una mesa de 6 con 10 000 fichas cada uno y un solo ganador.
	Final Fase = "final"
)

// Config son las condiciones de la competencia (docs/competencia.md).
type Config struct {
	ManosPorRonda  int           // Manos que dura como mucho una mesa de la clasificatoria.
	StackInicial   int64         // Fichas con las que empieza cada bot (y la final).
	JugadoresFinal int           // Jugadores de la mesa final.
	PausaNuevaMesa time.Duration // Silencio que, sin otra pista, se toma como cambio de mesa.
	Caducidad      time.Duration // Un estado guardado más viejo que esto es de otro torneo.
	FaseForzada    Fase          // Si no es vacía, no se deduce la fase: se usa esta.
}

// ConfigCompetencia son los valores de las bases.
func ConfigCompetencia() Config {
	return Config{
		ManosPorRonda:  50,
		StackInicial:   10000,
		JugadoresFinal: 6,
		PausaNuevaMesa: 30 * time.Second,
		Caducidad:      6 * time.Hour,
	}
}

// Ronda es el registro de una mesa ya jugada o en curso.
type Ronda struct {
	Numero      int       `json:"numero"`
	Fase        Fase      `json:"fase"`
	Inicio      time.Time `json:"inicio"`
	Jugadores   int       `json:"jugadores"`    // Sentados con fichas al empezar.
	StackInicio int64     `json:"stack_inicio"` // Nuestras fichas al empezar.
	StackFin    int64     `json:"stack_fin"`    // Nuestras fichas al terminar (0 = todavía no se sabe).
	Manos       int       `json:"manos"`
	Motivo      string    `json:"motivo"` // Por qué se la tomó como mesa nueva.
}

// Alias es lo que se sabe de un rival. Los alias son estables durante todo el
// torneo, así que lo visto en la ronda 2 vale en la 7.
type Alias struct {
	Stack       int64 `json:"stack"`        // Fichas la última vez que se lo vio.
	UltimaRonda int   `json:"ultima_ronda"` // Última ronda en la que compartimos mesa.
	Mesas       int   `json:"mesas"`        // Cuántas mesas compartimos.
	Eliminado   bool  `json:"eliminado"`    // Se lo vio quedarse sin fichas.
}

// Estado es todo lo que se guarda en disco.
type Estado struct {
	Version     int       `json:"version"`
	IDJugador   string    `json:"id_jugador"`
	Actualizado time.Time `json:"actualizado"`

	Fase         Fase      `json:"fase"`
	Ronda        int       `json:"ronda"`         // 1 = primera mesa. 0 = todavía no jugó.
	ManosRonda   int       `json:"manos_ronda"`   // Manos repartidas en la mesa actual.
	ManosTotales int       `json:"manos_totales"` // Manos repartidas en todo el torneo.
	InicioRonda  time.Time `json:"inicio_ronda"`
	InicioFinal  time.Time `json:"inicio_final"` // Cero si la final no empezó.

	Mesa          string    `json:"mesa"`            // Prefijo del id de mano ("mesa-1").
	NumeroMano    int       `json:"numero_mano"`     // Número de la última mano vista (0 = no se pudo leer).
	UltimaMano    string    `json:"ultima_mano"`     // id_mano de la última mano vista.
	UltimaManoFin time.Time `json:"ultima_mano_fin"` // Cuándo cerró la última mano.
	EnMesa        []string  `json:"en_mesa"`         // Todos los sentados en la mesa actual.

	Stack  int64            `json:"stack"` // Nuestras fichas al empezar la última mano.
	Rondas []Ronda          `json:"rondas"`
	Alias  map[string]Alias `json:"alias"`
}

const versionEstado = 1

// Contexto es el seguimiento en vivo. Lo usa una sola goroutine (la del SDK).
type Contexto struct {
	cfg    Config
	estado Estado
	ruta   string // "" = no se guarda en disco.

	// saldos es la última foto de las fichas de cada jugador durante la mano
	// en curso. No se persiste: sirve para estimar cómo quedó cada uno al
	// cerrar la mano.
	saldos map[string]int64

	Ahora func() time.Time                  // Reloj; se reemplaza en las pruebas.
	Log   func(formato string, args ...any) // Avisos; nil = no avisa.
}

// Nuevo arma un contexto vacío. Si ruta no es vacía, el estado se guarda ahí.
func Nuevo(cfg Config, ruta string) *Contexto {
	c := &Contexto{cfg: cfg, ruta: ruta, Ahora: time.Now}
	c.reiniciar("")
	return c
}

// Abrir es Nuevo, pero retoma el estado guardado en ruta si existe y es de
// este torneo. Un archivo ilegible o vencido se aparta con otro nombre (no se
// borra: puede servir para entender qué pasó) y se arranca de cero.
func Abrir(cfg Config, ruta string, log func(string, ...any)) *Contexto {
	c := Nuevo(cfg, ruta)
	c.Log = log
	c.Retomar()
	return c
}

// Retomar carga el estado guardado en disco, si hay uno que sirva. Devuelve si
// lo retomó.
func (c *Contexto) Retomar() bool {
	ruta := c.ruta
	if ruta == "" {
		return false
	}
	datos, err := os.ReadFile(ruta)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	if err != nil {
		c.avisar("estado: no se pudo leer %s (%v); se arranca de cero", ruta, err)
		return false
	}
	var guardado Estado
	if err := json.Unmarshal(datos, &guardado); err != nil || guardado.Version != versionEstado {
		c.avisar("estado: %s es ilegible o de otra versión; se aparta y se arranca de cero", ruta)
		c.apartar("ilegible")
		return false
	}
	if edad := c.Ahora().Sub(guardado.Actualizado); c.cfg.Caducidad > 0 && edad > c.cfg.Caducidad {
		c.avisar("estado: %s tiene %s sin actualizarse: es de otro torneo; se aparta y se arranca de cero", ruta, edad.Round(time.Minute))
		c.apartar("vencido")
		return false
	}
	if guardado.Alias == nil {
		guardado.Alias = map[string]Alias{}
	}
	// Los estados anteriores podían guardar una fase deducida distinta de la
	// fase con que se jugó. La ronda en curso anotaba la fase efectiva.
	var ultimaRonda *Ronda
	if n := len(guardado.Rondas); n > 0 && guardado.Rondas[n-1].Numero == guardado.Ronda {
		ultimaRonda = &guardado.Rondas[n-1]
		if ultimaRonda.Fase != "" {
			guardado.Fase = ultimaRonda.Fase
		}
	}
	c.estado = guardado
	c.estado.Fase = c.Fase()
	if guardado.Fase != Final || c.estado.Fase != Final {
		c.estado.InicioFinal = time.Time{}
	}
	if ultimaRonda != nil {
		ultimaRonda.Fase = c.estado.Fase
	}
	c.avisar("estado: retomado de %s (jugador %s, %s, ronda %d, mano %d de la ronda, %d fichas)",
		ruta, guardado.IDJugador, c.estado.Fase, guardado.Ronda, guardado.ManosRonda, guardado.Stack)
	return true
}

func (c *Contexto) reiniciar(id string) {
	c.estado = Estado{Version: versionEstado, IDJugador: id, Alias: map[string]Alias{}}
	c.estado.Fase = c.Fase()
	c.saldos = map[string]int64{}
}

func (c *Contexto) avisar(formato string, args ...any) {
	if c.Log != nil {
		c.Log(formato, args...)
	}
}

// apartar renombra el archivo de estado para que no se vuelva a leer.
func (c *Contexto) apartar(motivo string) {
	if c.ruta == "" {
		return
	}
	ext := filepath.Ext(c.ruta)
	destino := fmt.Sprintf("%s.%s-%s%s", strings.TrimSuffix(c.ruta, ext), motivo, c.Ahora().Format("20060102-150405"), ext)
	if err := os.Rename(c.ruta, destino); err != nil && !errors.Is(err, os.ErrNotExist) {
		c.avisar("estado: no se pudo apartar %s: %v", c.ruta, err)
	}
}

// Identificar fija quién es el bot en esta conexión (el id_jugador de la
// bienvenida). Si el estado retomado es de otro jugador, es de otro torneo:
// se aparta y se arranca de cero.
func (c *Contexto) Identificar(id string) {
	if id == "" || id == c.estado.IDJugador {
		return
	}
	if c.estado.IDJugador != "" {
		c.avisar("estado: el guardado es del jugador %s y ahora somos %s: es otro torneo; se arranca de cero", c.estado.IDJugador, id)
		c.apartar("otro-jugador")
		c.reiniciar(id)
		return
	}
	c.estado.IDJugador = id
}

// --- Consultas ----------------------------------------------------------------

// Estado devuelve una copia de lo que se sabe hasta ahora.
func (c *Contexto) Estado() Estado { return c.estado }

// Fase es la etapa del campeonato: la forzada por configuración si hay, y si
// no la deducida.
func (c *Contexto) Fase() Fase {
	if c.cfg.FaseForzada != "" {
		return c.cfg.FaseForzada
	}
	if c.estado.Fase == "" {
		return Clasificatoria
	}
	return c.estado.Fase
}

// Ritmo son las manos por minuto de la mesa actual (0 si todavía no se puede
// calcular). Depende de lo rápido que respondan los rivales: con bots lentos
// una mesa no llega a las 50 manos en sus 6 minutos.
func (c *Contexto) Ritmo() float64 {
	minutos := c.Ahora().Sub(c.estado.InicioRonda).Minutes()
	if c.estado.ManosRonda < 2 || minutos <= 0 {
		return 0
	}
	return float64(c.estado.ManosRonda-1) / minutos
}

// --- Eventos ------------------------------------------------------------------

// ManoInicio registra el comienzo de una mano. Devuelve si con ella empezó una
// mesa nueva (una ronda nueva de la clasificatoria, o la final) y por qué.
func (c *Contexto) ManoInicio(e botsdk.EstadoPublico) (nuevaMesa bool, motivo string) {
	if e.IDMano != "" && e.IDMano == c.estado.UltimaMano {
		// La misma mano otra vez (por ejemplo, tras reconectar): no se cuenta dos veces.
		c.observar(e)
		return false, ""
	}

	mesa, numero := partirIDMano(e.IDMano)
	nuevaMesa, motivo = c.esMesaNueva(e, mesa, numero)
	if nuevaMesa {
		motivo = c.abrirRonda(e, mesa, motivo)
	}

	c.estado.ManosRonda++
	c.estado.ManosTotales++
	c.estado.NumeroMano = numero
	c.estado.UltimaMano = e.IDMano
	if n := len(c.estado.Rondas); n > 0 {
		c.estado.Rondas[n-1].Manos = c.estado.ManosRonda
	}

	c.saldos = map[string]int64{}
	c.observar(e)
	if yo, ok := buscar(e, c.estado.IDJugador); ok {
		c.estado.Stack = yo.Saldo + yo.ApuestaRonda
	}
	for _, j := range e.Jugadores {
		if j.ID == c.estado.IDJugador {
			continue
		}
		a := c.estado.Alias[j.ID]
		a.Stack = j.Saldo + j.ApuestaRonda
		a.UltimaRonda = c.estado.Ronda
		a.Eliminado = !j.EnTorneo
		c.estado.Alias[j.ID] = a
	}
	return nuevaMesa, motivo
}

// esMesaNueva decide si esta mano es la primera de una mesa distinta. Usa
// varias pistas porque no se sabe qué forma tienen los id de mano del servidor
// del torneo, ni si dos rondas seguidas pueden repetir exactamente los mismos
// rivales (pasa cuando quedan pocos bots vivos).
func (c *Contexto) esMesaNueva(e botsdk.EstadoPublico, mesa string, numero int) (bool, string) {
	s := &c.estado
	if s.Ronda == 0 {
		return true, "primera mano"
	}
	if mesa != "" && s.Mesa != "" && mesa != s.Mesa {
		return true, fmt.Sprintf("cambió la mesa (%s → %s)", s.Mesa, mesa)
	}
	conocidos := make(map[string]bool, len(s.EnMesa))
	for _, id := range s.EnMesa {
		conocidos[id] = true
	}
	for _, j := range e.Jugadores {
		if !conocidos[j.ID] {
			return true, "hay un jugador que no estaba en la mesa (" + j.ID + ")"
		}
	}
	if numero > 0 && s.NumeroMano > 0 && mesa == s.Mesa {
		if numero < s.NumeroMano {
			return true, fmt.Sprintf("el número de mano volvió atrás (%d → %d)", s.NumeroMano, numero)
		}
		// Misma mesa y numeración que sigue: es la misma partida, aunque
		// hayamos estado desconectados un rato.
		return false, ""
	}
	// De aquí en más el id de mano no ayuda: quedan la cuenta de manos y el reloj.
	if c.Fase() == Clasificatoria && c.cfg.ManosPorRonda > 0 && s.ManosRonda >= c.cfg.ManosPorRonda {
		return true, fmt.Sprintf("la mesa anterior ya jugó sus %d manos", c.cfg.ManosPorRonda)
	}
	if pausa := c.Ahora().Sub(s.UltimaManoFin); c.cfg.PausaNuevaMesa > 0 && !s.UltimaManoFin.IsZero() && pausa >= c.cfg.PausaNuevaMesa {
		return true, fmt.Sprintf("pasaron %s desde la última mano", pausa.Round(time.Second))
	}
	return false, ""
}

// abrirRonda registra la mesa nueva y decide si es la final. Devuelve el motivo
// completo, con el aviso de la final si corresponde.
func (c *Contexto) abrirRonda(e botsdk.EstadoPublico, mesa, motivo string) string {
	s := &c.estado
	ahora := c.Ahora()

	var miStack int64
	if yo, ok := buscar(e, s.IDJugador); ok {
		miStack = yo.Saldo + yo.ApuestaRonda
	}
	faseAnterior := s.Fase

	conFichas := 0
	todosConStackInicial := true
	s.EnMesa = s.EnMesa[:0]
	for _, j := range e.Jugadores {
		s.EnMesa = append(s.EnMesa, j.ID)
		if !j.EnTorneo {
			continue
		}
		conFichas++
		if j.Saldo+j.ApuestaRonda != c.cfg.StackInicial {
			todosConStackInicial = false
		}
		if j.ID != s.IDJugador {
			a := s.Alias[j.ID]
			a.Mesas++
			s.Alias[j.ID] = a
		}
	}

	// La final: después de haber jugado al menos una mesa, aparecen 6 jugadores
	// y todos con las fichas iniciales. En la ronda 2 o posterior de la
	// clasificatoria eso no puede pasar por casualidad: nadie termina una mesa
	// de 50 manos con exactamente las fichas con las que empezó, y menos seis.
	if c.cfg.FaseForzada == "" && s.Fase != Final && s.Ronda >= 1 && conFichas == c.cfg.JugadoresFinal && todosConStackInicial {
		s.Fase = Final
		motivo += fmt.Sprintf("; %d jugadores con %d fichas cada uno: es la FINAL", conFichas, c.cfg.StackInicial)
	}
	if s.Fase == Final && s.InicioFinal.IsZero() {
		s.InicioFinal = ahora
	}

	// En la clasificatoria las fichas se arrastran: lo que tenemos al empezar
	// esta mesa es exactamente con lo que terminamos la anterior. En el
	// servidor de pruebas, la clasificatoria forzada admite reinicios de fichas.
	fichasArrastradas := c.cfg.FaseForzada != Clasificatoria || !todosConStackInicial
	if n := len(s.Rondas); n > 0 && faseAnterior == Clasificatoria && s.Fase == Clasificatoria && fichasArrastradas && miStack > 0 {
		s.Rondas[n-1].StackFin = miStack
	}

	s.Ronda++
	s.ManosRonda = 0
	s.InicioRonda = ahora
	s.Mesa = mesa
	s.Rondas = append(s.Rondas, Ronda{
		Numero:      s.Ronda,
		Fase:        s.Fase,
		Inicio:      ahora,
		Jugadores:   conFichas,
		StackInicio: miStack,
		Motivo:      motivo,
	})
	return motivo
}

// Observar registra una foto de la mesa a mitad de mano (un estado, o el
// estado de una solicitud de acción).
func (c *Contexto) Observar(e botsdk.EstadoPublico) { c.observar(e) }

func (c *Contexto) observar(e botsdk.EstadoPublico) {
	for _, j := range e.Jugadores {
		c.saldos[j.ID] = j.Saldo
	}
}

// ManoFin registra el cierre de una mano y guarda el estado en disco.
//
// Las fichas con que queda cada uno se estiman con la última foto de la mano
// más lo que cobró. Puede quedar corta en la última apuesta de quien no la
// cobró; la mano siguiente lo corrige con el dato exacto.
func (c *Contexto) ManoFin(r botsdk.ResultadoMano) error {
	s := &c.estado
	s.UltimaManoFin = c.Ahora()

	cobrado := make(map[string]int64, len(r.Repartos))
	for _, rep := range r.Repartos {
		cobrado[rep.IDJugador] += rep.Monto
	}
	for id, saldo := range c.saldos {
		final := saldo + cobrado[id]
		if id == s.IDJugador {
			if n := len(s.Rondas); n > 0 {
				s.Rondas[n-1].StackFin = final
			}
			continue
		}
		a := s.Alias[id]
		a.Stack = final
		if final == 0 {
			a.Eliminado = true
		}
		s.Alias[id] = a
	}
	return c.Guardar()
}

// Guardar escribe el estado en disco. Primero a un archivo temporal y después
// lo renombra: si el proceso muere a mitad de la escritura, el archivo bueno
// anterior sigue ahí.
func (c *Contexto) Guardar() error {
	if c.ruta == "" {
		return nil
	}
	c.estado.Actualizado = c.Ahora()
	datos, err := json.MarshalIndent(c.estado, "", " ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(c.ruta), 0o755); err != nil {
		return err
	}
	temporal := c.ruta + ".tmp"
	if err := os.WriteFile(temporal, datos, 0o644); err != nil {
		return err
	}
	// En Windows el reemplazo falla ("Access is denied") si justo en ese
	// instante otro proceso tiene abierto el archivo: un antivirus, un editor,
	// alguien mirándolo. Dura milisegundos, así que se reintenta un poco. Si no
	// sale, no es grave: el estado sigue en memoria y se guarda en la mano
	// siguiente.
	for intento := 1; ; intento++ {
		err = os.Rename(temporal, c.ruta)
		if err == nil || intento == intentosGuardar {
			return err
		}
		time.Sleep(time.Duration(intento) * 5 * time.Millisecond)
	}
}

// intentosGuardar es cuántas veces se intenta reemplazar el archivo de estado.
// En total espera como mucho 50 ms: no llega a notarse entre mano y mano.
const intentosGuardar = 5

// --- Auxiliares ---------------------------------------------------------------

// partirIDMano separa "mesa-1-mano-7" en ("mesa-1", 7). Si el id no tiene esa
// forma devuelve ("", 0): el servidor del torneo puede usar otra.
func partirIDMano(id string) (mesa string, numero int) {
	const separador = "-mano-"
	i := strings.LastIndex(id, separador)
	if i < 0 {
		return "", 0
	}
	n, err := strconv.Atoi(id[i+len(separador):])
	if err != nil || n <= 0 {
		return "", 0
	}
	return id[:i], n
}

func buscar(e botsdk.EstadoPublico, id string) (botsdk.JugadorPublico, bool) {
	if id == "" {
		return botsdk.JugadorPublico{}, false
	}
	for _, j := range e.Jugadores {
		if j.ID == id {
			return j, true
		}
	}
	return botsdk.JugadorPublico{}, false
}
