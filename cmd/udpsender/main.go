package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const NETWORK = "udp"

func main() {
	addr, err := net.ResolveUDPAddr(NETWORK, "localhost:42069")
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP(NETWORK, nil, addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf(">")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error during read: %s", err)
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Fatalf("Error during write: %s", err)
		}
	}
}
