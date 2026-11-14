// entrenar-liga evoluciona un genoma con lo que premia el torneo, jugando
// torneos enteros en el simulador rápido:
//
//	-modo clasificatoria   aptitud = premio esperado de la clasificatoria (plan A, v1)
//	-modo final            aptitud = probabilidad de ganar la mesa final  (plan B, v2)
//
// No pisa genoma.json: escribe genoma_clasificatoria.json o genoma_final.json
// en la carpeta del bot. Promoverlo (que el bot lo use) es un paso aparte,
// después de medirlo con simular-liga.
//
//	go run ./bots/lince/cmd/entrenar-liga -modo clasificatoria
//	go run ./bots/lince/cmd/entrenar-liga -modo final -clasificatoria bots/lince/genoma_clasificatoria.json
//	go run ./bots/lince/cmd/entrenar-liga -rapido          # prueba de humo, no escribe en la carpeta del bot
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/entrenamiento"
	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/estrategia"
)

func main() {
	log.SetFlags(0)
	var (
		modo         = flag.String("modo", "clasificatoria", "clasificatoria o final")
		poblacion    = flag.Int("poblacion", 48, "individuos por generación")
		generaciones = flag.Int("generaciones", 60, "cantidad de generaciones")
		torneos      = flag.Int("torneos", 40, "por evaluación: ligas por campo (clasificatoria) o finales/10 (final)")
		validacion   = flag.Int("validacion", 500, "ligas por campo (o finales/10) para elegir al campeón con torneos nunca vistos")
		listaN       = flag.String("n", "15,20,30", "inscritos por liga; se alternan (modo clasificatoria)")
		listaTopes   = flag.String("topes", "150,500,0", "topes de manos de la final; se alternan (modo final)")
		equilibrado  = flag.Bool("equilibrado", false, "modo final: la mitad de las finales contra finalistas tipo kit y la otra mitad contra los genomas de la carpeta rivales; la aptitud es el peor de los dos resultados")
		conFinal     = flag.Bool("con-final", false, "modo clasificatoria: cada liga en que clasifica juega también su mesa final (con -topes) y la aptitud es el premio realmente cobrado, en vez de suponer -m2")
		m2           = flag.Float64("m2", 0.20, "probabilidad de ganar la final que se supone al valorar llegar a ella (modo clasificatoria)")
		forma        = flag.Float64("forma", 3, "peso del puesto en la liga dentro de la aptitud de entrenamiento (modo clasificatoria)")
		coevolucion  = flag.Int("coevolucion", 2, "rivales que se reemplazan por campeones de generaciones anteriores")
		tamBib       = flag.Int("biblioteca", 20000, "repartos precalculados")
		muestras     = flag.Int("muestras", 200, "muestras de Monte Carlo por calle al precalcular la fuerza")
		semilla      = flag.Uint64("semilla", 2026, "semilla: misma semilla, mismo resultado")
		dir          = flag.String("dir", dirDelBot(), "carpeta del bot (genomas y caché)")
		salida       = flag.String("salida", "", "dónde guardar el campeón, relativo a -dir (por defecto genoma_<modo>.json)")
		rutaClasif   = flag.String("clasificatoria", "", "genoma de clasificatoria, para sentarlo como rival en la final (por defecto, genoma.json)")
		hilos        = flag.Int("hilos", 0, "núcleos a usar (0 = todos)")
		rapido       = flag.Bool("rapido", false, "prueba de humo: pocas generaciones y el resultado va a una carpeta temporal")
	)
	flag.Parse()
	if *modo != "clasificatoria" && *modo != "final" {
		log.Fatalf("modo desconocido %q", *modo)
	}

	darwin := leerGenoma(filepath.Join(*dir, "genoma_darwin.json"))
	fuertes, err := entrenamiento.LeerRivales(filepath.Join(*dir, "rivales"))
	if err != nil {
		log.Fatalf("rivales: %v", err)
	}
	// genoma_v0.json es el prudente de la v0: se usa como rival y como
	// referencia, y no cambia aunque se promueva otro genoma a genoma.json.
	prudente := leerGenoma(filepath.Join(*dir, "genoma_v0.json"))
	clasificatoria := leerGenoma(filepath.Join(*dir, "genoma.json"))
	if *rutaClasif != "" {
		clasificatoria = leerGenoma(*rutaClasif)
	}

	dirSalida := *dir
	if *rapido {
		*poblacion, *generaciones, *torneos, *validacion = 12, 4, 6, 20
		tmp, err := os.MkdirTemp("", "entrenar-liga-rapido-")
		if err != nil {
			log.Fatal(err)
		}
		dirSalida = tmp
	}
	if *salida == "" {
		*salida = "genoma_" + *modo + ".json"
	}
	rutaSalida := filepath.Join(dirSalida, *salida)

	inicio := time.Now()
	// La biblioteca de entrenamiento es la misma de simular-liga (comparte la
	// caché). La de validación tiene otra semilla: el campeón se elige con
	// cartas que el genético nunca jugó.
	entreno := biblioteca(*dir, "biblioteca-liga.cache", entrenamiento.ConfigBiblioteca{Repartos: *tamBib, Muestras: *muestras, Semilla: *semilla})
	valida := biblioteca(*dir, "biblioteca-validacion.cache", entrenamiento.ConfigBiblioteca{Repartos: *tamBib / 2, Muestras: *muestras, Semilla: *semilla + 77})
	fmt.Printf("bibliotecas listas en %s\n", time.Since(inicio).Round(time.Second))

	var aptitud, aptitudValidacion entrenamiento.AptitudFunc
	var columnas []string
	switch *modo {
	case "clasificatoria":
		campos := entrenamiento.Campos(darwin, prudente, fuertes...)
		a := entrenamiento.AptitudLiga{
			Liga: entrenamiento.LigaDeLasBases(), Campos: campos, Tamanos: enteros(*listaN),
			Ligas: *torneos, M2: *m2, Forma: *forma, Coevolucion: *coevolucion,
		}
		if *conFinal {
			a.TopesFinal = enteros(*listaTopes)
		}
		v := a
		v.Ligas, v.Forma, v.Coevolucion = *validacion, 0, 0
		aptitud, aptitudValidacion = a.Evaluar, v.Evaluar
		for _, c := range campos {
			columnas = append(columnas, "U "+c.Nombre, "M1 "+c.Nombre, "M1b "+c.Nombre)
		}
		columnas = append(columnas, "U", "M1", "M1b")
	case "final":
		topes := enteros(*listaTopes)
		a := entrenamiento.AptitudFinal{
			Rivales: entrenamiento.Finalistas(darwin, prudente, clasificatoria, fuertes...), Topes: topes,
			Finales: *torneos * 10, Coevolucion: *coevolucion,
		}
		if *equilibrado {
			if len(fuertes) == 0 {
				log.Fatal("-equilibrado necesita genomas en la carpeta rivales")
			}
			a.Rivales = entrenamiento.Finalistas(darwin, prudente, clasificatoria)
			// La final "fuerte": casi todos entrenados, con algún bot del kit.
			a.Fuertes = entrenamiento.Campo{Nombre: "fuertes", Tipos: []entrenamiento.TipoBot{{Nombre: "novato", Politica: entrenamiento.Novato{}}, {Nombre: "espia", Politica: entrenamiento.Espia{}}}, Pesos: []float64{0.5, 0.5}}
			for _, g := range fuertes {
				a.Fuertes.Tipos = append(a.Fuertes.Tipos, entrenamiento.TipoBot{Nombre: "entrenado", Politica: g})
				a.Fuertes.Pesos = append(a.Fuertes.Pesos, 4/float64(len(fuertes)))
			}
		}
		v := a
		v.Finales, v.Coevolucion = *validacion*10, 0
		aptitud, aptitudValidacion = a.Evaluar, v.Evaluar
		for _, t := range topes {
			columnas = append(columnas, "M2 % tope "+strconv.Itoa(t))
		}
		if *equilibrado {
			columnas = append(columnas, "M2 % contra kit", "M2 % contra fuertes")
		}
	}

	cfg := entrenamiento.ConfigGenetico{
		Poblacion: *poblacion, Generaciones: *generaciones, Elite: max(2, *poblacion/16),
		TamTorneo: 3, ProbCruce: 0.9, ProbMutacion: 0.15, Sigma: 0.12, Semilla: *semilla, Hilos: *hilos,
	}
	iniciales := []estrategia.Genoma{clasificatoria, prudente, darwin, estrategia.GenomaInicial()}
	iniciales = append(iniciales, fuertes...)

	fmt.Printf("modo %s: población %d, %d generaciones, %d torneos por evaluación\n\n", *modo, cfg.Poblacion, cfg.Generaciones, *torneos)
	fmt.Println("  gen      mejor   promedio       peor   tiempo")
	inicioEvolucion := time.Now()
	poblacionFinal, campeones := entrenamiento.EvolucionarCon(cfg, iniciales, entreno, aptitud, func(g entrenamiento.Generacion) {
		fmt.Printf("  %3d  %9.2f  %9.2f  %9.2f   %s\n", g.Numero, g.Mejor.Evaluacion.Aptitud, g.Promedio, g.Peor, time.Since(inicioEvolucion).Round(time.Second))
	})

	// Candidatos: los mejores de la última población, los campeones de las
	// últimas generaciones y, como referencia, los genomas de partida.
	nombres := []string{"v0 (prudente)", "Darwin original"}
	candidatos := []estrategia.Genoma{prudente, darwin}
	for i := 0; i < len(poblacionFinal) && i < 6; i++ {
		candidatos = append(candidatos, poblacionFinal[i].Genoma)
		nombres = append(nombres, fmt.Sprintf("última población #%d", i+1))
	}
	for i := max(len(campeones)-6, 0); i < len(campeones); i++ {
		candidatos = append(candidatos, campeones[i])
		nombres = append(nombres, fmt.Sprintf("campeón de la generación %d", i+1))
	}
	fmt.Printf("\nvalidación con torneos nunca vistos (%d candidatos)...\n\n", len(candidatos))
	mejor, resultados := entrenamiento.ElegirCampeonCon(candidatos, valida, aptitudValidacion, *semilla+1)
	// Las referencias no compiten: el campeón sale de lo entrenado.
	mejor = 2
	for i := 2; i < len(resultados); i++ {
		if resultados[i].Aptitud > resultados[mejor].Aptitud {
			mejor = i
		}
	}

	fmt.Printf("| Candidato | Aptitud | %s |\n| --- | ---: |%s\n", strings.Join(columnas, " | "), strings.Repeat(" ---: |", len(columnas)))
	for i, r := range resultados {
		marca := ""
		if i == mejor {
			marca = " **(elegido)**"
		}
		fmt.Printf("| %s%s | %.2f |", nombres[i], marca, r.Aptitud)
		for _, v := range r.PorEscenario {
			fmt.Printf(" %.3g |", v)
		}
		fmt.Println()
	}

	campeon := candidatos[mejor]
	archivo := estrategia.NuevoArchivo(campeon, cfg.Generaciones, resultados[mejor].Aptitud, *semilla)
	archivo.Bot = "Lince · genoma de " + *modo
	if err := archivo.Guardar(rutaSalida); err != nil {
		log.Fatalf("no se pudo guardar %s: %v", rutaSalida, err)
	}
	fmt.Printf("\ncampeón guardado en %s\n\n", rutaSalida)
	fmt.Println("| Gen | v0 | Campeón |\n| --- | ---: | ---: |")
	for i, def := range estrategia.Genes {
		fmt.Printf("| `%s` | %.4f | %.4f |\n", def.Nombre, prudente[i], campeon[i])
	}
	fmt.Printf("\ntotal: %s\n", time.Since(inicio).Round(time.Second))
}

func biblioteca(dir, nombre string, cfg entrenamiento.ConfigBiblioteca) entrenamiento.Biblioteca {
	ruta := filepath.Join(dir, "datos", nombre)
	_ = os.MkdirAll(filepath.Dir(ruta), 0o755)
	ultimo := time.Now()
	bib, _, err := entrenamiento.CargarOGenerar(ruta, cfg, func(hechos, total int) {
		if time.Since(ultimo) > 2*time.Second || hechos == total {
			fmt.Fprintf(os.Stderr, "\r  precalculando %s: %d/%d", nombre, hechos, total)
			ultimo = time.Now()
		}
	})
	if err != nil {
		log.Printf("aviso: %v", err)
	}
	return bib
}

func enteros(lista string) []int {
	var valores []int
	for _, texto := range strings.Split(lista, ",") {
		v, err := strconv.Atoi(strings.TrimSpace(texto))
		if err != nil {
			log.Fatalf("número inválido %q", texto)
		}
		valores = append(valores, v)
	}
	return valores
}

func leerGenoma(ruta string) estrategia.Genoma {
	datos, err := os.ReadFile(ruta)
	if err != nil {
		log.Fatalf("no se pudo leer el genoma: %v", err)
	}
	g, _, err := estrategia.LeerGenoma(datos)
	if err != nil {
		log.Fatalf("genoma ilegible %s: %v", ruta, err)
	}
	return g
}

func dirDelBot() string {
	enRepo := filepath.Join("bots", "lince")
	if _, err := os.Stat(filepath.Join(enRepo, "genoma.json")); err == nil {
		return enRepo
	}
	return "."
}
