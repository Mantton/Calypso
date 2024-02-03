package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (c *Checker) checkExpression(expr ast.Expression) {
	switch expr := expr.(type) {
	case *ast.FunctionExpression:
		c.checkFunctionExpression(expr)

	default:
		msg := fmt.Sprintf("expression check not implemented, %T", expr)
		panic(msg)
	}
}

func (c *Checker) checkFunctionExpression(expr *ast.FunctionExpression) {
	c.enterScope()

	// TODO: Params

	c.checkStatement(expr.Body)
	c.leaveScope(false)
}

func (c *Checker) evaluateExpression(expr ast.Expression) *SymbolInfo {
	c.currentNode = expr

	switch expr := expr.(type) {
	// Literals
	case *ast.IntegerLiteral:
		return c.resolveLiteral(INTEGER)
	case *ast.BooleanLiteral:
		return c.resolveLiteral(BOOLEAN)
	case *ast.FloatLiteral:
		return c.resolveLiteral(FLOAT)
	case *ast.StringLiteral:
		return c.resolveLiteral(STRING)
	case *ast.NullLiteral:
		return c.resolveLiteral(NULL)
	case *ast.VoidLiteral:
		return c.resolveLiteral(VOID)
	case *ast.ArrayLiteral:
		return c.evaluateArrayLiteral(expr)
	case *ast.IdentifierExpression:
		return c.evaluateIdentifierExpression(expr)
	case *ast.UnaryExpression:
		return c.evaluateUnaryExpression(expr)
	case *ast.GroupedExpression:
		return c.evaluateGroupedExpression(expr)
	default:
		msg := fmt.Sprintf("expression evaluation not implemented, %T", expr)
		panic(msg)
	}
}

func (c *Checker) evaluateIdentifierExpression(expr *ast.IdentifierExpression) *SymbolInfo {

	s, ok := c.find(expr.Value)

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is not defined", expr.Value),
			expr.Range(),
		)

		return unresolved
	}

	return s.TypeDesc
}

func (c *Checker) evaluateUnaryExpression(expr *ast.UnaryExpression) *SymbolInfo {
	op := expr.Op

	provided := c.evaluateExpression(expr.Expr)
	fmt.Println("[Unary]", provided)

	switch op {
	case token.NOT:
		// returns a boolean
	case token.SUB:
		// return the same type

	default:
		c.addError("Unsupported Unary Operand", expr.Expr.Range())
	}

	return unresolved

}

func (c *Checker) evaluateGroupedExpression(expr *ast.GroupedExpression) *SymbolInfo {
	return c.evaluateExpression(expr.Expr)
}

func (c *Checker) evaluateArrayLiteral(expr *ast.ArrayLiteral) *SymbolInfo {
	conc := c.resolveLiteral(ARRAY)
	symbol := newSymbolInfo(conc.Name, TypeSymbol)
	elementType := c.evaluateExpressionList(expr.Elements)
	symbol.ConcreteOf = conc
	symbol.addGenericArgument(elementType)
	return symbol
}

func (c *Checker) evaluateExpressionList(exprs []ast.Expression) *SymbolInfo {

	if len(exprs) == 0 {
		// No Elements, Array Can Contain Any Element
		return c.resolveLiteral(ANY)
	}

	var expected *SymbolInfo

	for _, expr := range exprs {
		if expected == nil {
			expected = c.evaluateExpression(expr)
			continue
		}

		provided := c.evaluateExpression(expr)

		err := c.validate(expected, provided)

		// If Unable to validate type, simple set list type as any
		if err != nil {
			return c.resolveLiteral(ANY)
		}

	}

	return expected

}
