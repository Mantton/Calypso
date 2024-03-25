package builder

import (
	"fmt"
	"time"

	"github.com/mantton/calypso/internal/calypso/commands/utils"
	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/parser"
	"github.com/mantton/calypso/internal/calypso/typechecker"
)

func CompileFileSet(set *utils.FileSet) error {

	astSet, err := parser.ParseFileSet(set)

	if err != nil {
		return err
	}

	fmt.Println(astSet)
	return nil
}

func Build(filepath string) bool {

	// Lexer / Scanner

	fmt.Println("\n[Parser] Starting")
	start := time.Now()

	lFile, err := lexer.NewFile(filepath)
	if err != nil {
		fmt.Printf("Error Reading File. %s; %s\n", filepath, err)
		return false
	}

	lexer := lexer.New(lFile)
	lexer.ScanAll()

	// Parser
	parser := parser.New(lFile)
	aFile := parser.Parse()

	if len(aFile.Errors) != 0 {
		for _, err := range aFile.Errors {
			fmt.Println(err)
		}
		return false
	}

	duration := time.Since(start)

	fmt.Println("[Parser] Completed.", "Took", duration)

	// type checker
	fmt.Println("\n[TypeChecker] Starting")
	start = time.Now()

	checker := typechecker.New(typechecker.STD, aFile)
	_ = checker.Check()

	if len(checker.Errors) != 0 {
		for _, err := range checker.Errors {
			fmt.Println(err)
		}
		return false
	}
	duration = time.Since(start)
	fmt.Println("[TypeChecker] Completed.", "Took", duration)

	// LIRGen
	// fmt.Println("\n[SSAGen] Starting")
	// start = time.Now()
	// exec := ssagen.Generate(file, sc)
	// duration = time.Since(start)
	// fmt.Println("[SSAGen] Completed.", "Took", duration)

	// LLVM IRGen
	// fmt.Println("\n[IRGen] Starting")
	// start = time.Now()
	// irgen.Compile(exec)
	// duration = time.Since(start)
	// fmt.Println("[IRGen] Completed.", "Took", duration)
	return true
}

func ErrorMessage(filepath string, err *lexer.CompilerError, lines []string) string {
	msg := fmt.Sprintf("\n%s:%d:%d -> %s", filepath, err.Range.Start.Line, err.Range.Start.Offset, err.Message)
	msg += fmt.Sprintf("\n\t%s", lines[max(0, err.Range.Start.Line-1)])
	// TODO: Arrow
	return msg
}
