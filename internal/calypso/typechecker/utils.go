package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/token"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) addError(msg string, pos token.SyntaxRange) {
	c.Errors.Add(&lexer.CompilerError{
		Message: msg,
		Range:   pos,
		File:    c.file.LexerFile,
	})
}

func (c *Checker) specialize(m types.Specialization, tParam *types.TypeParam, provided types.Type, expr ast.Expression) error {
	currentSpec, ok := m[tParam]

	// No Specialization
	if !ok {
		// Ensure Conformance
		err := types.Conforms(tParam.Constraints, provided)

		if err != nil {
			return err
		}

		m[tParam] = provided
		fmt.Printf("\t[Resolver] Specialized %s as %s\n", tParam, provided)
		return nil
	}

	// Specialization Found
	// has been specialized, ensure strict match
	temp := types.NewVar("", currentSpec)
	err := c.validateAssignment(temp, provided, expr, false)

	if err != nil {
		return err
	}

	return nil

}
