package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
)

func (t *TypeChecker) validate(provided, expected ExpressionType) bool {

	switch provided := provided.(type) {
	case *BaseType:
		ex, ok := expected.(*BaseType)

		if !ok {
			return false
		}

		// Both are the same type
		if ex.Ident() == provided.Ident() {
			return true
		}

		if ex.Ident() == "AnyLiteral" {
			return true
		}

		return false
	case *GenericType:
		ex, ok := expected.(*GenericType)
		if !ok {
			return false
		}

		if ex.Ident() != provided.Ident() {
			return false
		}

		if len(ex.Params) != len(provided.Params) {
			return false
		}

		for i, exParam := range ex.Params {
			provParam := provided.Params[i]

			res := t.validate(provParam, exParam)

			if !res {
				return false
			}
		}

	}

	return true
}

func (t *TypeChecker) mustValidate(provided, expected ExpressionType) {

	result := t.validate(provided, expected)
	if !result {
		msg := fmt.Sprintf("Expected `%s`, received `%s` instead", expected, provided)
		panic(msg)
	}
}

func (t *TypeChecker) evaluateTypeExpression(expr ast.TypeExpression) ExpressionType {
	switch expr := expr.(type) {
	case *ast.IdentifierTypeExpression:
		return t.evaluateIdentifierTypeExpression(expr)
	case *ast.GenericTypeExpression:
		return t.evaluateGenericTypeExpression(expr)
	}
	panic("evaluate type")
}

// * Utils
func (t *TypeChecker) define(ident string, expr ExpressionType) {
	if t.scopes.IsEmpty() {
		return
	}

	s, ok := t.scopes.Head()

	// No Scope
	if !ok {
		panic("unbalanced scopes")
	}

	msg := fmt.Sprintf("Defining `%s` as `%s`", ident, expr)
	fmt.Println(msg)

	s.Define(ident, expr)
}

func (t *TypeChecker) get(ident string) ExpressionType {
	if t.scopes.IsEmpty() {
		panic("unbalanced scopes")
	}

	for i := t.scopes.Length() - 1; i >= 0; i-- {
		s, ok := t.scopes.Get(i)

		if !ok {
			panic("unbalanced scope")
		}

		if v, ok := s.Get(ident); ok {
			return v
		}
	}

	fmt.Println(ident)
	panic("type not found in scope")
}

func (t *TypeChecker) evaluateIdentifierTypeExpression(expr *ast.IdentifierTypeExpression) ExpressionType {

	gen := t.get(expr.Identifier.Value)

	switch gen.(type) {

	case *BaseType:
		return gen
	case *GenericType:
		panic("Generic type requires x type arguments")
	}

	panic("bad path")
}

func (t *TypeChecker) evaluateGenericTypeExpression(expr *ast.GenericTypeExpression) ExpressionType {
	expectedType := t.get(expr.Identifier.Value)

	switch expectedType := expectedType.(type) {
	case *BaseType:
		msg := fmt.Sprintf("`%s` is not a generic type", expr.Identifier.Value)
		panic(msg)
	case *GenericType:

		expected := len(expectedType.Params)
		provided := len(expr.Arguments)

		if expected != provided {
			msg := fmt.Sprintf("Generic type `%s` requires %d arguments, received %d instead.", expr.Identifier.Value, expected, provided)
			panic(msg)
		}

		args := []ExpressionType{}
		for i, expectedArgument := range expectedType.Params {
			providedArgument := t.evaluateTypeExpression(expr.Arguments[i])

			t.mustValidate(providedArgument, expectedArgument)
			args = append(args, providedArgument)
		}

		return GenerateGenericType(expr.Identifier.Value, args...)
	}

	panic("bad path")

}
