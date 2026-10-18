package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"time"
)

// Identifikator procesa
var id int

// Stevilo vseh procesov
var N int

// Identifikator glavnega procesa - sporocilo bo razsiril med ostale
var root int

func checkError(err error) {
	// Pomozna funkcija za error handling
	if err != nil {
		panic(err)
	}
}

func runRoot(addr *net.UDPAddr, clientAddrs []*net.UDPAddr) {
	// Kolikokrat bomo poskusili ponovno poslati
	max_resend := 5

	// Odpremo povezavo
	conn, err := net.ListenUDP("udp", addr)
	checkError(err)
	defer conn.Close()

	// To check if we got messages from all
	receivedAcks := make(map[int]bool)

	for _, clientAddr := range clientAddrs {
		receivedAcks[clientAddr.Port] = false
	}

	// Definiraj buffer za ack - pregled, ce je odgovoril pravilen proces
	ackBuf := make([]byte, 1024)

	for _, clientAddr := range clientAddrs {
		clientConn, err := net.DialUDP("udp", nil, clientAddr)
		checkError(err)

		// Tvori sporocilo
		msg := fmt.Sprintf("Sporočilo za odjemalca na portu %d", clientAddr.Port)
		byteEncoded := []byte(msg)

		for i := 0; i < max_resend; i++ {
			// Posli sporocilo
			clientConn.Write(byteEncoded)
			// Pocakaj na odgovor
			conn.SetDeadline(time.Now().Add(1 * time.Second))
			nAck, err := conn.Read(ackBuf)

			if err == nil {
				// Dobili smo ack - ampak a smo ga dobili od tega procesa?
				senderPort, _ := strconv.Atoi(string(ackBuf[:nAck]))
				receivedAcks[senderPort] = true

				if senderPort == clientAddr.Port {
					// Iz zanke skocimo, le ce smo dobili odgovor od trenutnega odjemalca - sicer je bil nek drugi
					break
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
		clientConn.Close()
	}

	fmt.Println()
	// Check if all were successful
	for addr, ack := range receivedAcks {
		if !ack {
			fmt.Printf("Odgovora nisem prejel od procesa na portu %d\n", addr)
		}
	}
}

func runClient(localAddr *net.UDPAddr, rootAddr *net.UDPAddr) {
	// Poslusamo za sporocila glavnega procesa
	conn, err := net.ListenUDP("udp", localAddr)
	checkError(err)
	defer conn.Close()

	// Povezava za potrditev
	ackConn, err := net.DialUDP("udp", nil, rootAddr)
	checkError(err)
	defer ackConn.Close()

	buffer := make([]byte, 1024)

	// Tukaj blokiramo, dokler ne dobimo sporocila oz. dokler ne potece deadline (1s)
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	mLen, err := conn.Read(buffer)
	if err != nil {
		return
	}

	// Prejeli smo sporocilo
	message := string(buffer[:mLen])
	// Vrnimo potrditev
	ack := []byte(fmt.Sprint(localAddr.Port))
	ackConn.Write(ack)

	fmt.Printf("Odjemalec na portu %d je od glavnega procesa %d prejel sporočilo: %s\n", localAddr.Port, root, message)
}

func main() {
	// Preberi argumente ukazne vrstice
	idPtr := flag.Int("id", 0, "Process id")
	NPtr := flag.Int("n", 2, "Total number of processes")
	rootIdPtr := flag.Int("rid", 0, "Id of the root process")
	// Neobvezen a uporaben argument
	basePortPtr := flag.Int("p", 9150, "Base port number")

	flag.Parse()

	id = *idPtr
	N = *NPtr
	root = *rootIdPtr
	basePort := *basePortPtr

	if root >= N {
		fmt.Println("Identifikator glavnega procesa ne sme biti vecji ali enak stevilu vseh procesov.")
		return
	}

	if id == root {
		// Pozeni glavni proces
		port := basePort + root
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", port))
		checkError(err)

		// Pridobi naslove odjemalcev
		clientAddrs := make([]*net.UDPAddr, 0, N-1)
		for pId := 0; pId < N; pId++ {
			if pId != root {
				// Ne dodaj glavnega procesa!
				port = basePort + pId
				addrClient, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", port))
				checkError(err)
				clientAddrs = append(clientAddrs, addrClient)
			}
		}

		runRoot(addr, clientAddrs)
	} else {
		// Pozeni odjemalca
		portLocal := basePort + id
		portRoot := basePort + root

		localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", portLocal))
		checkError(err)
		rootAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", portRoot))
		checkError(err)

		runClient(localAddr, rootAddr)
	}

}
