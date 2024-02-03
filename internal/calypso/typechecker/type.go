package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
)

func (c *Checker) evaluateTypeExpression(e ast.TypeExpression) *SymbolInfo {
	switch expr := e.(type) {
	case *ast.IdentifierTypeExpression:
		return c.evaluateIdentifierTypeExpression(expr)

	}
	msg := fmt.Sprintf("type expression check not implemented, %T", e)
	panic(msg)
}

func (c *Checker) evaluateIdentifierTypeExpression(expr *ast.IdentifierTypeExpression) *SymbolInfo {
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
	symbolArgCount := len(sym.GenericArguments)
	var providedArguments []ast.TypeExpression

	if expr.Arguments != nil {
		providedArguments = expr.Arguments.Arguments
	}
	if symbolArgCount != len(providedArguments) {
		msg := fmt.Sprintf("`%s` expects %d generic argument(s) got %d", sym.Name, symbolArgCount, len(providedArguments))
		c.addError(msg, expr.Range())
		return unresolved
	}

	// Return early if there are no generic parameters
	if symbolArgCount == 0 {
		return sym
	}

	specialized := newSymbolInfo(sym.Name, TypeSymbol)
	specialized.ConcreteOf = sym

	// Ensure each argument are valid
	hasError := false
	for i, expectedArg := range sym.GenericArguments {
		arg := providedArguments[i]
		providedArg := c.evaluateTypeExpression(arg)

		err := c.validate(expectedArg, providedArg)

		if err != nil {
			c.addError(
				fmt.Sprintf("cannot pass `%s` expression as generic argument for `%s` of `%s`. %s",
					providedArg.Name,
					expectedArg.Name,
					sym.Name,
					err.Error()),
				arg.Range(),
			)

			specialized.addGenericArgument(unresolved)
			continue
		}

		specialized.addGenericArgument(providedArg)
	}

	if hasError {
		return unresolved
	}

	return specialized
}
