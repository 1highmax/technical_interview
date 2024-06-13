package main

import (
	"log"
	"os"
	"shred-tool/shred"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <file>", os.Args[0])
	}

	filePath := os.Args[1]
	err := shred.ShredFile(filePath)
	if err != nil {
		log.Fatalf("Error shredding file: %v", err)
	}
	log.Println("File successfully shredded")
}
