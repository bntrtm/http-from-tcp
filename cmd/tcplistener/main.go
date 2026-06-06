package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	fmt.Println("Reading data from TCP connection.")
	fmt.Println("=====================================")

	lines := make(chan string)

	go func() {
		defer f.Close()
		defer close(lines)

		currentLine := ""
		for {
			bytes := make([]byte, 8)
			bytesRead, err := f.Read(bytes)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				} else {
					log.Printf("Error while reading file: %s", err)
					return
				}
			}
			if bytesRead > 0 {
				currentLine += string(bytes)
				if strings.Contains(currentLine, "\n") {
					splits := strings.Split(currentLine, "\n")
					lines <- splits[0]
					currentLine = strings.Join(splits[1:], "\n")
				}
			}
		}
	}()

	return lines
}

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

		lines := getLinesChannel(conn)
		for line := range lines {
			fmt.Println(line)
		}
	}
}
