package compiler

import "github.com/mantton/calypso/internal/calypso/ast"

func (c *Compiler) VisitConstantDeclaration(n *ast.ConstantDeclaration) {
	n.Stmt.Accept(c)
}

func (c *Compiler) VisitStatementDeclaration(n *ast.StatementDeclaration) {
	n.Stmt.Accept(c)
}

func (c *Compiler) VisitFunctionDeclaration(n *ast.FunctionDeclaration) {
	n.Func.Accept(c)
}

func (c *Compiler) VisitStandardDeclaration(n *ast.StandardDeclaration) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitTypeDeclaration(n *ast.TypeDeclaration) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitExtensionDeclaration(n *ast.ExtensionDeclaration) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitConformanceDeclaration(n *ast.ConformanceDeclaration) {
	panic("MISSING_IMPLEMENTATION")
}
