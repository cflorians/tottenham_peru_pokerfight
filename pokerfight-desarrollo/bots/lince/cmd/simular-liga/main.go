// simular-liga juega miles de torneos completos (clasificatoria de 10 rondas y
// mesa final) en el simulador rápido y reporta las métricas del plan:
//
//	M1   probabilidad de llegar a la final
//	M1b  probabilidad de terminar 1.º en la clasificatoria
//	M2   probabilidad de ganar la final estando en ella
//
// El "héroe" es el bot que se mide: se sienta en una liga de N participantes
// cuyos otros N-1 salen de un campo (una hipótesis sobre cómo juegan los demás
// equipos). Todos los héroes juegan las mismas ligas: mismos sorteos y cartas.
//
//	go run ./bots/lince/cmd/simular-liga                        # línea base: 4 campos, N=20
//	go run ./bots/lince/cmd/simular-liga -n 15,20,30,40 -tabla  # todos los tamaños y tablas de valor
//	go run ./bots/lince/cmd/simular-liga -heroes v0,otro.json   # comparar un genoma nuevo con la v0
//
// La salida es markdown: se puede redirigir a un archivo de Simulaciones/.
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/entrenamiento"
	"github.com/aicspucp/pokerfight-arena-participantes/bots/lince/estrategia"
)

// Los tramos de stack de las tablas de valor, como fracción del stack medio de
// los que siguen vivos. Un tramo va desde su límite hasta el siguiente.
var tramos = []float64{0, 0.25, 0.5, 0.75, 1, 1.5, 2, 3}

func nombreTramo(i int) string {
	if i == len(tramos)-1 {
		return fmt.Sprintf("≥%g", tramos[i])
	}
	return fmt.Sprintf("%g-%g", tramos[i], tramos[i+1])
}

func tramoDe(relativo float64) int {
	for i := len(tramos) - 1; i > 0; i-- {
		if relativo >= tramos[i] {
			return i
		}
	}
	return 0
}

// politicaFinal, si no es nil, es con qué juegan los héroes la mesa final.
var politicaFinal estrategia.Politica

type heroe struct {
	nombre   string
	politica estrategia.Politica
}

// cuenta es lo que se acumula de un héroe en un campo y un tamaño de liga.
type cuenta struct {
	ligas        int
	final        int   // Ligas en que llegó a la final.
	primero      int   // Ligas en que terminó 1.º.
	fichas       int64 // Suma de las fichas al terminar.
	empezoRonda  []int // Ligas en que empezó la ronda r con fichas.
	quebroRonda  []int // Ligas en que quebró en la ronda r.
	finalGanadas []int // Por tope de manos.
	// Qué pasó en cada liga, para comparar dos héroes liga por liga (todos
	// juegan las mismas). Solo se llenan en la cuenta total.
	finalEnLiga   []bool
	primeroEnLiga []bool

	// Datos del campo entero (se llenan solo con el primer héroe).
	vivosTrasRonda []int64
	rondaDeSeis    []int // rondaDeSeis[r]: ligas en que quedaron 6 o menos tras la ronda r (0 = nunca).
	fichasPrimero  int64
	fichasSexto    int64
	conSextoVivo   int
	finalistas     map[string]int
	valor          [][]celda // [ronda][tramo]
}

type celda struct{ casos, final, primero int }

func nuevaCuenta(rondas, topes int) *cuenta {
	c := &cuenta{
		empezoRonda:    make([]int, rondas),
		quebroRonda:    make([]int, rondas),
		finalGanadas:   make([]int, topes),
		vivosTrasRonda: make([]int64, rondas),
		rondaDeSeis:    make([]int, rondas+1),
		finalistas:     map[string]int{},
		valor:          make([][]celda, rondas),
	}
	for r := range c.valor {
		c.valor[r] = make([]celda, len(tramos))
	}
	return c
}

func (c *cuenta) sumar(o *cuenta) {
	c.ligas += o.ligas
	c.final += o.final
	c.primero += o.primero
	c.fichas += o.fichas
	c.fichasPrimero += o.fichasPrimero
	c.fichasSexto += o.fichasSexto
	c.conSextoVivo += o.conSextoVivo
	for i := range c.empezoRonda {
		c.empezoRonda[i] += o.empezoRonda[i]
		c.quebroRonda[i] += o.quebroRonda[i]
		c.vivosTrasRonda[i] += o.vivosTrasRonda[i]
	}
	for i := range c.rondaDeSeis {
		c.rondaDeSeis[i] += o.rondaDeSeis[i]
	}
	for i := range c.finalGanadas {
		c.finalGanadas[i] += o.finalGanadas[i]
	}
	for k, v := range o.finalistas {
		c.finalistas[k] += v
	}
	for r := range c.valor {
		for t := range c.valor[r] {
			c.valor[r][t].casos += o.valor[r][t].casos
			c.valor[r][t].final += o.valor[r][t].final
			c.valor[r][t].primero += o.valor[r][t].primero
		}
	}
}

func main() {
	log.SetFlags(0)
	var (
		listaN      = flag.String("n", "20", "participantes por liga, separados por coma (por ejemplo 15,20,30,40)")
		listaCampos = flag.String("campos", "kit,agresivo,tight,mixto,fuerte", "campos de rivales a simular (fuerte solo existe si hay genomas en la carpeta rivales)")
		listaHeroes = flag.String("heroes", "v0,darwin,novato", "bots a medir: v0 (genoma_v0.json), actual (genoma.json), final (genoma_final.json), darwin, novato, o la ruta de un genoma .json")
		ligas       = flag.Int("ligas", 2000, "ligas por héroe, campo y tamaño")
		listaTopes  = flag.String("topes", "150,500,0", "topes de manos de la mesa final (0 = sin reloj)")
		tamBib      = flag.Int("biblioteca", 20000, "repartos precalculados")
		muestras    = flag.Int("muestras", 200, "muestras de Monte Carlo por calle al precalcular la fuerza")
		semilla     = flag.Uint64("semilla", 2026, "semilla: misma semilla, mismo resultado")
		dir         = flag.String("dir", dirDelBot(), "carpeta del bot (genomas y caché)")
		rutaCache   = flag.String("cache", filepath.Join("datos", "biblioteca-liga.cache"), "caché de la biblioteca, relativa a -dir (vacío para no usar)")
		paridad     = flag.Bool("paridad", false, "en vez de ligas, juega mesas sueltas de una ronda con las alineaciones de Simulaciones/pruebas_formato_real.md, para comparar el simulador con la arena real")
		diagnostico = flag.Bool("diagnostico", false, "en vez de ligas, cuenta con qué jugada se queda sin fichas el primer héroe en mesas sueltas de una ronda")
		rutaFinal   = flag.String("genoma-final", "", "genoma con el que los héroes juegan la mesa final (por defecto, el mismo de la liga)")
		tabla       = flag.Bool("tabla", false, "imprimir las tablas de valor P(final) y P(1.º) por ronda y stack")
	)
	flag.Parse()

	darwin := leerGenoma(filepath.Join(*dir, "genoma_darwin.json"))
	fuertes, err := entrenamiento.LeerRivales(filepath.Join(*dir, "rivales"))
	if err != nil {
		log.Fatalf("rivales: %v", err)
	}
	prudente := leerGenoma(filepath.Join(*dir, "genoma_v0.json"))

	var heroes []heroe
	for _, nombre := range strings.Split(*listaHeroes, ",") {
		switch nombre = strings.TrimSpace(nombre); nombre {
		case "v0":
			heroes = append(heroes, heroe{"v0 (prudente)", prudente})
		case "actual":
			heroes = append(heroes, heroe{"actual (genoma.json)", leerGenoma(filepath.Join(*dir, "genoma.json"))})
		case "final":
			heroes = append(heroes, heroe{"final (genoma_final.json)", leerGenoma(filepath.Join(*dir, "genoma_final.json"))})
		case "darwin":
			heroes = append(heroes, heroe{"Darwin original", darwin})
		case "novato":
			heroes = append(heroes, heroe{"Novato", entrenamiento.Novato{}})
		default:
			heroes = append(heroes, heroe{strings.TrimSuffix(filepath.Base(nombre), ".json"), leerGenoma(nombre)})
		}
	}
	if *rutaFinal != "" {
		politicaFinal = leerGenoma(*rutaFinal)
	}
	tamanos := enteros(*listaN)
	topes := enteros(*listaTopes)

	var campos []entrenamiento.Campo
	for _, c := range entrenamiento.Campos(darwin, prudente, fuertes...) {
		if strings.Contains(","+*listaCampos+",", ","+c.Nombre+",") {
			campos = append(campos, c)
		}
	}
	if len(campos) == 0 || len(heroes) == 0 || len(tamanos) == 0 || len(topes) == 0 {
		log.Fatal("faltan campos, héroes, tamaños o topes")
	}

	inicio := time.Now()
	cfgBib := entrenamiento.ConfigBiblioteca{Repartos: *tamBib, Muestras: *muestras, Semilla: *semilla}
	cache := *rutaCache
	if cache != "" {
		cache = filepath.Join(*dir, cache)
		_ = os.MkdirAll(filepath.Dir(cache), 0o755)
	}
	ultimo := time.Now()
	bib, _, err := entrenamiento.CargarOGenerar(cache, cfgBib, func(hechos, total int) {
		if time.Since(ultimo) > 2*time.Second || hechos == total {
			fmt.Fprintf(os.Stderr, "\r  precalculando repartos: %d/%d", hechos, total)
			ultimo = time.Now()
		}
	})
	if err != nil {
		log.Printf("aviso: %v", err)
	}
	fmt.Fprintf(os.Stderr, "\nbiblioteca lista: %d repartos (%s)\n", len(bib), time.Since(inicio).Round(time.Second))

	cfgLiga := entrenamiento.LigaDeLasBases()
	if *diagnostico {
		informarDiagnostico(cfgLiga, heroes[0], campos, bib, *ligas, *semilla)
		return
	}
	if *paridad {
		informarParidad(cfgLiga, darwin, prudente, bib, *ligas, *semilla)
		return
	}

	fmt.Printf("Simulador rápido · %d ligas por héroe, campo y tamaño · semilla %d · biblioteca de %d repartos con %d muestras\n\n",
		*ligas, *semilla, len(bib), *muestras)
	fmt.Printf("Liga: %d rondas de %d manos, %d fichas, ciegas %d/%d, pasan %d. Final: 6 jugadores, %d fichas, topes de manos %s (0 = %d).\n\n",
		cfgLiga.Rondas, cfgLiga.ManosPorRonda, cfgLiga.Mesa.Stack, cfgLiga.Mesa.CiegaChica, cfgLiga.Mesa.CiegaGrande, cfgLiga.Clasifican,
		cfgLiga.Mesa.Stack, *listaTopes, entrenamiento.TopeSinReloj)
	fmt.Println("Entre corchetes, el intervalo de confianza del 95 % (Wilson). U = 60 × M1 × M2 + 30 × M1b, con M2 promediada entre los topes.")

	for _, campo := range campos {
		for _, n := range tamanos {
			rivales := campo.Armar(n - 1)
			cuentas := make([]*cuenta, len(heroes))
			for h, he := range heroes {
				cuentas[h] = simular(cfgLiga, he, rivales, bib, *ligas, topes, *semilla, h == 0)
				fmt.Fprintf(os.Stderr, "  %s N=%d %s (%s)\n", campo.Nombre, n, he.nombre, time.Since(inicio).Round(time.Second))
			}
			informar(campo, n, rivales, heroes, cuentas, cfgLiga, topes, *tabla)
		}
	}
	fmt.Fprintf(os.Stderr, "total: %s\n", time.Since(inicio).Round(time.Second))
}

// simular juega `ligas` ligas con el héroe en el índice 0 y los rivales
// detrás. La liga k usa la misma semilla y el mismo tramo de la biblioteca sea
// cual sea el héroe.
func simular(cfg entrenamiento.ConfigLiga, he heroe, rivales []entrenamiento.TipoBot, bib entrenamiento.Biblioteca, ligas int, topes []int, semilla uint64, conCampo bool) *cuenta {
	politicas := make([]estrategia.Politica, 0, len(rivales)+1)
	nombres := make([]string, 0, len(rivales)+1)
	politicas, nombres = append(politicas, he.politica), append(nombres, "héroe")
	for _, r := range rivales {
		politicas, nombres = append(politicas, r.Politica), append(nombres, r.Nombre)
	}
	n := len(politicas)

	finalEnLiga, primeroEnLiga := make([]bool, ligas), make([]bool, ligas)
	hilos := runtime.NumCPU()
	parciales := make([]*cuenta, hilos)
	var espera sync.WaitGroup
	for w := 0; w < hilos; w++ {
		espera.Add(1)
		go func() {
			defer espera.Done()
			c := nuevaCuenta(cfg.Rondas, len(topes))
			parciales[w] = c
			finalistas := make([]estrategia.Politica, 0, cfg.Clasifican)
			for k := w; k < ligas; k += hilos {
				s := semilla*1_000_003 + uint64(k)
				desde := int(s * 2654435761 % uint64(len(bib)))
				res := entrenamiento.JugarLiga(cfg, politicas, bib, desde, s)

				c.ligas++
				c.fichas += res.Fichas[0]
				for r := 0; r < cfg.Rondas; r++ {
					if res.FichasAlEmpezar[r][0] > 0 {
						c.empezoRonda[r]++
					}
				}
				if q := res.RondaQuiebra[0]; q > 0 {
					c.quebroRonda[q-1]++
				}
				if res.Primero() == 0 {
					c.primero++
					primeroEnLiga[k] = true
				}
				if res.EsFinalista(0) {
					c.final++
					finalEnLiga[k] = true
					finalistas = finalistas[:0]
					miSilla := 0
					for i, f := range res.Finalistas {
						finalistas = append(finalistas, politicas[f])
						if f == 0 {
							miSilla = i
							if politicaFinal != nil {
								finalistas[i] = politicaFinal
							}
						}
					}
					for t, tope := range topes {
						ganador, _ := entrenamiento.JugarFinal(entrenamiento.FinalDeLasBases(tope), finalistas, bib, desde+7919*(t+1), s+uint64(t)+1)
						if ganador == miSilla {
							c.finalGanadas[t]++
						}
					}
				}
				if conCampo {
					anotarCampo(c, cfg, res, nombres, n)
				}
			}
		}()
	}
	espera.Wait()

	total := nuevaCuenta(cfg.Rondas, len(topes))
	for _, c := range parciales {
		total.sumar(c)
	}
	total.finalEnLiga, total.primeroEnLiga = finalEnLiga, primeroEnLiga
	return total
}

func anotarCampo(c *cuenta, cfg entrenamiento.ConfigLiga, res entrenamiento.ResultadoLiga, nombres []string, n int) {
	deSeis := 0
	for r, vivos := range res.VivosTrasRonda {
		c.vivosTrasRonda[r] += int64(vivos)
		if deSeis == 0 && vivos <= cfg.Clasifican {
			deSeis = r + 1
		}
	}
	c.rondaDeSeis[deSeis]++
	c.fichasPrimero += res.Fichas[res.Primero()]
	if len(res.Finalistas) == cfg.Clasifican {
		if f := res.Fichas[res.Finalistas[cfg.Clasifican-1]]; f > 0 {
			c.fichasSexto += f
			c.conSextoVivo++
		}
	}
	for _, f := range res.Finalistas {
		c.finalistas[nombres[f]]++
	}
	// Tablas de valor: todos los que empiezan cada ronda con fichas.
	totalFichas := float64(int64(n) * cfg.Mesa.Stack)
	for r := 0; r < cfg.Rondas; r++ {
		vivos := 0
		for _, f := range res.FichasAlEmpezar[r] {
			if f > 0 {
				vivos++
			}
		}
		if vivos == 0 {
			continue
		}
		medio := totalFichas / float64(vivos)
		for i, f := range res.FichasAlEmpezar[r] {
			if f <= 0 {
				continue
			}
			cel := &c.valor[r][tramoDe(float64(f)/medio)]
			cel.casos++
			if res.EsFinalista(i) {
				cel.final++
			}
			if res.Primero() == i {
				cel.primero++
			}
		}
	}
}

func informar(campo entrenamiento.Campo, n int, rivales []entrenamiento.TipoBot, heroes []heroe, cuentas []*cuenta, cfg entrenamiento.ConfigLiga, topes []int, conTabla bool) {
	composicion := map[string]int{}
	var orden []string
	for _, r := range rivales {
		if composicion[r.Nombre] == 0 {
			orden = append(orden, r.Nombre)
		}
		composicion[r.Nombre]++
	}
	var partes []string
	for _, nombre := range orden {
		partes = append(partes, fmt.Sprintf("%d %s", composicion[nombre], nombre))
	}
	fmt.Printf("\n## Campo %s · N = %d\n\nRivales: %s.\n\n", campo.Nombre, n, strings.Join(partes, ", "))

	// --- Métricas de cada héroe ---
	fmt.Print("| Bot | M1 llega a la final | M1b termina 1.º |")
	for _, t := range topes {
		fmt.Printf(" M2 tope %s |", nombreTope(t))
	}
	fmt.Println(" U | Fichas medias |")
	fmt.Print("| --- | ---: | ---: |")
	for range topes {
		fmt.Print(" ---: |")
	}
	fmt.Println(" ---: | ---: |")
	for h, he := range heroes {
		c := cuentas[h]
		m1 := float64(c.final) / float64(c.ligas)
		m1b := float64(c.primero) / float64(c.ligas)
		fmt.Printf("| %s | %s | %s |", he.nombre, porcentaje(c.final, c.ligas), porcentaje(c.primero, c.ligas))
		var m2 float64
		for t := range topes {
			fmt.Printf(" %s |", porcentaje(c.finalGanadas[t], c.final))
			if c.final > 0 {
				m2 += float64(c.finalGanadas[t]) / float64(c.final) / float64(len(topes))
			}
		}
		fmt.Printf(" %.2f | %d |\n", 60*m1*m2+30*m1b, c.fichas/int64(c.ligas))
	}

	// --- Comparación pareada contra el primer héroe ---
	if len(heroes) > 1 {
		fmt.Printf("\nDiferencia contra %s, liga por liga (mismas ligas para los dos; ✔ = el intervalo del 95 %% no toca el cero). U_A = 12 × M1 + 30 × M1b es el premio esperado del plan A con M2 fija en 0,20.\n\n", heroes[0].nombre)
		fmt.Println("| Bot | Δ M1 (puntos) | Δ M1b (puntos) | Δ U_A (soles) |\n| --- | ---: | ---: | ---: |")
		base := cuentas[0]
		for h := 1; h < len(heroes); h++ {
			c := cuentas[h]
			fmt.Printf("| %s | %s | %s | %s |\n", heroes[h].nombre,
				diferencia(c, base, 100, 0), diferencia(c, base, 0, 100), diferencia(c, base, 12, 30))
		}
	}

	// --- Quiebra por ronda ---
	fmt.Print("\nQuiebra por ronda (de las ligas en que empezó la ronda con fichas):\n\n| Bot |")
	for r := 1; r <= cfg.Rondas; r++ {
		fmt.Printf(" R%d |", r)
	}
	fmt.Print(" Llega vivo al final |\n| --- |")
	for r := 0; r <= cfg.Rondas; r++ {
		fmt.Print(" ---: |")
	}
	fmt.Println()
	for h, he := range heroes {
		c := cuentas[h]
		fmt.Printf("| %s |", he.nombre)
		quiebras := 0
		for r := 0; r < cfg.Rondas; r++ {
			quiebras += c.quebroRonda[r]
			if c.empezoRonda[r] == 0 {
				fmt.Print(" - |")
				continue
			}
			fmt.Printf(" %.1f %% |", 100*float64(c.quebroRonda[r])/float64(c.empezoRonda[r]))
		}
		fmt.Printf(" %.1f %% |\n", 100*float64(c.ligas-quiebras)/float64(c.ligas))
	}

	// --- El campo ---
	c := cuentas[0]
	fmt.Printf("\nEl campo (ligas con %s como héroe):\n\n| | ", heroes[0].nombre)
	for r := 1; r <= cfg.Rondas; r++ {
		fmt.Printf("R%d | ", r)
	}
	fmt.Print("\n| --- |")
	for r := 0; r < cfg.Rondas; r++ {
		fmt.Print(" ---: |")
	}
	fmt.Print("\n| Vivos al terminar la ronda (media) |")
	for r := 0; r < cfg.Rondas; r++ {
		fmt.Printf(" %.1f |", float64(c.vivosTrasRonda[r])/float64(c.ligas))
	}
	fmt.Print("\n| Ligas en que ya quedan 6 o menos (acumulado) |")
	acumulado := 0
	for r := 1; r <= cfg.Rondas; r++ {
		acumulado += c.rondaDeSeis[r]
		fmt.Printf(" %.0f %% |", 100*float64(acumulado)/float64(c.ligas))
	}
	fmt.Println()
	fmt.Printf("\n- Fichas del 1.º al terminar: **%d** de media.\n", c.fichasPrimero/int64(c.ligas))
	if c.conSextoVivo > 0 {
		fmt.Printf("- Fichas del 6.º (el corte), cuando llegan 6 o más con fichas (%.0f %% de las ligas): **%d** de media.\n",
			100*float64(c.conSextoVivo)/float64(c.ligas), c.fichasSexto/int64(c.conSextoVivo))
	} else {
		fmt.Println("- Nunca llegan 6 con fichas al final: clasifican los últimos eliminados.")
	}
	fmt.Print("- Quién llega a la final (finalistas por liga, de 6): ")
	partes = partes[:0]
	for _, nombre := range append([]string{"héroe"}, orden...) {
		porLiga := float64(c.finalistas[nombre]) / float64(c.ligas)
		if nombre == "héroe" {
			partes = append(partes, fmt.Sprintf("héroe %.2f", porLiga))
			continue
		}
		partes = append(partes, fmt.Sprintf("%s %.2f (%.0f %% de los suyos)", nombre, porLiga, 100*porLiga/float64(composicion[nombre])))
	}
	fmt.Println(strings.Join(partes, ", ") + ".")

	if !conTabla {
		return
	}
	for _, t := range []struct {
		titulo string
		valor  func(celda) int
	}{
		{"P(llegar a la final)", func(c celda) int { return c.final }},
		{"P(terminar 1.º)", func(c celda) int { return c.primero }},
	} {
		fmt.Printf("\nTabla de valor · %s según la ronda que empieza y el stack (como fracción del stack medio de los vivos). \"-\" = menos de 30 casos.\n\n| Ronda |", t.titulo)
		for i := range tramos {
			fmt.Printf(" %s |", nombreTramo(i))
		}
		fmt.Print("\n| --- |")
		for range tramos {
			fmt.Print(" ---: |")
		}
		fmt.Println()
		for r := 0; r < cfg.Rondas; r++ {
			fmt.Printf("| %d |", r+1)
			for i := range tramos {
				cel := c.valor[r][i]
				if cel.casos < 30 {
					fmt.Print(" - |")
					continue
				}
				fmt.Printf(" %.0f %% |", 100*float64(t.valor(cel))/float64(cel.casos))
			}
			fmt.Println()
		}
	}
}

// informarParidad juega una sola ronda (50 manos) con las tres alineaciones
// que se midieron en la arena real y reporta, por bot, el neto medio y el
// porcentaje de mesas en que quiebra. Si las réplicas se parecen a los bots de
// verdad, los números se parecen a los de la arena.
func informarParidad(cfg entrenamiento.ConfigLiga, darwin, prudente estrategia.Genoma, bib entrenamiento.Biblioteca, mesas int, semilla uint64) {
	tipos := map[string]estrategia.Politica{
		"darwin": darwin, "prudente": prudente, "espia": entrenamiento.Espia{}, "estadistico": entrenamiento.Estadistico{},
		"base": entrenamiento.Base{}, "novato": entrenamiento.Novato{}, "conservador": entrenamiento.Conservador{}, "aleatorio": entrenamiento.Aleatorio{},
	}
	alineaciones := [][]string{
		{"darwin", "espia", "estadistico", "base", "novato"},
		{"darwin", "espia", "estadistico", "novato", "conservador"},
		{"darwin", "espia", "estadistico", "base", "aleatorio", "conservador"},
		{"prudente", "espia", "estadistico", "base", "novato"},
		{"prudente", "espia", "estadistico", "novato", "conservador"},
		{"prudente", "espia", "estadistico", "base", "aleatorio", "conservador"},
	}
	type suma struct {
		mesas, quiebras int
		neto            int64
	}
	total := map[string]*suma{}
	var orden []string
	for a, alineacion := range alineaciones {
		politicas := make([]estrategia.Politica, len(alineacion))
		for i, nombre := range alineacion {
			politicas[i] = tipos[nombre]
			if total[nombre] == nil {
				total[nombre] = &suma{}
				orden = append(orden, nombre)
			}
		}
		for k := 0; k < mesas; k++ {
			s := semilla*7919 + uint64(a)*1_000_003 + uint64(k)
			stacks := make([]int64, len(politicas))
			for i := range stacks {
				stacks[i] = cfg.Mesa.Stack
			}
			// Se rota quién se sienta dónde para que la posición no favorezca a nadie.
			rot := k % len(politicas)
			enMesa := append(append([]estrategia.Politica{}, politicas[rot:]...), politicas[:rot]...)
			entrenamiento.JugarMesa(cfg.Mesa, enMesa, stacks, bib, int(s*2654435761%uint64(len(bib))), cfg.ManosPorRonda, s, nil)
			for i, f := range stacks {
				t := total[alineacion[(i+rot)%len(politicas)]]
				t.mesas++
				t.neto += f - cfg.Mesa.Stack
				if f <= 0 {
					t.quiebras++
				}
			}
		}
	}
	fmt.Printf("\nParidad: una ronda de %d manos, %d mesas por alineación (las de Simulaciones/pruebas_formato_real.md, con Darwin original y con el genoma prudente en su lugar).\n\n", cfg.ManosPorRonda, mesas)
	fmt.Println("| Bot (réplica) | Mesas | Neto medio | Quiebra |\n| --- | ---: | ---: | ---: |")
	for _, nombre := range orden {
		t := total[nombre]
		fmt.Printf("| %s | %d | %+d | %s |\n", nombre, t.mesas, t.neto/int64(t.mesas), porcentaje(t.quiebras, t.mesas))
	}
}

func nombreTope(t int) string {
	if t <= 0 {
		return "sin reloj"
	}
	return strconv.Itoa(t)
}

// porcentaje escribe k/n con su intervalo de Wilson del 95 %.
func porcentaje(k, n int) string {
	if n == 0 {
		return "-"
	}
	const z = 1.96
	p := float64(k) / float64(n)
	den := 1 + z*z/float64(n)
	centro := (p + z*z/(2*float64(n))) / den
	radio := z * math.Sqrt(p*(1-p)/float64(n)+z*z/(4*float64(n)*float64(n))) / den
	return fmt.Sprintf("%.1f %% [%.1f-%.1f]", 100*p, 100*max(centro-radio, 0), 100*min(centro+radio, 1))
}

// diferencia escribe la media de (a - b) liga por liga, con su intervalo del
// 95 %, de la cantidad pesoFinal × llegó a la final + pesoPrimero × terminó 1.º.
func diferencia(a, b *cuenta, pesoFinal, pesoPrimero float64) string {
	valor := func(c *cuenta, k int) float64 {
		var v float64
		if c.finalEnLiga[k] {
			v += pesoFinal
		}
		if c.primeroEnLiga[k] {
			v += pesoPrimero
		}
		return v
	}
	n := float64(len(a.finalEnLiga))
	var suma, cuadrados float64
	for k := range a.finalEnLiga {
		d := valor(a, k) - valor(b, k)
		suma += d
		cuadrados += d * d
	}
	media := suma / n
	radio := 1.96 * math.Sqrt(max(cuadrados/n-media*media, 0)/n)
	marca := ""
	if media-radio > 0 || media+radio < 0 {
		marca = " ✔"
	}
	return fmt.Sprintf("%+.2f [%+.2f a %+.2f]%s", media, media-radio, media+radio, marca)
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

// dirDelBot es la carpeta del bot, se corra desde la raíz del módulo o desde
// la carpeta del bot.
func dirDelBot() string {
	enRepo := filepath.Join("bots", "lince")
	if _, err := os.Stat(filepath.Join(enRepo, "genoma.json")); err == nil {
		return enRepo
	}
	return "."
}
