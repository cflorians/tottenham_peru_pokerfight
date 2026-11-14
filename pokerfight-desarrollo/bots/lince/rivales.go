package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aicspucp/pokerfight-arena-participantes/pkg/botsdk"

	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/estrategia"
)

// lecturaRivales lleva, para cada rival, cuántas veces apostó o subió y
// cuántas pagó (estrategia.Perfil). Con eso se llena Situacion.LecturaRival
// con el mismo criterio que el simulador (entrenamiento/simulador.go):
//
//   - bet, raise y all-in son agresivas; call es pasiva; check y fold no
//     cuentan. Las ciegas no llegan como acciones (la mesa no las anota en el
//     historial) y un pase o retiro puesto por la mesa ante un timeout cuenta
//     como lo que es: check o fold, nada.
//   - Un all-in cuenta siempre como agresivo, aunque sea un pago corto (no le
//     alcanza para igualar): el historial solo dice "allin", sin el monto, y el
//     simulador lo cuenta igual.
//
// La mesa manda en cada estado el historial completo de la mano en curso (no
// solo de la calle), así que basta con recordar cuántas acciones de esta mano
// ya se contaron y anotar solo las que siguen. Existe siempre, también con
// -sin-archivos: entonces simplemente no se guarda en disco.
type lecturaRivales struct {
	mu       sync.Mutex
	perfiles map[string]*estrategia.Perfil
	yo       string // Último id propio conocido, por si llega un estado sin él.

	idMano   string // Mano cuyas acciones se vienen contando.
	contadas int    // Cuántas acciones de idMano ya se anotaron.

	ruta            string // datos/rivales.json; vacía con -sin-archivos.
	manosSinGuardar int    // Manos terminadas desde el último guardado.
}

// manosEntreGuardados es cada cuántas manos se guardan los perfiles en disco.
// Perder unas pocas manos de lectura tras un reinicio no cambia nada.
const manosEntreGuardados = 5

func nuevaLecturaRivales(ruta string) *lecturaRivales {
	return &lecturaRivales{perfiles: map[string]*estrategia.Perfil{}, ruta: ruta}
}

// observar anota las acciones de la mano en curso que todavía no se contaron.
// Se puede llamar cuantas veces se quiera con el mismo estado, o con uno más
// viejo: nada se cuenta dos veces. Las acciones propias (yo) no se anotan.
func (r *lecturaRivales) observar(e botsdk.EstadoPublico, yo string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if yo != "" {
		r.yo = yo
	}
	if e.IDMano == "" {
		return
	}
	if e.IDMano != r.idMano {
		r.idMano, r.contadas = e.IDMano, 0
	}
	historial := e.HistorialAcciones
	if len(historial) <= r.contadas {
		return // Nada nuevo, o un estado más viejo que lo ya visto.
	}
	for _, a := range historial[r.contadas:] {
		if a.IDJugador == "" || a.IDJugador == r.yo {
			continue
		}
		agresiva := estrategia.EsAgresiva(a.Accion.Tipo)
		pasiva := a.Accion.Tipo == botsdk.Call
		if !agresiva && !pasiva {
			continue
		}
		p := r.perfiles[a.IDJugador]
		if p == nil {
			p = &estrategia.Perfil{}
			r.perfiles[a.IDJugador] = p
		}
		p.Anotar(agresiva, pasiva)
	}
	r.contadas = len(historial)
}

// lectura es estrategia.Perfil.Lectura del rival id: 0 si id es "" o si nunca
// se lo vio poner fichas.
func (r *lecturaRivales) lectura(id string) float64 {
	if id == "" {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	p := r.perfiles[id]
	if p == nil {
		return 0
	}
	return p.Lectura()
}

// --- Persistencia --------------------------------------------------------------

const versionRivales = 1

type perfilGuardado struct {
	Agresivas int `json:"agresivas"`
	Pasivas   int `json:"pasivas"`
}

type rivalesGuardados struct {
	Version  int                       `json:"version"`
	IDMano   string                    `json:"id_mano"`
	Contadas int                       `json:"contadas"`
	Perfiles map[string]perfilGuardado `json:"perfiles"`
}

// manoFin cuenta una mano terminada y guarda los perfiles cada
// manosEntreGuardados manos.
func (r *lecturaRivales) manoFin() error {
	r.mu.Lock()
	r.manosSinGuardar++
	toca := r.manosSinGuardar >= manosEntreGuardados
	r.mu.Unlock()
	if !toca {
		return nil
	}
	return r.guardar()
}

// guardar escribe los perfiles en disco (nada sin ruta). Escribe a un archivo
// temporal y lo renombra: un corte a mitad no deja un JSON roto.
func (r *lecturaRivales) guardar() error {
	r.mu.Lock()
	if r.ruta == "" {
		r.mu.Unlock()
		return nil
	}
	g := rivalesGuardados{Version: versionRivales, IDMano: r.idMano, Contadas: r.contadas, Perfiles: make(map[string]perfilGuardado, len(r.perfiles))}
	for id, p := range r.perfiles {
		g.Perfiles[id] = perfilGuardado{Agresivas: p.Agresivas, Pasivas: p.Pasivas}
	}
	r.manosSinGuardar = 0
	ruta := r.ruta
	r.mu.Unlock()

	datos, err := json.MarshalIndent(g, "", " ")
	if err != nil {
		return err
	}
	temporal := ruta + ".tmp"
	if err := os.WriteFile(temporal, datos, 0o644); err != nil {
		return err
	}
	return os.Rename(temporal, ruta)
}

// leer retoma los perfiles guardados por un proceso anterior. Que no exista el
// archivo no es un error.
func (r *lecturaRivales) leer() error {
	if r.ruta == "" {
		return nil
	}
	datos, err := os.ReadFile(r.ruta)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var g rivalesGuardados
	if err := json.Unmarshal(datos, &g); err != nil {
		return fmt.Errorf("%s ilegible: %w", r.ruta, err)
	}
	if g.Version != versionRivales {
		return fmt.Errorf("%s es de la versión %d (se esperaba %d): se empieza de cero", r.ruta, g.Version, versionRivales)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.idMano, r.contadas = g.IDMano, g.Contadas
	r.perfiles = make(map[string]*estrategia.Perfil, len(g.Perfiles))
	for id, p := range g.Perfiles {
		r.perfiles[id] = &estrategia.Perfil{Agresivas: p.Agresivas, Pasivas: p.Pasivas}
	}
	return nil
}

// cantidad es cuántos rivales tienen perfil (para el log).
func (r *lecturaRivales) cantidad() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.perfiles)
}

// textoLectura es la lectura para la línea de detalle: "" si es 0, y si no,
// algo como " · lectura +0,31".
func textoLectura(l float64) string {
	if l == 0 {
		return ""
	}
	return " · lectura " + strings.Replace(fmt.Sprintf("%+.2f", l), ".", ",", 1)
}
