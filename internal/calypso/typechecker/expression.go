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
	case *ast.AssignmentExpression:
		c.checkAssignmentExpression(expr)
	default:
		msg := fmt.Sprintf("expression check not implemented, %T", expr)
		panic(msg)
	}
}

func (c *Checker) checkFunctionExpression(expr *ast.FunctionExpression) {
	c.enterScope()

	sym := newSymbolInfo(expr.Identifier.Value, TypeSymbol)
	sym.FuncDesc = &FunctionDescriptor{}

	prev := c.currentSym
	c.currentSym = sym

	// Evaluate Generic Parameters
	if expr.GenericParams != nil {
		c.evaluateGenericParameters(sym, expr.GenericParams)
	}

	// Evaluate Annotated Return Type
	if expr.ReturnType != nil {
		sym.FuncDesc.AnnotatedReturnType = c.evaluateTypeExpression(expr.ReturnType)
	}

	// Params
	for _, param := range expr.Params {
		pSym := newSymbolInfo(param.Value, VariableSymbol)
		t := c.evaluateTypeExpression(param.AnnotatedType)
		pSym.TypeDesc = t
		c.define(pSym)
		sym.FuncDesc.Parameters = append(sym.FuncDesc.Parameters, t)
	}

	// Body
	c.checkStatement(expr.Body)
	c.currentSym = prev
	c.leaveScope(false)

	if sym.FuncDesc.AnnotatedReturnType != nil {
		sym.FuncDesc.ValidatedReturnType = sym.FuncDesc.AnnotatedReturnType
	} else if sym.FuncDesc.InferredReturnType != nil {
		sym.FuncDesc.ValidatedReturnType = sym.FuncDesc.InferredReturnType
	} else {
		sym.FuncDesc.ValidatedReturnType = c.resolveLiteral(VOID)
	}

	fn := newSymbolInfo(sym.Name, FunctionSymbol)
	fn.TypeDesc = sym
	c.define(fn)
}

func (c *Checker) checkAssignmentExpression(expr *ast.AssignmentExpression) {
	// TODO: Check Mutability
	c.evaluateAssignmentExpression(expr)
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
	case *ast.AssignmentExpression:
		return c.evaluateAssignmentExpression(expr)
	case *ast.CallExpression:
		return c.evaluateCallExpression(expr)
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

func (c *Checker) evaluateAssignmentExpression(expr *ast.AssignmentExpression) *SymbolInfo {

	lhs := c.evaluateExpression(expr.Target)
	rhs := c.evaluateExpression(expr.Value)

	err := c.validate(lhs, rhs)

	if err != nil {
		c.addError(err.Error(), expr.Range())
	}

	// Assignment Calls are void
	return c.resolveLiteral(VOID)
}

func (c *Checker) evaluateCallExpression(expr *ast.CallExpression) *SymbolInfo {

	target := c.evaluateExpression(expr.Target)

	if target.FuncDesc == nil {
		c.addError(
			fmt.Sprintf("`%s` is not a function", target.Name),
			expr.Target.Range(),
		)
		return unresolved
	}

	// Check Argument Count Matches function parameter count
	if len(expr.Arguments) != len(target.FuncDesc.Parameters) {
		c.addError(
			fmt.Sprintf("expected %d arguments, provided %d",
				len(target.FuncDesc.Parameters),
				len(expr.Arguments)),
			expr.Range(),
		)
		return target.FuncDesc.ValidatedReturnType

	}

	// Specialize Generics
	if len(target.GenericParams) != 0 {
		fmt.Println("Resolve Generic Parameters")
	}
	t := make(specializationTable)

	for i, arg := range expr.Arguments {
		provided := c.evaluateExpression(arg)
		expected := target.FuncDesc.Parameters[i]

		if expected.Type == GenericTypeSymbol {
			// Generic, First Find Specialization
			v, ok := t.get(expected)

			// If not found add
			if !ok {
				// Expected is a generic type, specialize
				err := c.add(t, expected, provided)

				if err != nil {
					c.addError(
						err.Error(),
						arg.Range(),
					)
					continue
				}
			} else {
				// specialization found, validate
				err := c.validate(v, provided)

				if err != nil {
					c.addError(
						err.Error(),
						arg.Range(),
					)
					continue
				}
			}

		} else {
			// Expected is not a generic, ensure strict conformance
			err := c.validate(expected, provided)

			if err != nil {
				c.addError(
					err.Error(),
					arg.Range(),
				)
			}

		}
	}

	if target.FuncDesc.ValidatedReturnType.Type == GenericTypeSymbol {
		v, ok := t.get(target.FuncDesc.ValidatedReturnType)

		if !ok {
			c.addError(
				fmt.Sprintf("unable to infer return type of generic `%s`", target.FuncDesc.ValidatedReturnType.Name),
				expr.Range(),
			)
		}

		return v

	} else {
		return target.FuncDesc.ValidatedReturnType
	}

}
