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

type Specializations map[string]types.Type

func NewSpecializations() Specializations {
	return make(Specializations)
}

func (s Specializations) specialize(tParam *types.TypeParam, provided types.Type, c *Checker, expr ast.Expression) error {
	currentSpec, ok := s[tParam.Name()]

	// Unwrap bounded params
	if tT := types.AsTypeParam(provided); tT != nil {
		provided = tT.Unwrapped()
	}

	// No Specialization
	if !ok {
		// Ensure Conformance
		err := types.Conforms(tParam.Constraints, provided)

		if err != nil {
			return err
		}

		s[tParam.Name()] = provided
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
