package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/entrenamiento"
)

// configLiga es todo lo que necesita una liga (sale de los flags).
type configLiga struct {
	N             int
	Tipos         []string
	Comandos      map[string][]string // Comando de cada tipo de bot del kit.
	ExeLince      string
	ArgsLince     []string // Flags extra para Lince (-genoma, -genoma-final).
	Rondas        int
	Manos         int
	Clasifican    int
	Paralelo      int
	TopeFinal     int
	RelojMesa     time.Duration
	RelojFinal    time.Duration
	TimeoutMs     uint64
	EsperaBots    time.Duration
	EsperaLince   time.Duration
	DireccionFija string // Donde juega siempre Lince.
	SinFinal      bool
	Salida        string
}

const stackInicial = 10000

// rondaLiga es una ronda de la clasificatoria.
type rondaLiga struct {
	Numero      int             `json:"numero"`
	Mesas       []resultadoMesa `json:"mesas"`
	VivosAntes  int             `json:"vivos_antes"`
	VivosTras   int             `json:"vivos_tras"`
	SumaFichas  int64           `json:"suma_fichas"`
	FichasLince int64           `json:"fichas_lince"` // Al terminar la ronda.
	TimeoutsLin uint64          `json:"timeouts_lince"`
	DuracionSeg float64         `json:"duracion_seg"`
}

// puestoLiga es una fila de la clasificación final.
type puestoLiga struct {
	Puesto       int    `json:"puesto"`
	ID           string `json:"id"`
	Fichas       int64  `json:"fichas"`
	RondaQuiebra int    `json:"ronda_quiebra,omitempty"`
}

// informeLiga son los datos crudos de una liga (es lo que va al JSON).
type informeLiga struct {
	Numero         int            `json:"numero"`
	Semilla        uint64         `json:"semilla"`
	Participantes  []participante `json:"participantes"`
	Rondas         []rondaLiga    `json:"rondas"`
	Clasificacion  []puestoLiga   `json:"clasificacion"`
	Finalistas     []string       `json:"finalistas"`
	Final          *resultadoMesa `json:"final,omitempty"`
	GanadorFinal   string         `json:"ganador_final,omitempty"`
	RondaQuiebraLi int            `json:"ronda_quiebra_lince"` // 0 = no quebró.
	PuestoLince    int            `json:"puesto_lince"`
	LinceEnFinal   bool           `json:"lince_en_final"`
	LinceGanoFinal bool           `json:"lince_gano_final"`
	TimeoutsLince  uint64         `json:"timeouts_lince"`
	Chequeo        chequeoLince   `json:"chequeo_lince"`
	Problemas      []string       `json:"problemas,omitempty"`
	DuracionSeg    float64        `json:"duracion_seg"`
}

// jugarLiga juega una clasificatoria completa y la final con procesos reales.
func jugarLiga(ctx context.Context, cfg configLiga, numero int, semilla uint64, progreso *log.Logger) (*informeLiga, error) {
	inicio := time.Now()
	dir := filepath.Join(cfg.Salida, fmt.Sprintf("liga-%02d", numero))
	dirDatos := filepath.Join(dir, "lince-datos")
	// Carpeta propia del ensayo: se vacía para que Lince no retome un estado viejo.
	if err := os.RemoveAll(dirDatos); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dirDatos, 0o755); err != nil {
		return nil, err
	}
	archivoBots, err := os.Create(filepath.Join(dir, "bots.log"))
	if err != nil {
		return nil, err
	}
	defer archivoBots.Close()
	salidaBots := &escrituraSegura{w: archivoBots}
	archivoServ, err := os.Create(filepath.Join(dir, "mesas.log"))
	if err != nil {
		return nil, err
	}
	defer archivoServ.Close()
	logServidor := log.New(archivoServ, "", log.LstdFlags|log.Lmicroseconds)

	campo := armarCampo(cfg.N, cfg.Tipos)
	inf := &informeLiga{Numero: numero, Semilla: semilla, Participantes: campo}
	n := len(campo)

	argsLince := append([]string{
		"-addr", cfg.DireccionFija,
		"-token", tokenLince,
		"-fase", "auto",
		"-silencioso",
		"-datos", dirDatos,
		"-semilla", strconv.FormatUint(semilla%1000003+1, 10),
	}, cfg.ArgsLince...)
	lince, err := lanzarLince(cfg.ExeLince, argsLince, filepath.Join(dir, "lince-salida.log"))
	if err != nil {
		return nil, err
	}
	defer lince.Detener()

	fichas := make([]int64, n)
	for i := range fichas {
		fichas[i] = stackInicial
	}
	quiebras := make([]marcaQuiebra, n)
	indice := make(map[string]int, n)
	for i, p := range campo {
		indice[p.ID] = i
	}
	var jugadasLince []mesaDeLince
	total := int64(n) * stackInicial

	for ronda := 1; ronda <= cfg.Rondas; ronda++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		inicioRonda := time.Now()
		var vivos []int
		for i, f := range fichas {
			if f > 0 {
				vivos = append(vivos, i)
			}
		}
		rl := rondaLiga{Numero: ronda, VivosAntes: len(vivos)}
		azar := rand.New(rand.NewPCG(mezclar(semilla, uint64(ronda)), 0x4C494741))
		azar.Shuffle(len(vivos), func(a, b int) { vivos[a], vivos[b] = vivos[b], vivos[a] })

		type tarea struct {
			cfg      configMesaReal
			asientos []asiento
			conLince bool
		}
		var tareas []tarea
		desde := 0
		for m, tam := range entrenamiento.TamanosMesas(len(vivos)) {
			sentados := vivos[desde : desde+tam]
			desde += tam
			t := tarea{cfg: configMesaReal{
				ID:             fmt.Sprintf("L%02d-R%02d-M%d", numero, ronda, m+1),
				Direccion:      "127.0.0.1:0",
				Manos:          int64(cfg.Manos),
				Reloj:          cfg.RelojMesa,
				TimeoutMs:      cfg.TimeoutMs,
				EsperaConexion: cfg.EsperaBots,
				Semilla:        mezclar(semilla, uint64(ronda), uint64(m)),
				CiegaChica:     50,
				CiegaGrande:    100,
			}}
			for _, i := range sentados {
				a := asiento{ID: campo[i].ID, Fichas: fichas[i]}
				if campo[i].EsLince {
					a.Token = tokenLince
					t.conLince = true
					t.cfg.Direccion = cfg.DireccionFija
					t.cfg.EsperaConexion = cfg.EsperaLince
				} else {
					a.Comando = cfg.Comandos[campo[i].Tipo]
				}
				t.asientos = append(t.asientos, a)
			}
			tareas = append(tareas, t)
		}

		resultados := make([]resultadoMesa, len(tareas))
		cupos := make(chan struct{}, max(cfg.Paralelo, 1))
		var espera sync.WaitGroup
		for k, t := range tareas {
			espera.Add(1)
			cupos <- struct{}{}
			go func(k int, t tarea) {
				defer espera.Done()
				defer func() { <-cupos }()
				resultados[k] = jugarMesaReal(ctx, t.cfg, ronda, t.asientos, salidaBots, logServidor)
				r := resultados[k]
				progreso.Printf("liga %d · %s: %d manos en %.0f s%s", numero, r.ID, r.Manos, r.DuracionSeg, notaMesa(r))
			}(k, t)
		}
		espera.Wait()

		for k, r := range resultados {
			for id, f := range r.FichasFin {
				fichas[indice[id]] = f
			}
			for id, q := range r.Quiebras {
				quiebras[indice[id]] = q
			}
			if tareas[k].conLince {
				jugadasLince = append(jugadasLince, mesaDeLince{Jugadores: len(r.Jugadores), FichasLince: r.FichasInicio[idLince]})
				rl.TimeoutsLin = r.Timeouts[idLince]
				inf.TimeoutsLince += r.Timeouts[idLince]
				if len(r.Ausentes) > 0 {
					for _, a := range r.Ausentes {
						if a == idLince {
							inf.Problemas = append(inf.Problemas, fmt.Sprintf("ronda %d: Lince no se conectó a tiempo a %s", ronda, r.ID))
						}
					}
				}
			}
			if r.Error != "" {
				inf.Problemas = append(inf.Problemas, fmt.Sprintf("%s: %s", r.ID, r.Error))
			}
			for _, a := range r.Ausentes {
				if a != idLince {
					inf.Problemas = append(inf.Problemas, fmt.Sprintf("%s: %s no se conectó", r.ID, a))
				}
			}
		}
		rl.Mesas = resultados
		for _, f := range fichas {
			rl.SumaFichas += f
			if f > 0 {
				rl.VivosTras++
			}
		}
		if rl.SumaFichas != total {
			inf.Problemas = append(inf.Problemas, fmt.Sprintf("ronda %d: la suma de fichas es %d (debería ser %d)", ronda, rl.SumaFichas, total))
		}
		rl.FichasLince = fichas[0]
		rl.DuracionSeg = time.Since(inicioRonda).Seconds()
		inf.Rondas = append(inf.Rondas, rl)
		progreso.Printf("liga %d · ronda %d terminada: %d con fichas, Lince %d, suma %d (%.0f s)", numero, ronda, rl.VivosTras, fichas[0], rl.SumaFichas, rl.DuracionSeg)

		if fichas[0] == 0 && inf.RondaQuiebraLi == 0 && len(jugadasLince) > 0 {
			inf.RondaQuiebraLi = ronda
			lince.Detener() // Eliminado: se termina su proceso.
		}
		if !lince.Vivo() && fichas[0] > 0 {
			inf.Chequeo.ProcesoMurioSol = true
		}
	}

	// Clasificación final, con el criterio del simulador.
	sorteo := rand.New(rand.NewPCG(mezclar(semilla, 0xD15E), 0x4C494741)).Perm(n)
	orden := ordenarClasificacion(fichas, quiebras, sorteo)
	for puesto, i := range orden {
		inf.Clasificacion = append(inf.Clasificacion, puestoLiga{Puesto: puesto + 1, ID: campo[i].ID, Fichas: fichas[i], RondaQuiebra: quiebras[i].Ronda})
		if campo[i].EsLince {
			inf.PuestoLince = puesto + 1
		}
	}
	clasifican := min(cfg.Clasifican, n)
	for _, i := range orden[:clasifican] {
		inf.Finalistas = append(inf.Finalistas, campo[i].ID)
		if campo[i].EsLince {
			inf.LinceEnFinal = true
		}
	}

	if !cfg.SinFinal && len(inf.Finalistas) >= 2 && ctx.Err() == nil {
		finalistas := append([]int(nil), orden[:clasifican]...)
		azar := rand.New(rand.NewPCG(mezclar(semilla, 0xF1A1), 0x4C494741))
		azar.Shuffle(len(finalistas), func(a, b int) { finalistas[a], finalistas[b] = finalistas[b], finalistas[a] })
		cfgFinal := configMesaReal{
			ID:             fmt.Sprintf("L%02d-final", numero),
			Direccion:      "127.0.0.1:0",
			Manos:          int64(cfg.TopeFinal),
			Reloj:          cfg.RelojFinal,
			TimeoutMs:      cfg.TimeoutMs,
			EsperaConexion: cfg.EsperaBots,
			Semilla:        mezclar(semilla, 0xF1A1),
			CiegaChica:     50,
			CiegaGrande:    100,
		}
		// Si Lince quebró pero igual clasifica (quedan menos de Clasifican con
		// fichas y pasan los últimos eliminados), en el torneo seguiría conectado:
		// aquí su proceso ya se terminó, así que se lo relanza con la misma
		// carpeta de datos (retoma su estado como tras un reinicio).
		if inf.LinceEnFinal && !lince.Vivo() {
			nuevo, err := lanzarLince(cfg.ExeLince, argsLince, filepath.Join(dir, "lince-salida-final.log"))
			if err != nil {
				inf.Problemas = append(inf.Problemas, "no se pudo relanzar a Lince para la final: "+err.Error())
			} else {
				lince = nuevo
				defer nuevo.Detener()
				inf.Problemas = append(inf.Problemas, "Lince clasificó habiendo quebrado: se relanzó su proceso para la final")
			}
		}
		var asientos []asiento
		for _, i := range finalistas {
			a := asiento{ID: campo[i].ID, Fichas: stackInicial}
			if campo[i].EsLince {
				a.Token = tokenLince
				cfgFinal.Direccion = cfg.DireccionFija
				cfgFinal.EsperaConexion = cfg.EsperaLince
			} else {
				a.Comando = cfg.Comandos[campo[i].Tipo]
			}
			asientos = append(asientos, a)
		}
		progreso.Printf("liga %d · mesa final: %v", numero, inf.Finalistas)
		r := jugarMesaReal(ctx, cfgFinal, cfg.Rondas+1, asientos, salidaBots, logServidor)
		progreso.Printf("liga %d · %s: %d manos en %.0f s%s", numero, r.ID, r.Manos, r.DuracionSeg, notaMesa(r))
		inf.Final = &r
		if len(r.Posiciones) > 0 {
			inf.GanadorFinal = r.Posiciones[0]
			inf.LinceGanoFinal = inf.GanadorFinal == idLince
		}
		if r.Error != "" {
			inf.Problemas = append(inf.Problemas, fmt.Sprintf("%s: %s", r.ID, r.Error))
		}
		var suma int64
		for _, f := range r.FichasFin {
			suma += f
		}
		if want := int64(len(asientos)) * stackInicial; suma != want {
			inf.Problemas = append(inf.Problemas, fmt.Sprintf("final: la suma de fichas es %d (debería ser %d)", suma, want))
		}
		if inf.LinceEnFinal {
			jugadasLince = append(jugadasLince, mesaDeLince{Final: true, Jugadores: len(asientos), FichasLince: stackInicial})
			inf.TimeoutsLince += r.Timeouts[idLince]
			for _, a := range r.Ausentes {
				if a == idLince {
					inf.Problemas = append(inf.Problemas, "final: Lince no se conectó a tiempo")
				}
			}
		}
	}

	murio := inf.Chequeo.ProcesoMurioSol || (inf.RondaQuiebraLi == 0 && !lince.Vivo())
	lince.Detener()
	inf.Chequeo = chequear(dirDatos, jugadasLince)
	inf.Chequeo.ProcesoMurioSol = murio
	if murio {
		inf.Problemas = append(inf.Problemas, "el proceso de Lince terminó solo antes de tiempo (ver lince-salida.log)")
	}
	inf.DuracionSeg = time.Since(inicio).Seconds()

	datos, err := json.MarshalIndent(inf, "", "  ")
	if err == nil {
		err = os.WriteFile(filepath.Join(dir, "liga.json"), datos, 0o644)
	}
	if err != nil {
		progreso.Printf("no se pudo escribir liga.json: %v", err)
	}
	return inf, nil
}

func notaMesa(r resultadoMesa) string {
	nota := ""
	if r.CerradaPorReloj {
		nota += " · cerrada por reloj"
	}
	if len(r.Ausentes) > 0 {
		nota += fmt.Sprintf(" · ausentes %v", r.Ausentes)
	}
	if r.Error != "" {
		nota += " · ERROR: " + r.Error
	}
	return nota
}

// escrituraSegura serializa las escrituras de varios procesos al mismo archivo.
type escrituraSegura struct {
	mu sync.Mutex
	w  io.Writer
}

func (e *escrituraSegura) Write(p []byte) (int, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.w.Write(p)
}
