package evaluator

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/parser"
	"github.com/mantton/calypso/internal/calypso/resolver"
)

type Evaluator struct {
}

func New() *Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) Run(filename, input string) error {

	// Lexer / Scanner
	lexer := lexer.New(filename, input)

	tokens := lexer.AllTokens()

	// TODO: Error Check Tokens

	// Parser
	parser := parser.New(tokens)

	file := parser.Parse()

	fmt.Println("Parsed File:", file)

	// TODO:: Error Check File

	// Resolver
	resolver := resolver.New()
	resolver.ResolveFile(file)
	// TODO: Error Check Resolver

	if len(resolver.Errors) != 0 {
		for _, err := range resolver.Errors {
			fmt.Println(err)
		}
	}

	// TODO: Type Checker
	return nil
}
