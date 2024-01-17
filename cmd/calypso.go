package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("CALYPSO")
	args := os.Args[1:] // gets args without program path
	fmt.Println(args)

	runREPL := len(args) < 1

	if runREPL {
		fmt.Println("Run REPL")
	} else {
		fmt.Println("Execute file")
	}
}
