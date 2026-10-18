package main

import "fmt"

type Student struct {
	ime     string
	priimek string
	ocene   []int
}

// Mapi se posiljajo po referenci
func dodajOceno(studenti map[string]Student, vpisnaStevilka string, ocena int) {
	// ocen pod 0 ali nad 10 ne sprejemamo
	if ocena < 0 || ocena > 10 {
		fmt.Printf("Neveljavna ocena (%d). Ocene morajo biti na intervalu od 0 do 10!\n", ocena)
		return
	}

	student, ok := studenti[vpisnaStevilka]

	// Ce studenta ni v slovarju, mu ne moremo vpisati ocene
	if !ok {
		fmt.Printf("Študenta z vpisno številko %s ni na seznamu!\n", vpisnaStevilka)
		return
	}

	// Slovar pa je reference
	student.ocene = append(student.ocene, ocena)
	studenti[vpisnaStevilka] = student
}

func povprecje(studenti map[string]Student, vpisnaStevilka string) float64 {
	student, ok := studenti[vpisnaStevilka]

	if !ok {
		return -1.0
	}

	if len(student.ocene) < 6 {
		return 0.0
	}

	sum := 0
	for _, ocena := range student.ocene {
		sum += ocena
	}
	avg := float64(sum) / float64(len(student.ocene))

	return avg
}

func izpisRedovalnice(studenti map[string]Student) {
	fmt.Println("REDOVALNICA:")
	// Zgoscena tabela kljucev nam ne zagotavlja vrstnega reda
	for vpisna, student := range studenti {
		ime := student.ime
		priimek := student.priimek
		ocene := student.ocene
		fmt.Printf("%s - %s %s: %v\n", vpisna, ime, priimek, ocene)
	}
}

func izpisiKoncniUspeh(studenti map[string]Student) {
	for vpisna, student := range studenti {
		ime := student.ime
		priimek := student.priimek
		povprecnaOcena := povprecje(studenti, vpisna)

		var status string

		switch {
		case povprecnaOcena >= 9.0:
			status = "Odličen student!"
		case povprecnaOcena >= 6.0:
			status = "Povprečen študent"
		default:
			status = "Neuspešen študent"
		}

		fmt.Printf("%s %s: povprečna ocena %.1f -> %s\n", ime, priimek, povprecnaOcena, status)
	}
}

func main() {
	// Studenti
	ana := Student{
		ime:     "Ana",
		priimek: "Novak",
		ocene:   []int{8, 8, 9, 8},
	}
	boris := Student{
		ime:     "Boris",
		priimek: "Kralj",
		ocene:   []int{10, 10, 9, 10, 10, 8},
	}
	janez := Student{
		ime:     "Janez",
		priimek: "Novak",
		ocene:   []int{6, 7, 8, 5, 5, 6, 8, 6},
	}

	// Slovar studentov
	studenti := map[string]Student{
		"1000": ana,
		"1001": boris,
		"1002": janez,
	}

	fmt.Println("Dodajanje ocen:")
	// Ta student ne obstaja
	dodajOceno(studenti, "9999", 5)
	// Te ocene ne moremo dodeliti
	dodajOceno(studenti, "1000", 40)
	// Ana je pri OUI dobila 10
	dodajOceno(studenti, "1000", 10)
	fmt.Println(studenti)

	fmt.Println()

	fmt.Println("Povprečje ocen:")
	fmt.Println(povprecje(studenti, "1000")) // 0.0, Ana ima le 5 ocen
	fmt.Println(povprecje(studenti, "1005")) // -1.0, student s to vpisno stevilko ne obstaja
	fmt.Println(povprecje(studenti, "1002")) // 6.375, Za Janeza znamo izracunati povprecje

	fmt.Println()

	fmt.Println("Izpis redovalnice:")
	izpisRedovalnice(studenti)

	fmt.Println()

	fmt.Println("Izpis končnih uspehov:")
	izpisiKoncniUspeh(studenti)
}
