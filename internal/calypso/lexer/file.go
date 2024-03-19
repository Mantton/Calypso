package lexer

import (
	"fmt"
	"os"
	"strings"

	"github.com/mantton/calypso/internal/calypso/token"
)

type File struct {
	Chars  []rune
	Length int
	Name   string
	Path   string
	Lines  []string
	Tokens []token.ScannedToken
}

func NewFile(path string) (*File, error) {
	// Read File
	data, err := os.ReadFile(path)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	input := string(data)
	lines := strings.Split(input, "\n")
	chars := []rune(input)
	return &File{
		Chars:  chars,
		Length: len(chars),
		Name:   path,
		Path:   path,
		Lines:  lines,
		Tokens: nil,
	}, nil
}
