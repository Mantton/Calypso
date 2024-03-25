package commands

import (
	"fmt"
	"os"
	"strings"
)

func Action(args []string) {
	isInvalid := len(args) < 1

	if isInvalid {
		fmt.Println(usage())
		return
	}

	action := args[0]
	arguments := args[1:]

	exitCode := 0
	switch strings.ToLower(action) {
	case "build":
		err := build(arguments)
		if err != nil {
			fmt.Println(err)
			exitCode = 1
		}
	default:
		fmt.Println(usage())
	}
	os.Exit(exitCode)
}

func usage() string {
	return `
Usage:
		calypso [COMMAND] ARGUMENTS

Commands:
		build
		help

Note: Use "calypso help [COMMAND] for more information about a specific command"
`
}
