package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const inputFilePath = "messages.txt"

func getLinesChannel(f io.ReadCloser) <-chan string {
	fmt.Println("Reading data from file.")
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
	file, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("Could not open file: %s\n", err)
	}

	lines := getLinesChannel(file)
	for line := range lines {
		fmt.Println("read:", line)
	}
}
