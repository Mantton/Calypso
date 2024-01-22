package evaluator

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/parser"
	"github.com/mantton/calypso/internal/calypso/resolver"
	"github.com/mantton/calypso/internal/calypso/typechecker"
)

type Evaluator struct {
}

func New() *Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) Evaluate(filepath, input string) int {

	// Lexer / Scanner
	lexer := lexer.New(input)
	tokens := lexer.AllTokens()

	// Parser
	parser := parser.New(tokens)
	file := parser.Parse()

	if len(file.Errors) != 0 {
		for _, err := range file.Errors {
			fmt.Println(e.ErrorMessage(filepath, err))
		}
		return 1
	}

	// Resolver
	resolver := resolver.New()
	resolver.ResolveFile(file)

	if len(resolver.Errors) != 0 {
		for _, err := range resolver.Errors {
			fmt.Println(e.ErrorMessage(filepath, err))
		}
		return 1
	}

	// Type Checker
	typeChecker := typechecker.New(typechecker.STD)
	typeChecker.CheckFile(file)
	fmt.Println()

	if len(typeChecker.Errors) != 0 {
		for _, err := range typeChecker.Errors {
			fmt.Println(e.ErrorMessage(filepath, err))
		}
		return 1
	}

	fmt.Println("Done")
	return 0
}

func (e *Evaluator) ErrorMessage(filepath string, error *lexer.Error) string {
	return fmt.Sprintf("\n%s:%d:%d\n\t%s", filepath, error.Range.Start.Line, error.Range.Start.Offset, error.Message)
}
