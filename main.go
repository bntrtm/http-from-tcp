package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

const inputFilePath = "messages.txt"

func main() {
	file, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("Could not open file: %s\n", err)
	}
	defer file.Close()

	fmt.Printf("Reading data from %s\n", inputFilePath)
	fmt.Println("=====================================")

	for {
		bytes := make([]byte, 8)
		bytesRead, err := file.Read(bytes)
		if err != nil {
			if errors.Is(err, io.EOF) {
				os.Exit(0)
			} else {
				log.Fatalf("Error while reading file: %s", err)
			}
		}
		if bytesRead > 0 {
			fmt.Printf("read: %s\n", bytes)
		}
	}
}
