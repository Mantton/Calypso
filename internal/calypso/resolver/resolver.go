package resolver

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/collections"
	"github.com/mantton/calypso/internal/calypso/lexer"
)

type Resolver struct {
	Errors lexer.ErrorList
	scopes collections.Stack[*Scope]
}

func New() *Resolver {
	return &Resolver{
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

func (r *Resolver) Define(ident *ast.IdentifierExpression) {
	if r.scopes.IsEmpty() {
		return
	}

	s, ok := r.scopes.Head()

	// No Scope
	if !ok {
		panic("unbalanced scopes")
	}

	if _, ok := s.Get(ident.Value); !ok {

		msg := fmt.Sprintf("Variable `%s` is not declared in current scope", ident.Value)
		panic(r.error(msg, ident))
	}

	s.Define(ident.Value)
}

func (r *Resolver) Declare(ident *ast.IdentifierExpression) {
	if r.scopes.IsEmpty() {
		return
	}

	s, ok := r.scopes.Head()

	// No Scope
	if !ok {
		panic("unbalanced scopes")
	}

	if s.Has(ident.Value) {
		msg := fmt.Sprintf("Variable `%s` is already declared in current scope", ident.Value)
		panic(r.error(msg, ident))
	}

	s.Declare(ident.Value)
}

func (r *Resolver) ExpectInFile(ident *ast.IdentifierExpression) {
	if r.scopes.IsEmpty() {
		panic("unbalanced scopes")
	}
	for i := r.scopes.Length() - 1; i >= 0; i-- {
		s, ok := r.scopes.Get(i)

		if !ok {
			panic("unbalanced scope")
		}

		if s.Has(ident.Value) {
			return
		}
	}

	msg := fmt.Sprintf("Variable `%s` cannot be found in the current scope", ident.Value)
	panic(r.error(msg, ident))
}

// * Resolvers
func (r *Resolver) ResolveFile(file *ast.File) {

	r.enterScope()

	if len(file.Declarations) != 0 {
		for _, decl := range file.Declarations {
			r.resolveDeclaration(decl)
		}
	}

	r.leaveScope()
}

func (r *Resolver) resolveDeclaration(decl ast.Declaration) {

	func() {
		defer func() {
			if err := recover(); err != nil {
				if err, y := err.(lexer.Error); y {
					r.Errors.Add(err)
				} else {
					panic(r)
				}
			}
		}()
		switch decl := decl.(type) {
		case *ast.ConstantDeclaration:
			stmt := decl.Stmt
			r.resolveStatement(stmt)
		case *ast.FunctionDeclaration:
			expr := decl.Func
			r.resolveExpression(expr)
		case *ast.StatementDeclaration:
			stmt := decl.Stmt
			r.resolveStatement(stmt)
		case *ast.StandardDeclaration, *ast.TypeDeclaration:
			// TODO: Add To Scope
			break
		default:
			msg := fmt.Sprintf("expression declaration not implemented, %T", decl)
			panic(msg)
		}
	}()

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
	case *ast.AliasStatement:
		r.resolveAliasStatement(stmt)
	case *ast.StructStatement:
		r.resolveStructStatement(stmt)
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

func (r *Resolver) error(message string, expr ast.Expression) lexer.Error {
	return lexer.Error{
		Range:   expr.Range(),
		Message: message,
	}
}
