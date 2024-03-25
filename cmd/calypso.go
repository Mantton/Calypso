package main

import (
	"os"

	"github.com/mantton/calypso/internal/calypso/commands"
)

func main() {
	args := os.Args[1:] // gets args without program path
	commands.Action(args)
}
