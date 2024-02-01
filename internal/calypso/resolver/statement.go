package resolver

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
)

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

// * Statements
func (r *Resolver) resolveVariableStatement(stmt *ast.VariableStatement) {
	s := newSymbolInfo(stmt.Identifier.Value, VariableSymbol)
	r.declare(s, stmt.Identifier)
	r.resolveExpression(stmt.Value)
	r.define(s, stmt.Identifier)
}

func (r *Resolver) resolveBlockStatement(block *ast.BlockStatement) {
	for _, stmt := range block.Statements {
		r.resolveStatement(stmt)
	}
}

func (r *Resolver) resolveExpressionStatement(stmt *ast.ExpressionStatement) {
	r.resolveExpression(stmt.Expr)
}

func (r *Resolver) resolveIfStatement(stmt *ast.IfStatement) {
	r.resolveExpression(stmt.Condition)
	r.resolveStatement(stmt.Action)

	if stmt.Alternative != nil {
		r.resolveStatement(stmt.Alternative)
	}
}

func (r *Resolver) resolveReturnStatement(stmt *ast.ReturnStatement) {
	if stmt.Value == nil {
		return
	}

	r.resolveExpression(stmt.Value)
}

func (r *Resolver) resolveWhileStatement(stmt *ast.WhileStatement) {
	r.resolveExpression(stmt.Condition)
	r.resolveStatement(stmt.Action)
}

func (r *Resolver) resolveAliasStatement(stmt *ast.AliasStatement) {
	s := newSymbolInfo(stmt.Identifier.Value, AliasSymbol)
	r.declare(s, stmt.Identifier)
	r.define(s, stmt.Identifier)
}

func (r *Resolver) resolveStructStatement(stmt *ast.StructStatement) {
	s := newSymbolInfo(stmt.Identifier.Value, StructSymbol)
	r.declare(s, stmt.Identifier)
	r.define(s, stmt.Identifier)
}
