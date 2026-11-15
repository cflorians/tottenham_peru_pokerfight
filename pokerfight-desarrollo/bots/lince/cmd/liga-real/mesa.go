package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/aicspucp/pokerfight-arena-participantes/internal/arena"
	"github.com/aicspucp/pokerfight-arena-participantes/internal/crupier"
	"github.com/aicspucp/pokerfight-arena-participantes/internal/mesa"
	"github.com/aicspucp/pokerfight-arena-participantes/internal/protocolo"
)

// Una mesa real: el servidor de mesa del kit en un puerto local, con los bots
// del kit lanzados como subprocesos y Lince (si está sentado) conectándose
// solo a la dirección fija.

// configMesaReal son los parámetros de una mesa.
type configMesaReal struct {
	ID        string
	Direccion string // host:puerto; puerto 0 = uno libre.
	Manos     int64  // Tope de manos (-1 = sin tope).
	// Reloj es el tope de tiempo de la mesa (0 = sin tope). Se respeta al
	// terminar una mano: la mano en curso se termina y la mesa cierra con las
	// fichas que tenga cada uno, así ninguna ficha queda en un pozo sin repartir.
	Reloj          time.Duration
	TimeoutMs      uint64
	EsperaConexion time.Duration // Cuánto se espera a que se conecten todos antes de repartir.
	Semilla        uint64
	CiegaChica     uint64
	CiegaGrande    uint64
}

// asiento es un jugador sentado en la mesa, en orden de silla.
type asiento struct {
	ID      string
	Fichas  int64
	Comando []string // Comando del bot del kit; nil = Lince, que se conecta solo.
	Token   string   // Solo para Lince: el token fijo con el que se presenta.
}

// resultadoMesa es cómo terminó una mesa.
type resultadoMesa struct {
	ID              string                  `json:"id"`
	Direccion       string                  `json:"direccion"`
	Jugadores       []string                `json:"jugadores"` // En orden de silla.
	FichasInicio    map[string]int64        `json:"fichas_inicio"`
	FichasFin       map[string]int64        `json:"fichas_fin"`
	Posiciones      []string                `json:"posiciones"` // Orden final de la mesa (1.º primero).
	Manos           int                     `json:"manos"`
	DuracionSeg     float64                 `json:"duracion_seg"`
	CerradaPorReloj bool                    `json:"cerrada_por_reloj,omitempty"`
	CorteDuro       bool                    `json:"corte_duro,omitempty"` // Se cortó a mitad de una mano: puede haber fichas perdidas.
	Ausentes        []string                `json:"ausentes,omitempty"`   // No se conectaron antes de empezar.
	Timeouts        map[string]uint64       `json:"timeouts"`
	Quiebras        map[string]marcaQuiebra `json:"quiebras,omitempty"`
	Error           string                  `json:"error,omitempty"`
}

// conexionAusente ocupa la silla de un jugador que todavía no se conectó. Con
// ella todos quedan sentados con sus fichas desde el principio, en la silla
// sorteada: cuando el bot se conecta, el servidor lo toma como una reconexión
// y reemplaza esta conexión por la real. Si nunca llega, la mesa le juega la
// acción segura (como el servidor real con un bot caído) y sus fichas siguen
// contando.
type conexionAusente struct{}

func (conexionAusente) EnviarMensaje(protocolo.MensajeMesa, uint64) error {
	return mesa.ErrConexionCerrada
}
func (conexionAusente) SolicitarAccion(uint64) (protocolo.Accion, error) {
	return protocolo.Accion{}, mesa.ErrConexionCerrada
}
func (conexionAusente) Cerrar() error { return nil }

// observadorMesa sigue la mesa mano a mano: cuenta las manos, anota quién
// quiebra y en qué mano, y cierra la mesa cuando se cumple el reloj.
type observadorMesa struct {
	inicio   time.Time
	reloj    time.Duration
	cancelar context.CancelFunc

	manos    int
	saldos   map[string]int64
	quiebras map[string]marcaQuiebra
	porReloj bool
	ronda    int
	mu       sync.Mutex
}

func (o *observadorMesa) ManoTerminada(m mesa.ManoJugada) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.manos++
	for _, j := range m.Estado.Jugadores {
		antes := o.saldos[j.ID]
		if antes > 0 && j.Saldo == 0 {
			o.quiebras[j.ID] = marcaQuiebra{Ronda: o.ronda, Mano: o.manos, StackAntes: antes}
		}
		o.saldos[j.ID] = j.Saldo
	}
	if o.reloj > 0 && time.Since(o.inicio) >= o.reloj && o.cancelar != nil {
		o.porReloj = true
		o.cancelar()
	}
}

// corteDuro es el margen sobre el reloj tras el cual la mesa se corta aunque
// esté a mitad de una mano (solo para que nada quede colgado para siempre).
const corteDuro = 3 * time.Minute

// jugarMesaReal juega una mesa completa y devuelve el resultado. Nunca deja
// procesos de bots del kit vivos: al volver, todos terminaron o fueron
// matados. ronda es el número de ronda que se anota en las quiebras.
func jugarMesaReal(ctx context.Context, cfg configMesaReal, ronda int, asientos []asiento, salidaBots io.Writer, logServidor *log.Logger) (res resultadoMesa) {
	inicio := time.Now()
	res = resultadoMesa{
		ID:           cfg.ID,
		FichasInicio: map[string]int64{},
		FichasFin:    map[string]int64{},
		Timeouts:     map[string]uint64{},
	}
	stacks := make(map[string]uint64, len(asientos))
	for _, a := range asientos {
		res.Jugadores = append(res.Jugadores, a.ID)
		res.FichasInicio[a.ID] = a.Fichas
		res.FichasFin[a.ID] = a.Fichas // Si la mesa no llega a jugarse, nadie pierde nada.
		stacks[a.ID] = uint64(a.Fichas)
	}
	defer func() { res.DuracionSeg = time.Since(inicio).Seconds() }()

	obs := &observadorMesa{saldos: map[string]int64{}, quiebras: map[string]marcaQuiebra{}, ronda: ronda, reloj: cfg.Reloj}
	for _, a := range asientos {
		obs.saldos[a.ID] = a.Fichas
	}

	m := mesa.NuevaMesa(cfg.ID, mesa.ConfigMesa{
		MaxJugadores: uint8(len(asientos)),
		MinJugadores: uint8(len(asientos)),
		Timeout:      cfg.TimeoutMs,
		Observador:   obs,
	}, mesa.ConfigPartida{
		CiegaMenor:      cfg.CiegaChica,
		CiegaMayor:      cfg.CiegaGrande,
		StackInicial:    10000,
		CantidadRondas:  cfg.Manos,
		StackPorJugador: stacks,
	}, crupier.NuevoConSemilla(cfg.Semilla))

	// Todos sentados desde ya, en su silla, con sus fichas (ver conexionAusente).
	for silla, a := range asientos {
		if err := m.SentarJugadorEnSilla(a.ID, a.ID, silla, conexionAusente{}); err != nil {
			res.Error = fmt.Sprintf("no se pudo sentar a %s: %v", a.ID, err)
			return res
		}
	}

	// Tokens: cada bot del kit recibe uno secreto y nuevo; Lince usa siempre el
	// mismo, que la mesa traduce al mismo id en todas las mesas.
	porToken := make(map[string]string, len(asientos))
	tokens := make([]string, len(asientos))
	for i, a := range asientos {
		tokens[i] = a.Token
		if a.Comando != nil {
			tokens[i] = tokenAleatorio()
		}
		porToken[tokens[i]] = a.ID
	}
	validar := func(token string) (string, string, error) {
		id, ok := porToken[token]
		if !ok {
			return "", "", errors.New("token desconocido en esta mesa")
		}
		return id, id, nil
	}

	var muConectados sync.Mutex
	conectados := map[string]bool{}
	servidor := mesa.NuevoServidor(mesa.ConfigServidor{
		Direccion:          cfg.Direccion,
		TimeoutHandshakeMs: 10000,
		MaxConexiones:      len(asientos),
		Log:                logServidor,
		AlConectar: func(id string) {
			muConectados.Lock()
			conectados[id] = true
			muConectados.Unlock()
		},
	}, m, validar)

	// El orden de los defers importa (como en la arena): primero se cierra el
	// servidor, que les corta el socket a los bots, y recién después se los
	// detiene, que espera a que salgan solos o los mata.
	var sesiones []arena.Sesion
	defer func() {
		var espera sync.WaitGroup
		for _, s := range sesiones {
			espera.Add(1)
			go func(s arena.Sesion) {
				defer espera.Done()
				s.Detener()
			}(s)
		}
		espera.Wait()
	}()

	// La dirección fija de Lince puede tardar un instante en liberarse tras la
	// mesa anterior: se reintenta.
	var errAbrir error
	for intento := 0; intento < 20; intento++ {
		if errAbrir = servidor.Abrir(); errAbrir == nil {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if errAbrir != nil {
		res.Error = fmt.Sprintf("no se pudo abrir %s: %v", cfg.Direccion, errAbrir)
		return res
	}
	defer servidor.Cerrar()
	res.Direccion = servidor.Direccion()

	ctxMesa, cancelarMesa := context.WithCancel(ctx)
	defer cancelarMesa()
	go func() {
		if err := servidor.Servir(ctxMesa); err != nil {
			logServidor.Printf("%s: servidor detenido: %v", cfg.ID, err)
		}
	}()

	lanzador := arena.LanzadorProceso{Salida: salidaBots}
	for i, a := range asientos {
		if a.Comando == nil {
			continue
		}
		s, err := lanzador.Lanzar(ctxMesa, arena.Participante{ID: a.ID, Comando: a.Comando, Token: tokens[i]}, servidor.Direccion())
		if err != nil {
			// La mesa se juega igual: ese bot queda ausente y juega la acción
			// segura, como un bot caído en el torneo.
			logServidor.Printf("%s: no se pudo lanzar %s: %v", cfg.ID, a.ID, err)
			continue
		}
		sesiones = append(sesiones, s)
	}

	// Se espera a que lleguen todos (o se cumpla el plazo).
	limite := time.Now().Add(cfg.EsperaConexion)
	for {
		muConectados.Lock()
		n := len(conectados)
		muConectados.Unlock()
		if n >= len(asientos) || time.Now().After(limite) || ctx.Err() != nil {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	muConectados.Lock()
	for _, a := range asientos {
		if !conectados[a.ID] {
			res.Ausentes = append(res.Ausentes, a.ID)
		}
	}
	muConectados.Unlock()
	if ctx.Err() != nil {
		res.Error = ctx.Err().Error()
		return res
	}

	// Dos contextos: el del reloj (lo cancela el observador entre manos) y,
	// por debajo, un corte duro por si una mano no termina nunca.
	ctxDuro := ctxMesa
	if cfg.Reloj > 0 {
		var cancelarDuro context.CancelFunc
		ctxDuro, cancelarDuro = context.WithTimeout(ctxMesa, cfg.Reloj+corteDuro)
		defer cancelarDuro()
	}
	ctxJuego, cancelarJuego := context.WithCancel(ctxDuro)
	defer cancelarJuego()
	obs.mu.Lock()
	obs.inicio = time.Now()
	obs.cancelar = cancelarJuego
	obs.mu.Unlock()

	resumen, err := m.Jugar(ctxJuego)
	servidor.Cerrar()

	obs.mu.Lock()
	res.Manos = obs.manos
	res.CerradaPorReloj = obs.porReloj
	res.Quiebras = obs.quiebras
	obs.mu.Unlock()
	if err != nil {
		switch {
		case res.CerradaPorReloj:
			// Cierre normal por reloj, entre dos manos.
		case ctxDuro.Err() != nil && ctx.Err() == nil:
			res.CorteDuro = true
			res.Error = "corte duro a mitad de una mano: " + err.Error()
		default:
			res.Error = err.Error()
		}
	}
	for _, p := range resumen.Posiciones {
		res.FichasFin[p.IDJugador] = int64(p.SaldoFinal)
		res.Timeouts[p.IDJugador] = p.Timeouts
		res.Posiciones = append(res.Posiciones, p.IDJugador)
	}
	return res
}

func tokenAleatorio() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("liga-real: no se pudo generar un token: " + err.Error())
	}
	return "sesion-" + hex.EncodeToString(b[:])
}
