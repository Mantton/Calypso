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
	case *ast.BinaryExpression:
		return c.evaluateBinaryExpression(expr)
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

	rhs := c.evaluateExpression(expr.Expr)
	var err error
	// TODO: Operand Standards
	switch op {
	case token.NOT:
		err = c.validate(rhs, c.resolveLiteral(BOOLEAN))

		if err == nil {
			return c.resolveLiteral(BOOLEAN)
		}

		// NOT Operand Standard

	case token.SUB:
		err := c.validate(rhs, c.resolveLiteral(INTEGER))
		if err == nil {
			return c.resolveLiteral(INTEGER)
		}

		err = c.validate(rhs, c.resolveLiteral(FLOAT))

		if err == nil {
			return c.resolveLiteral(FLOAT)
		}

	default:
		err = fmt.Errorf("unsupported unary operand `%s`", token.LookUp(op))
	}

	if err != nil {
		panic("there should be an error here")
	}

	c.addError(err.Error(), expr.Range())

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

func (c *Checker) evaluateBinaryExpression(e *ast.BinaryExpression) *SymbolInfo {

	lhs := c.evaluateExpression(e.Left)
	rhs := c.evaluateExpression(e.Right)
	op := e.Op

	err := c.validate(lhs, rhs)

	if err != nil {
		c.addError(err.Error(), e.Range())
		return unresolved
	}

	// TODO: Operator Standards
	switch op {
	case token.ADD:
		// Integers, Floats, Operator Standards
		err = c.validate(lhs, c.resolveLiteral(INTEGER))
		if err == nil {
			return c.resolveLiteral(INTEGER)
		}

		err = c.validate(lhs, c.resolveLiteral(FLOAT))

		if err == nil {
			return c.resolveLiteral(FLOAT)
		}
	case token.SUB:
		// Integers, Floats, Operator Standards
		err = c.validate(lhs, c.resolveLiteral(INTEGER))
		if err == nil {
			return c.resolveLiteral(INTEGER)
		}

		err = c.validate(lhs, c.resolveLiteral(FLOAT))

		if err == nil {
			return c.resolveLiteral(FLOAT)
		}
	case token.QUO, token.MUL:
		// Integers, Floats
		err = c.validate(lhs, c.resolveLiteral(INTEGER))
		if err == nil {
			return c.resolveLiteral(INTEGER)
		}

		err = c.validate(lhs, c.resolveLiteral(FLOAT))

		if err == nil {
			return c.resolveLiteral(FLOAT)
		}

	case token.LSS, token.GTR, token.LEQ, token.GEQ:
		// Integers, Floats, Operator Standards
		err = c.validate(lhs, c.resolveLiteral(INTEGER))
		if err == nil {
			return c.resolveLiteral(INTEGER)
		}

		err = c.validate(lhs, c.resolveLiteral(FLOAT))

		if err == nil {
			return c.resolveLiteral(FLOAT)
		}
	case token.EQL, token.NEQ:
		// Integers, Floats, Booleans
		err = c.validate(lhs, c.resolveLiteral(INTEGER))
		if err == nil {
			return c.resolveLiteral(INTEGER)
		}

		err = c.validate(lhs, c.resolveLiteral(FLOAT))

		if err == nil {
			return c.resolveLiteral(FLOAT)
		}

		err = c.validate(lhs, c.resolveLiteral(BOOLEAN))

		if err == nil {
			return c.resolveLiteral(BOOLEAN)
		}
	default:
		err = fmt.Errorf("unsupported binary operand `%s`", token.LookUp(op))

	}

	if err != nil {
		panic("there should be an error here")
	}

	c.addError(err.Error(), e.Range())
	return unresolved
}
