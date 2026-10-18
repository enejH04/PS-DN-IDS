package main

import (
	"fmt"
	"sync"
)

type narocilo interface {
	obdelaj()
}

type izdelek struct {
	imeIzdelka string
	cena       float64
	teza       float64
}

type eknjiga struct {
	naslovKnjige string
	cena         float64
}

type spletniTecaj struct {
	imeTecaja   string
	trajanjeUre int
	cenaUre     float64
}

// Merjenje skupne cene izdelkov
var promet float64 = 0.0
var stNarocil int = 0

// Ker sta promet in stNarocil deljeni vir, ga bomo morali zaklepati
var varLock sync.Mutex

// Da zagotovimo, da se outputi ne pomesajo, Printf zaklenemo. 2 kljucavnici
// zmanjsata nepotrebno cakanje
var writerLock sync.Mutex

// Gorutine mora glavna gorutina na koncu seveda se pocakati
var wg sync.WaitGroup

// Opomba - stvari bodo konsistentne kar se tice vrednost, ampak je izpis
// lahko out of order (npr. najprej se izpisejo podatki o narocilu 3,
// potem o narocilu 1 ...). Problem je, da Gojev Writer ne zaklepa pisanja. Ni
// garancije, da ne bo prislo do pomesanja outputov (OS stvari so tukaj vazne).

// Each call allocates a buffer, fills it, and does a single call to the
// writer when it's done.  There is no shared state.
// The writer might not be safe, but it usually is.  I think what's happening is
// that the operating system is occasionally interleaving pieces when some
// internal buffer fills up and a scheduling event occurs.
func (i izdelek) obdelaj() {
	defer wg.Done()

	varLock.Lock()
	promet += i.cena
	stNarocil++
	narociloId := stNarocil
	varLock.Unlock()

	writerLock.Lock()
	fmt.Printf("\nŠtevilka naročila: %d\nIme izdelka: %s\nCena: %.2f €\nTeža: %.2f kg\n",
		narociloId,
		i.imeIzdelka,
		i.cena,
		i.teza)
	writerLock.Unlock()
}
func (k eknjiga) obdelaj() {
	defer wg.Done()

	varLock.Lock()
	promet += k.cena
	stNarocil++
	narociloId := stNarocil
	varLock.Unlock()

	writerLock.Lock()
	fmt.Printf("\nŠtevilka naročila: %d\nNaslov knjige: %s\nCena: %.2f €\n",
		narociloId,
		k.naslovKnjige,
		k.cena,
	)
	writerLock.Unlock()
}

func (s spletniTecaj) obdelaj() {
	defer wg.Done()

	cena := s.cenaUre * float64(s.trajanjeUre)
	varLock.Lock()
	promet += cena
	stNarocil++
	narociloId := stNarocil
	varLock.Unlock()

	writerLock.Lock()
	fmt.Printf("\nŠtevilka naročila: %d\nIme tečaja: %s\nTrajanje ure: %d\nCena ure: %.2f €\nSkupna cena: %.2f €\n",
		narociloId,
		s.imeTecaja,
		s.trajanjeUre,
		s.cenaUre,
		cena,
	)
	writerLock.Unlock()
}

func main() {
	i1 := spletniTecaj{"Go 101", 3, 10.0}
	i2 := eknjiga{"An Introduction to Statistical Learning", 30.0}
	i3 := spletniTecaj{"Statistical Rethinking", 10, 35.0}
	i4 := izdelek{"Lenovo Thinkpad X1 Carbon G13", 2000.0, 1.15}
	i5 := eknjiga{"Crime and Punishment", 15}
	i6 := izdelek{"Dell Ultrasharp", 450, 6.4}
	i7 := spletniTecaj{"React 101", 10, 20.0}
	i8 := eknjiga{"Algorithms in C++", 60}
	i9 := izdelek{"Sladoled", 1.5, 0.2}
	i10 := eknjiga{"The Bootstrap and Edgeworth Expansion", 60}
	i11 := izdelek{"Burek", 1.20, 0.2}

	nakupi := []narocilo{i1, i2, i3, i4, i5, i6, i7, i8, i9, i10, i11}

	wg.Add(len(nakupi))

	for _, n := range nakupi {
		go n.obdelaj()
	}

	wg.Wait()

	fmt.Printf("\nObdelal sem %d od %d naročil. Skupni znesek je %.2f €!\n",
		stNarocil,
		len(nakupi),
		promet,
	)
}
