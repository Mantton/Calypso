package compiler

import "github.com/mantton/calypso/internal/calypso/ast"

func (c *Compiler) VisitIfStatement(n *ast.IfStatement)                 {}
func (c *Compiler) VisitExpressionStatement(n *ast.ExpressionStatement) {}
func (c *Compiler) VisitWhileStatement(n *ast.WhileStatement)           {}
func (c *Compiler) VisitReturnStatement(n *ast.ReturnStatement)         {}
func (c *Compiler) VisitBlockStatement(n *ast.BlockStatement)           {}
func (c *Compiler) VisitVariableStatement(n *ast.VariableStatement)     {}
func (c *Compiler) VisitFunctionStatement(n *ast.FunctionStatement)     {}
func (c *Compiler) VisitAliasStatement(n *ast.AliasStatement)           {}
func (c *Compiler) VisitStructStatement(n *ast.StructStatement)         {}
