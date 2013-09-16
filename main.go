package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("%s job-dir", os.Args[0])
	}
	root := os.Args[1]

	hankie := hankie{}

	// Register what should run on it
	if err := hankie.Register(root); err != nil {
		log.Fatal(err)
	}

	// Converge docker states to new state
	if err := hankie.Converge(); err != nil {
		log.Fatal(err)
	}
}
