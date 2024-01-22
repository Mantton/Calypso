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

func (t *TypeChecker) mustValidate(provided, expected ExpressionType, node ast.Node) {

	result := t.validate(provided, expected)
	if !result {
		msg := fmt.Sprintf("Expected `%s`, received `%s` instead", expected, provided)
		panic(t.error(msg, node))
	}
}

func (t *TypeChecker) evaluateTypeExpression(expr ast.TypeExpression) ExpressionType {
	switch expr := expr.(type) {
	case *ast.IdentifierTypeExpression:
		return t.evaluateIdentifierTypeExpression(expr)
	case *ast.GenericTypeExpression:
		return t.evaluateGenericTypeExpression(expr)
	}
	msg := fmt.Sprintf("type expression check not implemented, %T", expr)
	panic(msg)
}

// * Utils
func (t *TypeChecker) register(ident string, expr ExpressionType) {
	if t.scopes.IsEmpty() {
		return
	}

	s, ok := t.scopes.Head()

	// No Scope
	if !ok {
		panic("unbalanced scopes")
	}

	msg := fmt.Sprintf("Registering `%s` as `%s`", ident, expr)
	fmt.Println(msg)

	s.Define(ident, expr)
}

func (t *TypeChecker) define(ident *ast.IdentifierExpression, expr ExpressionType) {
	if t.scopes.IsEmpty() {
		return
	}

	s, ok := t.scopes.Head()

	// No Scope
	if !ok {
		panic("unbalanced scopes")
	}

	msg := fmt.Sprintf("Defining `%s` as `%s`", ident.Value, expr)
	fmt.Println(msg)

	if s.Has(ident.Value) {
		msg = fmt.Sprintf("type `%s` is already declared in the current scope", ident.Value)
		panic(t.error(msg, ident))
	}

	s.Define(ident.Value, expr)
}

func (t *TypeChecker) get(ident *ast.IdentifierExpression) ExpressionType {
	if t.scopes.IsEmpty() {
		panic("unbalanced scopes")
	}

	for i := t.scopes.Length() - 1; i >= 0; i-- {
		s, ok := t.scopes.Get(i)

		if !ok {
			panic("unbalanced scope")
		}

		if v, ok := s.Get(ident.Value); ok {
			return v
		}
	}

	msg := fmt.Sprintf("type `%s` cannot not be found in the current scope", ident.Value)
	panic(t.error(msg, ident))
}

func (t *TypeChecker) evaluateIdentifierTypeExpression(expr *ast.IdentifierTypeExpression) ExpressionType {

	gen := t.get(expr.Identifier)

	switch v := gen.(type) {

	case *BaseType:
		return gen
	case *GenericType:
		args := len(v.Params)
		msg := fmt.Sprintf("Generic type `%s` requires %d arguments", v, args)
		panic(t.error(msg, expr))
	}

	panic("bad path")
}

func (t *TypeChecker) evaluateGenericTypeExpression(expr *ast.GenericTypeExpression) ExpressionType {
	expectedType := t.get(expr.Identifier)

	switch expectedType := expectedType.(type) {
	case *BaseType:
		msg := fmt.Sprintf("`%s` is not a generic type", expr.Identifier.Value)
		panic(t.error(msg, expr))
	case *GenericType:

		expected := len(expectedType.Params)
		provided := len(expr.Arguments)

		if expected != provided {
			msg := fmt.Sprintf("Generic type `%s` requires %d arguments, received %d instead.", expectedType, expected, provided)
			panic(t.error(msg, expr))
		}

		args := []ExpressionType{}
		for i, expectedArgument := range expectedType.Params {
			providedArgument := t.evaluateTypeExpression(expr.Arguments[i])

			t.mustValidate(providedArgument, expectedArgument, expr.Arguments[i])
			args = append(args, providedArgument)
		}

		return GenerateGenericType(expr.Identifier.Value, args...)
	}

	panic("bad path")

}
