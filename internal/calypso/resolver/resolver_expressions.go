package resolver

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/scope"
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
		r.ExpectInFile(expr.Value)
		return
	}

	// Value in Scope

	// Value in Scope But is Just Declared
	if state == scope.DECLARED {
		panic("cannot read local variable in its own definition")
	}

	// Value IN Scope & Is Defined, do nothing
}

func (r *Resolver) resolveFunctionExpression(expr *ast.FunctionExpression) {
	r.Declare(expr.Name)
	r.Define(expr.Name)

	// Enter Function Scope
	r.enterScope()

	// Declare & Define Parameters
	for _, param := range expr.Params {
		r.Declare(param.Value)
		r.Define(param.Value)
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
		panic("expected identifier")
	}

	r.ExpectInFile(target.Value)

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
