package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Meritev struct {
	vrsta    string
	vrednost float32
}

// Generiramo nakljucna stevila med -1 in 1
func spremljajTemperaturo(meritve chan<- Meritev, konecTipka <-chan struct{}) {
	var zacetnaMeritev float32 = 20.0

	for {
		select {
		case <-konecTipka:
			return
		default:
			novaMeritev := zacetnaMeritev + (2*rand.Float32()-1)*0.1
			meritve <- Meritev{vrsta: "temperatura", vrednost: novaMeritev}
			zacetnaMeritev = novaMeritev
			time.Sleep(100 * time.Millisecond)
		}
	}
}
func spremljajTlak(meritve chan<- Meritev, konecTipka <-chan struct{}) {
	var zacetnaMeritev float32 = 1.013

	for {
		select {
		case <-konecTipka:
			return
		default:
			novaMeritev := zacetnaMeritev + (2*rand.Float32()-1)*0.0001
			meritve <- Meritev{vrsta: "tlak", vrednost: novaMeritev}
			zacetnaMeritev = novaMeritev
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func spremljajVlaznost(meritve chan<- Meritev, konecTipka <-chan struct{}) {
	var zacetnaMeritev float32 = 50.0

	for {
		select {
		case <-konecTipka:
			return
		default:
			novaMeritev := zacetnaMeritev + (2*rand.Float32()-1)*0.001
			meritve <- Meritev{vrsta: "vlažnost", vrednost: novaMeritev}
			zacetnaMeritev = novaMeritev
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func spremljajTipko(konecTipka chan<- struct{}) {
	fmt.Scanln()
	close(konecTipka)
}

func main() {
	// Definiraj kanal za meritve
	meritve := make(chan Meritev, 3)
	konecTipka := make(chan struct{})

	// Ne zapiras, ker noces imeti send after closed
	go spremljajTipko(konecTipka)
	go spremljajTemperaturo(meritve, konecTipka)
	go spremljajTlak(meritve, konecTipka)
	go spremljajVlaznost(meritve, konecTipka)

	// Waitgroupa tukaj ne rabimo, ker itak ne rabimo cakati na rezultat.
	for {
		select {
		case meritev := <-meritve:
			var koncnica string
			switch meritev.vrsta {
			case "temperatura":
				koncnica = "C"
			case "vlažnost":
				koncnica = "%"
			case "tlak":
				koncnica = "bar"
			}
			fmt.Printf("%s: %.4f %s\n", meritev.vrsta, meritev.vrednost, koncnica)
		case <-time.After(5 * time.Second):
			fmt.Println("Sistem se ne odziva! Minilo je 5s od zadnjega sporočila!")
			return
		case <-konecTipka:
			return
		}
	}
}
