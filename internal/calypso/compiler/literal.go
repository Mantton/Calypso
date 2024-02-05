package compiler

import "github.com/mantton/calypso/internal/calypso/ast"

// Base Literals
func (c *Compiler) VisitIntegerLiteral(n *ast.IntegerLiteral) {}
func (c *Compiler) VisitFloatLiteral(n *ast.FloatLiteral)     {}
func (c *Compiler) VisitStringLiteral(n *ast.StringLiteral)   {}
func (c *Compiler) VisitBooleanLiteral(n *ast.BooleanLiteral) {}
func (c *Compiler) VisitNullLiteral(n *ast.NullLiteral)       {}
func (c *Compiler) VisitVoidLiteral(n *ast.VoidLiteral)       {}

// Complex Literals
func (c *Compiler) VisitCompositeLiteral(n *ast.CompositeLiteral) {}
func (c *Compiler) VisitArrayLiteral(n *ast.ArrayLiteral)         {}
func (c *Compiler) VisitMapLiteral(n *ast.MapLiteral)             {}
