package main

import (
	"fmt"
	"log"
	"net"

	"github.com/bntrtm/http-from-tcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Could not establish TCP connection: %s\n", err)
			return
		}
		fmt.Println("Connection accepted:", conn.LocalAddr())

		r, err := request.RequestFromReader(conn)
		if err != nil {
			panic(err)
		}

		format := "Request line:\n- Method: %s\n- Target: %s\n- Version: %s"
		fmt.Printf(format, r.RequestLine.Method, r.RequestLine.RequestTarget, r.RequestLine.HTTPVersion)

	}
}
