package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
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

// Za lazje spremljanje pravilnosti imam namesto stNarocil dve spremenljivki
var stOddanihNarocil int = 0
var stObdelanihNarocil int = 0

// Ker sta promet, stOddanihNarocil in stObdelanihNarocil deljeni viri, jih bomo morali zaklepati
var oddajaLock sync.Mutex
var obdelavaLock sync.Mutex

// Da zagotovimo, da se outputi ne pomesajo, Printf zaklenemo. 2 kljucavnici
// zmanjsata nepotrebno cakanje
var writerLock sync.Mutex

// Definiramo wait groupa za producerje in consumerje
var wgProizvajalci, wgObdelovalci sync.WaitGroup

func (i izdelek) obdelaj() {
	obdelavaLock.Lock()
	promet += i.cena
	stObdelanihNarocil++
	narociloId := stObdelanihNarocil
	obdelavaLock.Unlock()

	// Obdelava traja 500 ms
	time.Sleep(500 * time.Millisecond)
	writerLock.Lock()
	fmt.Printf("\nŠtevilka naročila: %d\nIme izdelka: %s\nCena: %.2f €\nTeža: %.2f kg\n",
		narociloId,
		i.imeIzdelka,
		i.cena,
		i.teza)
	writerLock.Unlock()
}
func (k eknjiga) obdelaj() {
	obdelavaLock.Lock()
	promet += k.cena
	stObdelanihNarocil++
	narociloId := stObdelanihNarocil
	obdelavaLock.Unlock()

	// Obdelava traja 500 ms
	time.Sleep(500 * time.Millisecond)
	writerLock.Lock()
	fmt.Printf("\nŠtevilka naročila: %d\nNaslov knjige: %s\nCena: %.2f €\n",
		narociloId,
		k.naslovKnjige,
		k.cena,
	)
	writerLock.Unlock()
}

func (s spletniTecaj) obdelaj() {
	cena := s.cenaUre * float64(s.trajanjeUre)
	obdelavaLock.Lock()
	promet += cena
	stObdelanihNarocil++
	narociloId := stObdelanihNarocil
	obdelavaLock.Unlock()

	// Obdelava traja 500 ms
	time.Sleep(500 * time.Millisecond)
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

func proizvajaj(izdelki []narocilo, vrstaNarocil chan<- narocilo, pritiskTipke <-chan struct{}) {
	defer wgProizvajalci.Done()
	for {
		select {
		case <-pritiskTipke:
			return
		default:
			izdelekIdx := rand.Intn(len(izdelki))

			oddajaLock.Lock()
			stOddanihNarocil++
			oddajaLock.Unlock()

			vrstaNarocil <- izdelki[izdelekIdx]
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func obdeluj(vrstaNarocil <-chan narocilo) {
	defer wgObdelovalci.Done()
	for {
		// Dokler je se kaj v vrsti oz. ta ni zaprta beremo
		if narocilo, ok := <-vrstaNarocil; ok {
			narocilo.obdelaj()
		} else {
			return
		}
	}
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

	// Izdelki, ki so na voljo - med temi bomo izbirali
	izdelki := []narocilo{i1, i2, i3, i4, i5, i6, i7, i8, i9, i10, i11}

	stProizvajalcev := 5
	stObdelovalcev := 2

	// Kanal v katerega se nalagajo narocila za obdelavo - naj ima kapaciteto 5
	// Recimo, da je prostora za 10 narocil
	narociloCap := 10
	vrstaNarocil := make(chan narocilo, narociloCap)

	// Kanal, ki skrbi za zapiranje, ce pride do pritiska tipke - izkljucno za
	// sinhronizacijo
	pritiskTipke := make(chan struct{})

	wgProizvajalci.Add(stProizvajalcev)
	wgObdelovalci.Add(stObdelovalcev)

	for i := 0; i < stProizvajalcev; i++ {
		go proizvajaj(izdelki, vrstaNarocil, pritiskTipke)
	}

	for i := 0; i < stObdelovalcev; i++ {
		go obdeluj(vrstaNarocil)
	}

	// Ko uporabnik pritisne enter, se konca sprejemanje novih narocil
	go func() {
		// Cakaj na pritisk tipke
		fmt.Scanln()
		close(pritiskTipke)
	}()

	// Pocakamo proizvajalce
	wgProizvajalci.Wait()

	// Ko so proizvajalci koncali lahko zapremo vrsto narocil - teh ni vec
	close(vrstaNarocil)

	wgObdelovalci.Wait()

	fmt.Printf("Oddanih je bilo %d naročil. Od tega sem jih obdelal %d. Skupna vrednost je bila %.2f €.\n", stOddanihNarocil, stObdelanihNarocil, promet)
}
