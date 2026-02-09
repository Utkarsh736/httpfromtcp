package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	// Resolve UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal("ResolveUDPAddr failed:", err)
	}

	// Dial UDP (creates *UDPConn)
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatal("DialUDP failed:", err)
	}
	defer conn.Close()

	// Read from stdin
	reader := bufio.NewReader(os.Stdin)

	for {
		// Print prompt
		fmt.Print("> ")

		// Read line from stdin
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println("ReadString error:", err)
			continue
		}

		// Write line to UDP (includes \n)
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Println("Write error:", err)
		}
	}
}
