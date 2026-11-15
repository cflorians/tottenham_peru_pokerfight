package main

import (
	"fmt"
	"io"
	"strings"
)

// imprimirLiga escribe el informe markdown de una liga.
func imprimirLiga(w io.Writer, cfg configLiga, inf *informeLiga) {
	fmt.Fprintf(w, "## Liga %d (semilla %d)\n\n", inf.Numero, inf.Semilla)
	fmt.Fprintln(w, "| Ronda | Mesas | Con fichas antes | Con fichas después | Fichas de Lince | Timeouts Lince | Suma de fichas | Duración |")
	fmt.Fprintln(w, "| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: |")
	total := int64(cfg.N) * stackInicial
	for _, r := range inf.Rondas {
		lince := fmt.Sprint(r.FichasLince)
		if r.FichasLince == 0 {
			if inf.RondaQuiebraLi == r.Numero {
				lince = "**0 (cae aquí)**"
			} else {
				lince = "—"
			}
		}
		suma := "ok"
		if r.SumaFichas != total {
			suma = fmt.Sprintf("**%d ≠ %d**", r.SumaFichas, total)
		}
		var notas []string
		for _, m := range r.Mesas {
			if m.CerradaPorReloj {
				notas = append(notas, m.ID+" por reloj")
			}
		}
		dur := fmt.Sprintf("%.0f s", r.DuracionSeg)
		if len(notas) > 0 {
			dur += " (" + strings.Join(notas, ", ") + ")"
		}
		fmt.Fprintf(w, "| %d | %d | %d | %d | %s | %d | %s | %s |\n", r.Numero, len(r.Mesas), r.VivosAntes, r.VivosTras, lince, r.TimeoutsLin, suma, dur)
	}
	fmt.Fprintln(w)

	if inf.RondaQuiebraLi > 0 {
		fmt.Fprintf(w, "Lince cayó en la ronda %d. ", inf.RondaQuiebraLi)
	}
	fmt.Fprintf(w, "Lince terminó %d.º de %d en la clasificatoria.\n\n", inf.PuestoLince, cfg.N)

	fmt.Fprintln(w, "Clasificados a la final:")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "| Puesto | Jugador | Fichas o caída |")
	fmt.Fprintln(w, "| ---: | --- | --- |")
	for _, p := range inf.Clasificacion[:len(inf.Finalistas)] {
		detalle := fmt.Sprint(p.Fichas)
		if p.Fichas == 0 {
			detalle = fmt.Sprintf("eliminado en la ronda %d", p.RondaQuiebra)
		}
		fmt.Fprintf(w, "| %d | %s | %s |\n", p.Puesto, nombre(p.ID), detalle)
	}
	fmt.Fprintln(w)

	switch {
	case inf.Final == nil:
		fmt.Fprintln(w, "Sin mesa final.")
	case inf.GanadorFinal == "":
		fmt.Fprintf(w, "La mesa final no produjo resultado: %s\n", inf.Final.Error)
	default:
		fmt.Fprintf(w, "**Final**: ganó %s en %d manos (%.0f s). Orden: %s.\n", nombre(inf.GanadorFinal), inf.Final.Manos, inf.Final.DuracionSeg, strings.Join(nombres(inf.Final.Posiciones), ", "))
	}
	fmt.Fprintf(w, "\nTimeouts de Lince en toda la liga: %d. Duración de la liga: %.0f s.\n\n", inf.TimeoutsLince, inf.DuracionSeg)

	ch := inf.Chequeo
	fmt.Fprintln(w, "**Chequeo del estado de Lince** (estado.json y log):")
	fmt.Fprintln(w)
	if ch.EstadoLeido {
		fmt.Fprintf(w, "- Cree estar en la ronda %d, fase %s (jugador %s). De verdad jugó %d mesas y la última fue de la %s.\n", ch.Ronda, ch.Fase, ch.IDJugador, ch.RondaEsperada, ch.FaseEsperada)
		fmt.Fprintf(w, "- Fases por ronda según Lince: %s.\n", strings.Join(ch.FasesPorRonda, ", "))
	}
	if ch.Coincide {
		fmt.Fprintln(w, "- Coincide: rondas, fases, fichas al empezar cada ronda y jugadores por mesa.")
	} else {
		for _, d := range ch.Diferencias {
			fmt.Fprintf(w, "- NO coincide: %s\n", d)
		}
	}
	fmt.Fprintf(w, "- Log: %d líneas con PÁNICO y %d con ERROR.\n", ch.Panicos, ch.Errores)
	for _, l := range ch.EjemplosLog {
		fmt.Fprintf(w, "  - `%s`\n", recortar(l, 200))
	}
	if len(inf.Problemas) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "**Problemas del ensayo:**")
		fmt.Fprintln(w)
		for _, p := range inf.Problemas {
			fmt.Fprintf(w, "- %s\n", p)
		}
	}
	fmt.Fprintln(w)
}

// imprimirAgregado resume varias ligas.
func imprimirAgregado(w io.Writer, cfg configLiga, informes []*informeLiga) {
	k := len(informes)
	final, primero, gano, coincide := 0, 0, 0, 0
	quiebras := make([]int, cfg.Rondas+1)
	var timeouts uint64
	var duracion float64
	for _, inf := range informes {
		if inf.LinceEnFinal {
			final++
		}
		if inf.PuestoLince == 1 {
			primero++
		}
		if inf.LinceGanoFinal {
			gano++
		}
		if inf.Chequeo.Coincide {
			coincide++
		}
		if inf.RondaQuiebraLi > 0 && inf.RondaQuiebraLi <= cfg.Rondas {
			quiebras[inf.RondaQuiebraLi]++
		}
		timeouts += inf.TimeoutsLince
		duracion += inf.DuracionSeg
	}
	fmt.Fprintf(w, "## Resumen de %d ligas\n\n", k)
	fmt.Fprintln(w, "| Métrica | Ligas | % |")
	fmt.Fprintln(w, "| --- | ---: | ---: |")
	fila := func(nombre string, v int) {
		fmt.Fprintf(w, "| %s | %d de %d | %.0f %% |\n", nombre, v, k, 100*float64(v)/float64(k))
	}
	fila("Lince llegó a la final", final)
	fila("Lince 1.º de la clasificatoria", primero)
	fila("Lince ganó la final", gano)
	fila("Estado de Lince coincide con la realidad", coincide)
	fmt.Fprintln(w)
	fmt.Fprint(w, "Quiebras de Lince por ronda:")
	alguna := false
	for r := 1; r <= cfg.Rondas; r++ {
		if quiebras[r] > 0 {
			fmt.Fprintf(w, " R%d: %d;", r, quiebras[r])
			alguna = true
		}
	}
	if !alguna {
		fmt.Fprint(w, " ninguna.")
	}
	fmt.Fprintf(w, "\n\nTimeouts de Lince: %d en total. Duración media de una liga: %.0f s.\n\n", timeouts, duracion/float64(k))
}

func nombre(id string) string {
	if id == idLince {
		return "**Lince**"
	}
	return id
}

func nombres(ids []string) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = nombre(id)
	}
	return out
}

func recortar(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
