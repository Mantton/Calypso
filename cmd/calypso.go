package main

import (
	"fmt"
	"os"

	"github.com/mantton/calypso/internal/calypso/evaluator"
)

func main() {
	fmt.Println()

	args := os.Args[1:] // gets args without program path
	isInvalid := len(args) < 1

	if isInvalid {
		fmt.Println("Usage calypso [script]")
	} else {
		path := args[0]

		data, err := os.ReadFile(path)

		if err != nil {
			fmt.Println(err.Error())
			return
		}
		eval := evaluator.New()
		code := eval.Evaluate(path, string(data))

		os.Exit(code)
	}
}
