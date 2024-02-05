package compiler

import "github.com/mantton/calypso/internal/calypso/ast"

func (c *Compiler) VisitIdentifierTypeExpression(n *ast.IdentifierTypeExpression) {}
func (c *Compiler) VisitArrayTypeExpression(n *ast.ArrayTypeExpression)           {}
func (c *Compiler) VisitMapTypeExpression(n *ast.MapTypeExpression)               {}
