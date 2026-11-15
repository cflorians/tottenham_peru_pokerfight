// liga-real juega una clasificatoria completa (rondas con fichas arrastradas y
// mesas sorteadas de nuevo en cada ronda) y la mesa final con los ejecutables
// de verdad: el servidor de mesa del kit, los bots del kit como subprocesos y
// UN solo proceso de Lince que sigue conectado (reconectando) toda la liga,
// como en el torneo.
//
// Sirve para dos cosas: ensayo general del bot (que aguante las rondas, las
// cuente bien y detecte la final) y comparar con el simulador rápido
// (bots/lince/entrenamiento/liga.go, del que copia las reglas y supuestos).
//
// Desde pokerfight-desarrollo, con los bots compilados en bin/:
//
//	go build -o bin/ ./bots/lince/cmd/liga-real
//	./bin/liga-real                                  # una liga de 20 con los valores de las bases
//	./bin/liga-real -n 10 -rondas 3 -manos 20        # liga corta de humo
//	./bin/liga-real -ligas 5 -semilla 7 > informe.md # varias ligas y resumen agregado
//
// El informe (markdown) va por stdout; el progreso, por stderr. En -salida
// queda una carpeta por liga con liga.json (datos crudos), la carpeta de datos
// de Lince (estado.json, log, manos), la salida de los bots y el log de las mesas.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func main() {
	var (
		n          = flag.Int("n", 20, "participantes de la liga (Lince más n-1 bots del kit)")
		tipos      = flag.String("tipos", "darwin,espia,estadistico,base,conservador,aleatorio,novato", "tipos de bots del kit que completan el campo, repartidos de forma pareja")
		rondas     = flag.Int("rondas", 10, "rondas de la clasificatoria")
		manos      = flag.Int("manos", 50, "manos por mesa de la clasificatoria")
		clasifican = flag.Int("clasifican", 6, "cuántos pasan a la mesa final")
		semilla    = flag.Uint64("semilla", 1, "semilla de la primera liga (sorteo de mesas y cartas); la liga k usa semilla+k-1")
		ligas      = flag.Int("ligas", 1, "cuántas ligas seguidas")
		paralelo   = flag.Int("paralelo", 3, "mesas de una ronda que se juegan a la vez")
		topeFinal  = flag.Int("tope-final", 2000, "tope de manos de la mesa final (al llegar gana el que tiene más fichas)")
		relojFinal = flag.Duration("reloj-final", 0, "tope de reloj de la mesa final (0 = sin tope; las bases dicen 1h)")
		relojMesa  = flag.Duration("reloj-mesa", 6*time.Minute, "tope de reloj de cada mesa de la clasificatoria")
		timeoutMs  = flag.Uint64("timeout-ms", 5000, "plazo por jugada")
		esperaBots = flag.Duration("espera-bots", 30*time.Second, "cuánto se espera a que se conecten los bots del kit")
		esperaLin  = flag.Duration("espera-lince", 60*time.Second, "cuánto se espera a que Lince se (re)conecte a su mesa")
		puerto     = flag.Int("puerto-lince", 19500, "puerto local fijo donde juega Lince en todas las rondas")
		sinFinal   = flag.Bool("sin-final", false, "no jugar la mesa final")
		salida     = flag.String("salida", filepath.Join("bots", "lince", "datos", "liga-real"), "carpeta donde se deja todo")
		genoma     = flag.String("genoma", "", "genoma de la clasificatoria para Lince (por defecto, el incluido)")
		genomaFin  = flag.String("genoma-final", "", "genoma de la final para Lince (por defecto, el incluido)")
		dirBin     = flag.String("bin", "bin", "carpeta de los ejecutables")
		python     = flag.String("python", "python", "intérprete para el bot novato")
	)
	flag.Parse()
	progreso := log.New(os.Stderr, "liga-real: ", log.Ltime)

	listaTipos := partirLista(*tipos)
	if *n < 2 {
		progreso.Fatalf("-n debe ser al menos 2")
	}
	comandos := map[string][]string{}
	for _, t := range listaTipos {
		c, ok := comandoDe(t, *dirBin, *python)
		if !ok {
			progreso.Fatalf("tipo de bot desconocido: %q (conocidos: darwin, espia, estadistico, base, conservador, aleatorio, bandido, novato)", t)
		}
		comandos[t] = c
	}
	exeLince := ejecutable(*dirBin, "lince")
	// Antes de jugar, que todos los ejecutables arranquen (Smart App Control
	// puede bloquear uno recién compilado).
	if err := comprobarArranque(exeLince); err != nil {
		progreso.Fatalf("lince no arranca: %v", err)
	}
	for _, t := range listaTipos {
		if err := comprobarArranque(comandos[t][0], comandos[t][1:]...); err != nil {
			progreso.Fatalf("%s no arranca: %v", t, err)
		}
	}

	var argsLince []string
	if *genoma != "" {
		argsLince = append(argsLince, "-genoma", *genoma)
	}
	if *genomaFin != "" {
		argsLince = append(argsLince, "-genoma-final", *genomaFin)
	}
	cfg := configLiga{
		N: *n, Tipos: listaTipos, Comandos: comandos, ExeLince: exeLince, ArgsLince: argsLince,
		Rondas: *rondas, Manos: *manos, Clasifican: *clasifican, Paralelo: *paralelo,
		TopeFinal: *topeFinal, RelojMesa: *relojMesa, RelojFinal: *relojFinal, TimeoutMs: *timeoutMs,
		EsperaBots: *esperaBots, EsperaLince: *esperaLin,
		DireccionFija: fmt.Sprintf("127.0.0.1:%d", *puerto), SinFinal: *sinFinal, Salida: *salida,
	}

	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	fmt.Printf("# Liga real: %d participantes, %d rondas de %d manos, semilla %d\n\n", cfg.N, cfg.Rondas, cfg.Manos, *semilla)
	fmt.Printf("Campo: Lince + %d bots del kit (%s). Reloj por mesa %s, plazo %d ms, %d mesas en paralelo.\n\n",
		cfg.N-1, strings.Join(listaTipos, ", "), cfg.RelojMesa, cfg.TimeoutMs, cfg.Paralelo)

	inicio := time.Now()
	var informes []*informeLiga
	for k := 1; k <= *ligas; k++ {
		inf, err := jugarLiga(ctx, cfg, k, *semilla+uint64(k-1), progreso)
		if err != nil {
			progreso.Printf("liga %d: %v", k, err)
			break
		}
		informes = append(informes, inf)
		imprimirLiga(os.Stdout, cfg, inf)
	}
	if len(informes) > 1 {
		imprimirAgregado(os.Stdout, cfg, informes)
	}
	fmt.Printf("Tiempo total: %s.\n", time.Since(inicio).Round(time.Second))
}

func partirLista(s string) []string {
	var lista []string
	for _, t := range strings.Split(s, ",") {
		if t = strings.TrimSpace(t); t != "" {
			lista = append(lista, t)
		}
	}
	return lista
}

func ejecutable(dir, nombre string) string {
	ruta := filepath.Join(dir, nombre)
	if runtime.GOOS == "windows" {
		ruta += ".exe"
	}
	return ruta
}

// comandoDe es el comando de cada tipo de bot del kit (los mismos de
// bots/lince/pruebas/mesas.py). La mesa le agrega -addr y -token.
func comandoDe(tipo, dir, python string) ([]string, bool) {
	switch tipo {
	case "darwin", "espia":
		return []string{ejecutable(dir, tipo), "-silencioso"}, true
	case "estadistico":
		return []string{ejecutable(dir, tipo), "-simulaciones", "2000"}, true
	case "base", "conservador", "aleatorio", "bandido":
		return []string{ejecutable(dir, tipo)}, true
	case "novato":
		return []string{python, filepath.Join("bots", "novato", "bot.py"), "-silencioso"}, true
	}
	return nil, false
}

// comprobarArranque corre el programa con -h (o --version para Python) y
// solo falla si no se puede ejecutar: el código de salida de -h no importa.
func comprobarArranque(exe string, args ...string) error {
	ctx, cancelar := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelar()
	var cmd *exec.Cmd
	if strings.HasSuffix(strings.ToLower(filepath.Base(exe)), "python") || strings.HasPrefix(strings.ToLower(filepath.Base(exe)), "python") {
		cmd = exec.CommandContext(ctx, exe, "--version")
	} else {
		cmd = exec.CommandContext(ctx, exe, "-h")
	}
	if err := cmd.Run(); err != nil {
		if _, esSalida := err.(*exec.ExitError); esSalida && ctx.Err() == nil {
			return nil
		}
		return err
	}
	return nil
}
