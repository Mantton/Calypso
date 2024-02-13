package main

import (
	"fmt"
	"os"

	"github.com/mantton/calypso/internal/calypso/builder"
)

func main() {

	// panicMode := flag.Bool("panic", false, "go panics will not be handled")
	// flag.Parse()

	// if !*panicMode {
	// 	defer func() {
	// 		if r := recover(); r != nil {
	// 			fmt.Println(r)
	// 			os.Exit(1)
	// 		}
	// 	}()
	// }

	args := os.Args[1:] // gets args without program path
	isInvalid := len(args) < 1

	if isInvalid {
		fmt.Println("Usage calypso [script]")
	} else {
		path := args[0]

		data, err := os.ReadFile(path)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		builder.Build(path, string(data))

		os.Exit(0)
	}
}
