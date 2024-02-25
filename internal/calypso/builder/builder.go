package builder

import (
	"fmt"
	"strings"
	"time"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/irgen"
	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/parser"
	"github.com/mantton/calypso/internal/calypso/resolver"
	"github.com/mantton/calypso/internal/calypso/ssagen"
	t "github.com/mantton/calypso/internal/calypso/typechecker"
)

func Build(filepath, input string) *ast.File {

	// Lexer / Scanner

	fmt.Println("\n[Parser] Starting")
	start := time.Now()

	lexer := lexer.New(input)
	tokens := lexer.AllTokens()

	lines := strings.Split(input, "\n")

	// Parser
	parser := parser.New(tokens)
	file := parser.Parse()

	if len(file.Errors) != 0 {
		for _, err := range file.Errors {
			fmt.Println(ErrorMessage(filepath, err, lines))
		}
		return nil
	}
	duration := time.Since(start)

	fmt.Println("[Parser] Completed.", "Took", duration)

	// Resolver
	fmt.Println("\n[Resolver] Starting")
	start = time.Now()

	resolver := resolver.New()
	resolver.ResolveFile(file)

	if len(resolver.Errors) != 0 {
		for _, err := range resolver.Errors {
			fmt.Println(ErrorMessage(filepath, err, lines))
		}
		return nil
	}
	duration = time.Since(start)

	fmt.Println("[Resolver] Completed.", "Took", duration)

	fmt.Println("\n[TypeChecker] Starting")
	start = time.Now()

	checker := t.New(t.STD)
	sc := checker.CheckFile(file)

	if len(checker.Errors) != 0 {
		for _, err := range checker.Errors {
			fmt.Println(ErrorMessage(filepath, err, lines))
		}
		return nil
	}
	duration = time.Since(start)

	fmt.Println("[TypeChecker] Completed.", "Took", duration)

	fmt.Println("\n[SSAGen] Starting")
	exec := ssagen.Generate(file, sc)
	fmt.Println("[SSAGen] Completed.", "Took", duration)

	fmt.Println("\n[IRGen] Starting")
	irgen.Compile(exec)
	fmt.Println("[IRGen] Completed.", "Took", duration)
	return file
}

func ErrorMessage(filepath string, err *lexer.Error, lines []string) string {
	msg := fmt.Sprintf("\n%s:%d:%d -> %s", filepath, err.Range.Start.Line, err.Range.Start.Offset, err.Message)
	msg += fmt.Sprintf("\n\t%s", lines[max(0, err.Range.Start.Line-1)])
	// TODO: Arrow
	return msg
}
