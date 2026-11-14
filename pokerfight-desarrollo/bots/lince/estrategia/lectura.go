package estrategia

// Lectura de rivales: cuán seguido apuesta cada uno.
//
// Una apuesta no dice lo mismo viniendo de un bot que apuesta una vez cada
// diez manos que de uno que apuesta casi siempre: al primero hay que creerle,
// al segundo se lo puede pagar y resubir con manos peores. Un genoma con los
// umbrales fijos no puede hacer las dos cosas (Simulaciones/a3_campo_fuerte.md);
// con este dato, sí.
//
// El dato es simple a propósito: se junta solo, mirando jugar, sin arriesgar
// fichas, y se puede calcular igual en el simulador y en la mesa real.

// AgresionNeutra es la agresión que se le supone a un rival del que no se sabe
// nada: de cada 100 veces que pone fichas, 45 son apuestas o subidas y 55 son
// pagos. Está entre los bots del kit (que van del 20 % al 65 %) y lejos de los
// que apuestan casi siempre (95 %).
const AgresionNeutra = 0.45

// pesoPrevio es cuántas acciones "imaginarias" de un rival neutro se le suman
// a cada perfil: con pocas manos vistas la lectura se queda cerca de cero, y
// hacen falta unas cuantas apuestas seguidas para creer que alguien es agresivo.
const pesoPrevio = 8

// Perfil es lo que se le vio hacer a un rival.
type Perfil struct {
	Agresivas int // Apuestas, subidas y all-in.
	Pasivas   int // Pagos.
}

// Anotar suma una acción al perfil. `agresiva` es bet, raise o all-in;
// `pasiva` es call. Pasar y retirarse no cuentan.
func (p *Perfil) Anotar(agresiva, pasiva bool) {
	switch {
	case agresiva:
		p.Agresivas++
	case pasiva:
		p.Pasivas++
	}
}

// Lectura es cuánto más agresivo que un rival neutro es este: va de -0,45 (solo
// paga) a +0,55 (solo apuesta), y vale 0 si todavía no se lo vio jugar.
func (p Perfil) Lectura() float64 {
	return (float64(p.Agresivas)+pesoPrevio*AgresionNeutra)/(float64(p.Agresivas+p.Pasivas)+pesoPrevio) - AgresionNeutra
}
