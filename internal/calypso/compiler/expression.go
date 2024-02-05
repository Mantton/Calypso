package compiler

import "github.com/mantton/calypso/internal/calypso/ast"

// Base Expressions
func (c *Compiler) VisitFunctionExpression(n *ast.FunctionExpression)     {}
func (c *Compiler) VisitIdentifierExpression(n *ast.IdentifierExpression) {}

// Complex Expressions
func (c *Compiler) VisitGroupedExpression(n *ast.GroupedExpression)                   {}
func (c *Compiler) VisitCallExpression(n *ast.CallExpression)                         {}
func (c *Compiler) VisitUnaryExpression(n *ast.UnaryExpression)                       {}
func (c *Compiler) VisitBinaryExpression(n *ast.BinaryExpression)                     {}
func (c *Compiler) VisitAssignmentExpression(n *ast.AssignmentExpression)             {}
func (c *Compiler) VisitIndexExpression(n *ast.IndexExpression)                       {}
func (c *Compiler) VisitPropertyExpression(n *ast.PropertyExpression)                 {}
func (c *Compiler) VisitKeyValueExpression(n *ast.KeyValueExpression)                 {}
func (c *Compiler) VisitCompositeLiteralBodyClause(n *ast.CompositeLiteralBodyClause) {}
