package resolver

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
)

func (r *Resolver) resolveExpression(expr ast.Expression) {
	switch expr := expr.(type) {
	case *ast.IdentifierExpression:
		r.resolveIdentifierExpression(expr)
	case *ast.FunctionExpression:
		r.resolveFunctionExpression(expr)
	case *ast.AssignmentExpression:
		r.resolveAssignmentExpression(expr)
	case *ast.BinaryExpression:
		r.resolveBinaryExpression(expr)
	case *ast.UnaryExpression:
		r.resolveUnaryExpression(expr)
	case *ast.CallExpression:
		r.resolveCallExpression(expr)
	case *ast.GroupedExpression:
		r.resolveGroupedExpression(expr)
	case *ast.PropertyExpression:
		r.resolvePropertyExpression(expr)
	case *ast.IndexExpression:
		r.resolveIndexExpression(expr)
	case *ast.ArrayLiteral:
		r.resolveArrayLiteral(expr)
	case *ast.MapLiteral:
		r.resolveMapLiteral(expr)
	case *ast.CompositeLiteral:
		r.resolveCompositeLiteral(expr)
	case *ast.IntegerLiteral, *ast.StringLiteral, *ast.CharLiteral, *ast.FloatLiteral, *ast.BooleanLiteral, *ast.NullLiteral, *ast.VoidLiteral:
		return // Do nothing, no expressions to parse
	default:
		msg := fmt.Sprintf("expression resolution not implemented, %T", expr)
		panic(msg)
	}

}

// * Expressions
func (r *Resolver) resolveIdentifierExpression(expr *ast.IdentifierExpression) {
	s := r.expect(expr)

	if s.State == SymbolDeclared {
		panic("use of variable before definition")
	}
}

func (r *Resolver) resolveFunctionExpression(expr *ast.FunctionExpression) {

	s := newSymbolInfo(expr.Identifier.Value, FunctionSymbol)
	r.declare(s, expr.Identifier)
	r.define(s, expr.Identifier)

	// Enter Function Scope
	r.enterScope()

	// Declare & Define Parameters
	for _, param := range expr.Params {
		p := newSymbolInfo(param.Value, VariableSymbol)
		r.declare(p, param)
		r.define(p, param)
	}

	// Resolve Function Body
	r.resolveStatement(expr.Body)

	// Resolution Complete, leave function scope
	r.leaveScope(false)
}

func (r *Resolver) resolveAssignmentExpression(expr *ast.AssignmentExpression) {
	r.resolveExpression(expr.Value)

	target, ok := expr.Target.(*ast.IdentifierExpression)

	if !ok {
		panic(r.error("expected identifier expression", expr.Value))
	}

	r.expect(target)

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
	r.resolveExpression(expr.Property)
}

func (r *Resolver) resolveIndexExpression(expr *ast.IndexExpression) {
	r.resolveExpression(expr.Target)
	r.resolveExpression(expr.Index)
}

func (r *Resolver) resolveArrayLiteral(expr *ast.ArrayLiteral) {
	for _, elem := range expr.Elements {
		r.resolveExpression(elem)
	}
}

func (r *Resolver) resolveMapLiteral(expr *ast.MapLiteral) {

	for _, pairs := range expr.Pairs {
		r.resolveExpression(pairs.Key)
		r.resolveExpression(pairs.Value)

	}

}

func (r *Resolver) resolveCompositeLiteral(expr *ast.CompositeLiteral) {

	r.resolveIdentifierExpression(expr.Identifier)

	for _, pair := range expr.Pairs {
		// r.resolveExpression(pair.Key) // TODO: Sure we don't want to resolve this?
		r.resolveExpression(pair.Value)
	}
}
