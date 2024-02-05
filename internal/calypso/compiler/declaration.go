package compiler

import "github.com/mantton/calypso/internal/calypso/ast"

func (c *Compiler) VisitConstantDeclaration(n *ast.ConstantDeclaration)       {}
func (c *Compiler) VisitStatementDeclaration(n *ast.StatementDeclaration)     {}
func (c *Compiler) VisitFunctionDeclaration(n *ast.FunctionDeclaration)       {}
func (c *Compiler) VisitStandardDeclaration(n *ast.StandardDeclaration)       {}
func (c *Compiler) VisitTypeDeclaration(n *ast.TypeDeclaration)               {}
func (c *Compiler) VisitExtensionDeclaration(n *ast.ExtensionDeclaration)     {}
func (c *Compiler) VisitConformanceDeclaration(n *ast.ConformanceDeclaration) {}
