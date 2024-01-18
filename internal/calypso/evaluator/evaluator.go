package evaluator

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/lexer"
)

type Evaluator struct {
}

func New() *Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) Run(filename, input string) error {

	l := lexer.New(filename, input)

	tokens := l.AllTokens()

	fmt.Println(tokens)

	return nil
}
