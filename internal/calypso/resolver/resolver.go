package resolver

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/collections"
)

type Resolver struct {
	Errors []string
	scopes collections.Stack[*Scope]
}

func New() *Resolver {
	return &Resolver{
		Errors: []string{},
		scopes: collections.Stack[*Scope]{},
	}
}

// * Scopes
func (r *Resolver) enterScope() {
	r.scopes.Push(NewScope())

	if r.scopes.Length() > 1000 {
		panic("exceeded max scope depth") // TODO : Error
	}
}

func (r *Resolver) leaveScope() {
	r.scopes.Pop()
}

func (r *Resolver) Define(ident string) {
	if r.scopes.IsEmpty() {
		return
	}

	s, ok := r.scopes.Head()

	// No Scope
	if !ok {
		panic("unbalanced scopes")
	}

	s.Define(ident)
}

func (r *Resolver) Declare(ident string) {
	if r.scopes.IsEmpty() {
		return
	}

	s, ok := r.scopes.Head()

	// No Scope
	if !ok {
		panic("unbalanced scopes")
	}

	s.Declare(ident)
}

func (r *Resolver) ExpectInFile(ident string) {
	if r.scopes.IsEmpty() {
		panic("unbalanced scopes")
	}
	for i := r.scopes.Length() - 1; i >= 0; i-- {
		s, ok := r.scopes.Get(i)

		if !ok {
			panic("unbalanced scope")
		}

		if s.Has(ident) {
			return
		}
	}

	fmt.Println(ident)
	panic("ident not found in scope")
}

// * Resolvers
func (r *Resolver) ResolveFile(file *ast.File) {

	r.enterScope()
	if len(file.Constants) != 0 {

		for _, decl := range file.Constants {
			r.resolveDeclaration(decl)
		}

	}

	if len(file.Functions) != 0 {
		for _, decl := range file.Functions {
			r.resolveDeclaration(decl)
		}
	}
	r.leaveScope()
}

func (r *Resolver) resolveDeclaration(decl ast.Declaration) {
	switch decl := decl.(type) {
	case *ast.ConstantDeclaration:
		stmt := decl.Stmt
		r.resolveStatement(stmt)
	case *ast.FunctionDeclaration:
		expr := decl.Func
		r.resolveExpression(expr)
	default:
		msg := fmt.Sprintf("expression declaration not implemented, %T", decl)
		panic(msg)

	}

}

func (r *Resolver) resolveStatement(stmt ast.Statement) {
	switch stmt := stmt.(type) {
	case *ast.VariableStatement:
		r.resolveVariableStatement(stmt)
	case *ast.BlockStatement:
		r.resolveBlockStatement(stmt)
	case *ast.ExpressionStatement:
		r.resolveExpressionStatement(stmt)
	case *ast.IfStatement:
		r.resolveIfStatement(stmt)
	case *ast.ReturnStatement:
		r.resolveReturnStatement(stmt)
	case *ast.WhileStatement:
		r.resolveWhileStatement(stmt)
	default:
		msg := fmt.Sprintf("statement resolution not implemented, %T", stmt)
		panic(msg)
	}
}

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
	case *ast.IntegerLiteral, *ast.StringLiteral, *ast.FloatLiteral, *ast.BooleanLiteral, *ast.NullLiteral, *ast.VoidLiteral:
		return // Do nothing, no expressions to parse
	default:
		msg := fmt.Sprintf("expression resolution not implemented, %T", expr)
		panic(msg)
	}

}
