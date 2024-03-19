package main

import (
	"fmt"
	"os"

	"github.com/mantton/calypso/internal/calypso/builder"
)

func main() {

	args := os.Args[1:] // gets args without program path
	isInvalid := len(args) < 1

	if isInvalid {
		fmt.Println("Usage calypso [script]")
	} else {

		path := args[0]
		state := builder.Build(path)

		if !state {
			os.Exit(1)
			return
		}

		os.Exit(0)
	}
}
