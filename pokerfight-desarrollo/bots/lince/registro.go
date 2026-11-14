package main

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/aicspucp/pokerfight-arena-participantes/pkg/botsdk"
)

// registroManos guarda cada mano terminada como una línea JSON en
// datos/manos-AAAAMMDD.jsonl: quiénes jugaron, con cuántas fichas, todas las
// acciones de todos, las cartas comunitarias y las que se mostraron.
//
// Es la materia prima para saber cómo juegan los rivales de verdad (lo lee
// pruebas/rivales.py): contra los bots del kit podemos medir lo que queramos,
// pero los otros equipos solo se ven en el servidor de la organización, y el
// servidor no guarda el historial por nosotros.
type registroManos struct {
	mu      sync.Mutex
	archivo *os.File

	idMano string
	inicio []jugadorRegistrado
	mias   botsdk.Mano
	ultimo botsdk.EstadoPublico
}

type jugadorRegistrado struct {
	ID    string `json:"id"`
	Saldo int64  `json:"saldo"`
}

type manoRegistrada struct {
	Hora         time.Time                 `json:"hora"`
	IDMano       string                    `json:"id_mano"`
	Fase         string                    `json:"fase"`
	Yo           string                    `json:"yo"`
	MisCartas    botsdk.Mano               `json:"mis_cartas"`
	CiegaGrande  int64                     `json:"ciega_grande"`
	Boton        int                       `json:"boton"`
	Jugadores    []jugadorRegistrado       `json:"jugadores"` // En orden de silla, con las fichas al empezar la mano.
	Acciones     []botsdk.AccionRegistrada `json:"acciones"`
	Comunitarias []botsdk.Carta            `json:"comunitarias"`
	Repartos     []botsdk.Reparto          `json:"repartos"`
	Mostradas    map[string]botsdk.Mano    `json:"mostradas,omitempty"`
	Descripcion  map[string]string         `json:"descripcion,omitempty"`
}

func abrirRegistroManos(ruta string) (*registroManos, error) {
	archivo, err := os.OpenFile(ruta, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &registroManos{archivo: archivo}, nil
}

func (r *registroManos) cerrar() {
	if r != nil {
		r.archivo.Close()
	}
}

// manoInicio anota quiénes están y con cuántas fichas.
func (r *registroManos) manoInicio(e botsdk.EstadoPublico, mias botsdk.Mano) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.idMano, r.mias, r.ultimo = e.IDMano, mias, e
	r.inicio = r.inicio[:0]
	for _, j := range e.Jugadores {
		// Las ciegas ya están puestas en mano_inicio: se devuelven al saldo
		// para registrar las fichas con que cada uno empezó la mano.
		r.inicio = append(r.inicio, jugadorRegistrado{ID: j.ID, Saldo: j.Saldo + j.ApuestaRonda})
	}
}

// observar se queda con el estado más completo de la mano en curso.
func (r *registroManos) observar(e botsdk.EstadoPublico) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if e.IDMano != r.idMano {
		return
	}
	if len(e.HistorialAcciones) >= len(r.ultimo.HistorialAcciones) {
		r.ultimo = e
	}
}

// manoFin escribe la mano. Un error de disco no se propaga: perder un
// registro no puede costar una jugada.
func (r *registroManos) manoFin(res botsdk.ResultadoMano, fase, yo string) error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if res.IDMano != r.idMano {
		return nil // Entramos con la mano empezada: no hay con qué armarla.
	}
	linea, err := json.Marshal(manoRegistrada{
		Hora: time.Now(), IDMano: res.IDMano, Fase: fase, Yo: yo, MisCartas: r.mias,
		CiegaGrande: r.ultimo.CiegaGrande, Boton: r.ultimo.PosicionBoton,
		Jugadores: r.inicio, Acciones: r.ultimo.HistorialAcciones,
		Comunitarias: res.Comunitarias, Repartos: res.Repartos, Mostradas: res.Mostradas, Descripcion: res.Descripcion,
	})
	if err != nil {
		return err
	}
	_, err = r.archivo.Write(append(linea, '\n'))
	return err
}
