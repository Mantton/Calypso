package resolver

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
)

// * Expressions
func (r *Resolver) resolveIdentifierExpression(expr *ast.IdentifierExpression) {
	s, ok := r.scopes.Head()

	// No Scope
	if !ok {
		panic("unbalanced scopes")
	}

	state, ok := s.Get(expr.Value)

	// Value Not in Scope
	if !ok {
		r.ExpectInFile(expr)
		return
	}

	// Value in Scope

	// Value in Scope But is Just Declared
	if state == DECLARED {
		msg := fmt.Sprintf("`%s` cannot be used in it's own definition", expr.Value)
		panic(r.error(msg, expr))
	}

	// Value IN Scope & Is Defined, do nothing
}

func (r *Resolver) resolveFunctionExpression(expr *ast.FunctionExpression) {
	r.Declare(expr.Identifier)
	r.Define(expr.Identifier)

	// Enter Function Scope
	r.enterScope()

	// Declare & Define Parameters
	for _, param := range expr.Params {
		r.Declare(param)
		r.Define(param)
	}

	// Resolve Function Body
	r.resolveStatement(expr.Body)

	// Resolution Complete, leave function scope
	r.leaveScope()
}

func (r *Resolver) resolveAssignmentExpression(expr *ast.AssignmentExpression) {
	// TODO: what if you're assigning a property or an index
	r.resolveExpression(expr.Value)

	target, ok := expr.Target.(*ast.IdentifierExpression)

	if !ok {
		panic(r.error("expected identifier expression", expr.Value))
	}

	r.ExpectInFile(target)

}

func (r *Resolver) resolveBinaryExpression(expr *ast.BinaryExpression) {

	r.resolveExpression(expr.Left)
	r.resolveExpression(expr.Right)
}

func (r *Resolver) resolveUnaryExpression(expr *ast.UnaryExpression) {

	r.resolveExpression(expr.Expr)
}

func (r *Resolver) resolveCallExpression(expr *ast.CallExpression) {
	r.resolveExpression(expr.Target)

	for _, arg := range expr.Arguments {
		r.resolveExpression(arg)
	}
}

func (r *Resolver) resolveGroupedExpression(expr *ast.GroupedExpression) {
	r.resolveExpression(expr.Expr)
}

func (r *Resolver) resolvePropertyExpression(expr *ast.PropertyExpression) {
	r.resolveExpression(expr.Target)
}

func (r *Resolver) resolveIndexExpression(expr *ast.IndexExpression) {
	r.resolveExpression(expr.Target)
}

func (r *Resolver) resolveArrayLiteral(expr *ast.ArrayLiteral) {
	for _, elem := range expr.Elements {
		r.resolveExpression(elem)
	}
}

func (r *Resolver) resolveMapLiteral(expr *ast.MapLiteral) {

	for key, value := range expr.Pairs {
		r.resolveExpression(value)
		r.resolveExpression(key)

	}

}
