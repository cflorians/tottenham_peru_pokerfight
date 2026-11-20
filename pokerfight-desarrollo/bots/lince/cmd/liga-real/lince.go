package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/torneo"
)

// Lince corre como en el torneo: UN proceso para toda la liga, sin
// -una-partida, que reconecta solo. Cuando su mesa de una ronda cierra, la
// conexión se corta y Lince reintenta cada pocos segundos contra la misma
// dirección; la mesa de la ronda siguiente donde le toca jugar escucha en esa
// dirección, y la mesa lo reconoce por su token fijo (siempre el mismo id).

// tokenLince es el token de ensayo (no el del equipo: la liga es local).
const tokenLince = "lince-ensayo"

type procesoLince struct {
	cmd       *exec.Cmd
	terminado chan struct{}
	unaVez    sync.Once
	salida    io.Closer
}

// lanzarLince arranca el proceso. La variable POKERFIGHT_TOKEN se quita del
// entorno: el token del equipo no tiene nada que hacer en una liga local.
func lanzarLince(exe string, args []string, rutaSalida string) (*procesoLince, error) {
	archivo, err := os.Create(rutaSalida)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(exe, args...)
	cmd.Stdout = archivo
	cmd.Stderr = archivo
	for _, v := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(v), "POKERFIGHT_TOKEN=") {
			cmd.Env = append(cmd.Env, v)
		}
	}
	if err := cmd.Start(); err != nil {
		archivo.Close()
		return nil, fmt.Errorf("no se pudo ejecutar %s: %w", exe, err)
	}
	p := &procesoLince{cmd: cmd, terminado: make(chan struct{}), salida: archivo}
	go func() {
		cmd.Wait()
		close(p.terminado)
	}()
	return p, nil
}

// Vivo dice si el proceso sigue corriendo.
func (p *procesoLince) Vivo() bool {
	select {
	case <-p.terminado:
		return false
	default:
		return true
	}
}

// Detener mata el proceso (Lince no termina solo: reconecta para siempre) y
// espera a que salga. Es idempotente.
func (p *procesoLince) Detener() {
	p.unaVez.Do(func() {
		if p.Vivo() && p.cmd.Process != nil {
			p.cmd.Process.Kill()
		}
		select {
		case <-p.terminado:
		case <-time.After(10 * time.Second):
		}
		p.salida.Close()
	})
}

// esperarGuardado da tiempo a que Lince cierre la última mano liquidada.
// El servidor ya envió mano_fin, pero Lince todavía puede estar escribiendo
// el temporal que luego se publica como estado.json.
func (p *procesoLince) esperarGuardado(ctx context.Context, dirDatos, mano string) bool {
	if mano == "" {
		return false
	}
	ctx, cancelar := context.WithTimeout(ctx, 2*time.Second)
	defer cancelar()
	tic := time.NewTicker(25 * time.Millisecond)
	defer tic.Stop()
	for {
		if estadoConfirmaMano(dirDatos, mano) {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-p.terminado:
			return false
		case <-tic.C:
		}
	}
}

func estadoConfirmaMano(dirDatos, mano string) bool {
	estado, err := leerEstadoLince(dirDatos)
	return err == nil && estado.IDJugador == idLince && manoGuardada(estado) == mano
}

func leerEstadoLince(dirDatos string) (torneo.Estado, error) {
	var estado torneo.Estado
	datos, err := os.ReadFile(filepath.Join(dirDatos, "estado.json"))
	if err == nil {
		err = json.Unmarshal(datos, &estado)
	}
	return estado, err
}

func manoGuardada(e torneo.Estado) string {
	if e.UltimaManoFin.IsZero() {
		return ""
	}
	return e.UltimaMano
}

// --- Chequeo del ensayo -------------------------------------------------------

// chequeoLince compara lo que Lince cree que jugó (su estado.json) con lo que
// de verdad pasó, y cuenta los problemas de su log.
type chequeoLince struct {
	EstadoLeido     bool     `json:"estado_leido"`
	IDJugador       string   `json:"id_jugador"`
	Ronda           int      `json:"ronda"`
	Fase            string   `json:"fase"`
	FasesPorRonda   []string `json:"fases_por_ronda"`
	RondaEsperada   int      `json:"ronda_esperada"`
	FaseEsperada    string   `json:"fase_esperada"`
	Coincide        bool     `json:"coincide"`
	Diferencias     []string `json:"diferencias,omitempty"`
	Panicos         int      `json:"panicos"`
	Errores         int      `json:"errores"`
	EjemplosLog     []string `json:"ejemplos_log,omitempty"`
	ProcesoMurioSol bool     `json:"proceso_murio_solo,omitempty"`

	UltimaManoEsperada string `json:"ultima_mano_esperada,omitempty"`
	UltimaManoGuardada string `json:"ultima_mano_guardada,omitempty"`
}

// mesaDeLince es una mesa que jugó Lince, para el chequeo.
type mesaDeLince struct {
	Final       bool
	Jugadores   int
	FichasLince int64
	UltimaMano  string
}

func ultimaManoEsperada(jugadas []mesaDeLince) string {
	if len(jugadas) == 0 {
		return ""
	}
	return jugadas[len(jugadas)-1].UltimaMano
}

// chequear lee el estado y el log de la carpeta de datos de Lince.
func chequear(dirDatos string, jugadas []mesaDeLince) chequeoLince {
	ch := chequeoLince{
		RondaEsperada: len(jugadas), FaseEsperada: string(torneo.Clasificatoria),
		UltimaManoEsperada: ultimaManoEsperada(jugadas),
	}
	if n := len(jugadas); n > 0 && jugadas[n-1].Final {
		ch.FaseEsperada = string(torneo.Final)
	}

	e, err := leerEstadoLince(dirDatos)
	if err != nil {
		ch.Diferencias = append(ch.Diferencias, "no se pudo leer estado.json: "+err.Error())
	} else {
		ch.EstadoLeido = true
		ch.IDJugador, ch.Ronda, ch.Fase = e.IDJugador, e.Ronda, string(e.Fase)
		ch.UltimaManoGuardada = manoGuardada(e)
		if ch.Fase == "" {
			ch.Fase = string(torneo.Clasificatoria)
		}
		for _, r := range e.Rondas {
			ch.FasesPorRonda = append(ch.FasesPorRonda, string(r.Fase))
		}
		if e.IDJugador != idLince {
			ch.Diferencias = append(ch.Diferencias, fmt.Sprintf("id_jugador %q (esperado %q)", e.IDJugador, idLince))
		}
		if e.Ronda != ch.RondaEsperada {
			ch.Diferencias = append(ch.Diferencias, fmt.Sprintf("cree estar en la ronda %d y jugó %d mesas", e.Ronda, ch.RondaEsperada))
		}
		if ch.Fase != ch.FaseEsperada {
			ch.Diferencias = append(ch.Diferencias, fmt.Sprintf("cree estar en la fase %s y la última mesa fue de la %s", ch.Fase, ch.FaseEsperada))
		}
		for i, j := range jugadas {
			if i >= len(e.Rondas) {
				break
			}
			r := e.Rondas[i]
			esperada := torneo.Clasificatoria
			if j.Final {
				esperada = torneo.Final
			}
			if r.Fase != esperada {
				ch.Diferencias = append(ch.Diferencias, fmt.Sprintf("ronda %d anotada como %s (era %s)", i+1, r.Fase, esperada))
			}
			if r.StackInicio != j.FichasLince {
				ch.Diferencias = append(ch.Diferencias, fmt.Sprintf("ronda %d: anotó %d fichas al empezar (eran %d)", i+1, r.StackInicio, j.FichasLince))
			}
			if r.Jugadores != j.Jugadores {
				ch.Diferencias = append(ch.Diferencias, fmt.Sprintf("ronda %d: anotó %d jugadores con fichas (eran %d)", i+1, r.Jugadores, j.Jugadores))
			}
		}
	}
	if ch.UltimaManoEsperada != "" && ch.UltimaManoGuardada != ch.UltimaManoEsperada {
		ch.Diferencias = append(ch.Diferencias, fmt.Sprintf("última mano guardada %q (esperada %q)",
			ch.UltimaManoGuardada, ch.UltimaManoEsperada))
	}

	logs, _ := filepath.Glob(filepath.Join(dirDatos, "lince-*.log"))
	for _, ruta := range logs {
		f, err := os.Open(ruta)
		if err != nil {
			continue
		}
		escaner := bufio.NewScanner(f)
		escaner.Buffer(make([]byte, 64<<10), 4<<20)
		for escaner.Scan() {
			linea := escaner.Text()
			esPanico := strings.Contains(linea, "PÁNICO")
			esError := strings.Contains(linea, "ERROR")
			if esPanico {
				ch.Panicos++
			}
			if esError {
				ch.Errores++
			}
			if (esPanico || esError) && len(ch.EjemplosLog) < 5 {
				ch.EjemplosLog = append(ch.EjemplosLog, linea)
			}
		}
		f.Close()
	}
	ch.Coincide = len(ch.Diferencias) == 0
	return ch
}
