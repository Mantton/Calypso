package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
)

func (t *TypeChecker) validate(provided, expected ExpressionType) {
	panic("validate")
}

func (t *TypeChecker) evaluateTypeExpression(expr ast.TypeExpression) ExpressionType {
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

	msg := fmt.Sprintf("Declaring `%s` as `%s`", ident, expr.Name())
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
