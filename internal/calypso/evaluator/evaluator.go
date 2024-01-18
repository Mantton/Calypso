package evaluator

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/parser"
)

type Evaluator struct {
}

func New() *Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) Run(filename, input string) error {

	l := lexer.New(filename, input)

	tokens := l.AllTokens()

	p := parser.New(tokens)

	f := p.Parse()

	fmt.Println("Parsed File:", f)

	return nil
}
