package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/symbols"
)

func (c *Checker) evaluateTypeExpression(e ast.TypeExpression) *symbols.SymbolInfo {
	switch expr := e.(type) {
	case *ast.IdentifierTypeExpression:
		return c.evaluateIdentifierTypeExpression(expr)

	}
	msg := fmt.Sprintf("type expression check not implemented, %T", e)
	panic(msg)
}

func (c *Checker) evaluateIdentifierTypeExpression(expr *ast.IdentifierTypeExpression) *symbols.SymbolInfo {
	// Identifier means this is predefined

	// Get from table
	sym, ok := c.resolveGenericArgument(expr.Identifier.Value)

	if !ok {
		msg := fmt.Sprintf("Unable to locate `%s`", expr.Identifier.Value)
		c.addError(msg, expr.Range())
		return unresolved
	}

	// Ensure it is a type
	if !c.isType((sym.Type)) {
		msg := fmt.Sprintf("`%s` is not a type", expr.Identifier.Value)
		c.addError(msg, expr.Range())
		return unresolved
	}

	// Validate Types
	expectedArgumentCount := len(sym.GenericParams)
	var providedArguments []ast.TypeExpression

	if expr.Arguments != nil {
		providedArguments = expr.Arguments.Arguments
	}
	if expectedArgumentCount != len(providedArguments) {
		msg := fmt.Sprintf("`%s` expects %d generic argument(s) got %d", sym.Name, expectedArgumentCount, len(providedArguments))
		c.addError(msg, expr.Range())
		return unresolved
	}

	// Return early if there are no generic parameters
	if expectedArgumentCount == 0 {
		return sym
	}

	specialized := symbols.NewSymbol(sym.Name, symbols.TypeSymbol)
	specialized.SpecializedOf = sym

	// Ensure each argument are valid
	hasError := false
	for i, expected := range sym.GenericParams {
		arg := providedArguments[i]
		provided := c.evaluateTypeExpression(arg)

		var err error
		if expected.Type == symbols.GenericTypeSymbol {
			_, ok := sym.Specializations[expected]
			if !ok {
				err = c.specialize(specialized.Specializations, expected, provided)
			}
		} else {
			err = c.validate(expected, provided, specialized.Specializations)
		}

		if err != nil {
			c.addError(
				fmt.Sprintf("cannot pass `%s` expression as argument for `%s` of `%s`. %s",
					provided.Name,
					expected.Name,
					sym.Name,
					err.Error()),
				arg.Range(),
			)
			continue
		}
	}

	if hasError {
		return unresolved
	}

	return specialized
}
