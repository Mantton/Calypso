package main

import (
	"fmt"
	"os"

	"github.com/mantton/calypso/internal/repl"
)

func main() {
	fmt.Println("CALYPSO")
	fmt.Println()

	args := os.Args[1:] // gets args without program path
	runREPL := len(args) < 1

	if runREPL {
		repl.Run()
	} else {
		fmt.Println("Execute file")
	}
}
