package resolver

import "github.com/mantton/calypso/internal/calypso/ast"

// * Statements
func (r *Resolver) resolveVariableStatement(stmt *ast.VariableStatement) {
	r.Declare(stmt.Identifier)
	r.resolveExpression(stmt.Value)
	r.Define(stmt.Identifier)
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
	// TODO: resolve alias identifier
}

func (r *Resolver) resolveStructStatement(stmt *ast.StructStatement) {
	// TODO: resolve struct identifier
}
