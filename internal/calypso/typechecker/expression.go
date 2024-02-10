package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/symbols"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (c *Checker) checkExpression(expr ast.Expression) {

	fmt.Printf(
		"\nChecking Expression: %T @ Line %d\n",
		expr,
		expr.Range().Start.Line,
	)
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

	sym := symbols.NewSymbol(expr.Identifier.Value, symbols.FunctionSymbol)
	sym.FuncDesc = &symbols.FunctionDescriptor{}
	ok := c.define(sym)

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is already defined", sym.Name),
			expr.Identifier.Range(),
		)
		return
	}

	c.enterScope()

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
		pSym := symbols.NewSymbol(param.Value, symbols.VariableSymbol)
		t := c.evaluateTypeExpression(param.AnnotatedType)
		pSym.TypeDesc = t
		c.define(pSym)
		sym.FuncDesc.Parameters = append(sym.FuncDesc.Parameters, pSym)
	}

	// Body
	c.checkStatement(expr.Body)
	c.currentSym = prev
	c.leaveScope(false)

	if sym.FuncDesc.AnnotatedReturnType != nil {
		sym.FuncDesc.ValidatedReturnType = sym.FuncDesc.AnnotatedReturnType
		fmt.Println()
	} else if sym.FuncDesc.InferredReturnType != nil {
		sym.FuncDesc.ValidatedReturnType = sym.FuncDesc.InferredReturnType
	} else {
		sym.FuncDesc.ValidatedReturnType = c.resolveLiteral(symbols.VOID)
	}

}

func (c *Checker) checkAssignmentExpression(expr *ast.AssignmentExpression) {
	// TODO: Check Mutability
	c.evaluateAssignmentExpression(expr)
}

func (c *Checker) evaluateExpression(expr ast.Expression) *symbols.SymbolInfo {
	c.currentNode = expr

	switch expr := expr.(type) {
	// Literals
	case *ast.IntegerLiteral:
		return c.resolveLiteral(symbols.INTEGER)
	case *ast.BooleanLiteral:
		return c.resolveLiteral(symbols.BOOLEAN)
	case *ast.FloatLiteral:
		return c.resolveLiteral(symbols.FLOAT)
	case *ast.StringLiteral:
		return c.resolveLiteral(symbols.STRING)
	case *ast.NullLiteral:
		return c.resolveLiteral(symbols.NULL)
	case *ast.VoidLiteral:
		return c.resolveLiteral(symbols.VOID)
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
	case *ast.CompositeLiteral:
		return c.evaluateCompositeLiteral(expr)
	default:
		msg := fmt.Sprintf("expression evaluation not implemented, %T", expr)
		panic(msg)
	}
}

func (c *Checker) evaluateIdentifierExpression(expr *ast.IdentifierExpression) *symbols.SymbolInfo {

	s, ok := c.find(expr.Value)

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is not defined", expr.Value),
			expr.Range(),
		)

		return unresolved
	}

	switch s.Type {
	case symbols.VariableSymbol:
		return s.TypeDesc
	default:
		return s
	}
}

func (c *Checker) evaluateUnaryExpression(expr *ast.UnaryExpression) *symbols.SymbolInfo {
	op := expr.Op

	rhs := c.evaluateExpression(expr.Expr)
	var err error
	// TODO: Operand Standards
	switch op {
	case token.NOT:
		err = c.validate(rhs, c.resolveLiteral(symbols.BOOLEAN), nil)

		if err == nil {
			return c.resolveLiteral(symbols.BOOLEAN)
		}

		// NOT Operand Standard

	case token.SUB:
		err := c.validate(rhs, c.resolveLiteral(symbols.INTEGER), nil)
		if err == nil {
			return c.resolveLiteral(symbols.INTEGER)
		}

		err = c.validate(rhs, c.resolveLiteral(symbols.FLOAT), nil)

		if err == nil {
			return c.resolveLiteral(symbols.FLOAT)
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

func (c *Checker) evaluateGroupedExpression(expr *ast.GroupedExpression) *symbols.SymbolInfo {
	return c.evaluateExpression(expr.Expr)
}

func (c *Checker) evaluateArrayLiteral(expr *ast.ArrayLiteral) *symbols.SymbolInfo {
	conc := c.resolveLiteral(symbols.ARRAY)
	symbol := symbols.NewSymbol(conc.Name, symbols.TypeSymbol)
	elementType := c.evaluateExpressionList(expr.Elements)
	symbol.SpecializedOf = conc
	// Specialize Array Generic With Element Type
	c.specialize(symbol.Specializations, conc.GenericParams[0], elementType)
	return symbol
}

func (c *Checker) evaluateExpressionList(exprs []ast.Expression) *symbols.SymbolInfo {

	if len(exprs) == 0 {
		// No Elements, Array Can Contain Any Element
		return c.resolveLiteral(symbols.ANY)
	}

	var expected *symbols.SymbolInfo

	for _, expr := range exprs {
		if expected == nil {
			expected = c.evaluateExpression(expr)
			continue
		}

		provided := c.evaluateExpression(expr)

		err := c.validate(expected, provided, nil)

		// If Unable to validate type, simple set list type as any
		if err != nil {
			return c.resolveLiteral(symbols.ANY)
		}

	}

	return expected

}

func (c *Checker) evaluateBinaryExpression(e *ast.BinaryExpression) *symbols.SymbolInfo {

	lhs := c.evaluateExpression(e.Left)
	rhs := c.evaluateExpression(e.Right)
	op := e.Op

	if lhs.Type == symbols.GenericTypeSymbol || rhs.Type == symbols.GenericTypeSymbol {
		c.addError(
			"unable to perform binary operation on generic types",
			e.Range(),
		)
		return unresolved
	}

	err := c.validate(lhs, rhs, nil)

	if err != nil {
		c.addError(err.Error(), e.Range())
		return unresolved
	}

	// TODO: Operator Standards
	switch op {
	case token.ADD:
		// Integers, Floats, Operator Standards
		err = c.validate(lhs, c.resolveLiteral(symbols.INTEGER), nil)
		if err == nil {
			return c.resolveLiteral(symbols.INTEGER)
		}

		err = c.validate(lhs, c.resolveLiteral(symbols.FLOAT), nil)

		if err == nil {
			return c.resolveLiteral(symbols.FLOAT)
		}
	case token.SUB:
		// Integers, Floats, Operator Standards
		err = c.validate(lhs, c.resolveLiteral(symbols.INTEGER), nil)
		if err == nil {
			return c.resolveLiteral(symbols.INTEGER)
		}

		err = c.validate(lhs, c.resolveLiteral(symbols.FLOAT), nil)

		if err == nil {
			return c.resolveLiteral(symbols.FLOAT)
		}
	case token.QUO, token.MUL:
		// Integers, Floats
		err = c.validate(lhs, c.resolveLiteral(symbols.INTEGER), nil)
		if err == nil {
			return c.resolveLiteral(symbols.INTEGER)
		}

		err = c.validate(lhs, c.resolveLiteral(symbols.FLOAT), nil)

		if err == nil {
			return c.resolveLiteral(symbols.FLOAT)
		}

	case token.LSS, token.GTR, token.LEQ, token.GEQ:
		// Integers, Floats, Operator Standards
		err = c.validate(lhs, c.resolveLiteral(symbols.INTEGER), nil)
		if err == nil {
			return c.resolveLiteral(symbols.INTEGER)
		}

		err = c.validate(lhs, c.resolveLiteral(symbols.FLOAT), nil)

		if err == nil {
			return c.resolveLiteral(symbols.FLOAT)
		}
	case token.EQL, token.NEQ:
		// Integers, Floats, Booleans
		err = c.validate(lhs, c.resolveLiteral(symbols.INTEGER), nil)
		if err == nil {
			return c.resolveLiteral(symbols.INTEGER)
		}

		err = c.validate(lhs, c.resolveLiteral(symbols.FLOAT), nil)

		if err == nil {
			return c.resolveLiteral(symbols.FLOAT)
		}

		err = c.validate(lhs, c.resolveLiteral(symbols.BOOLEAN), nil)

		if err == nil {
			return c.resolveLiteral(symbols.BOOLEAN)
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

func (c *Checker) evaluateAssignmentExpression(expr *ast.AssignmentExpression) *symbols.SymbolInfo {

	lhs := c.evaluateExpression(expr.Target)
	rhs := c.evaluateExpression(expr.Value)

	err := c.validate(lhs, rhs, nil)

	if err != nil {
		c.addError(err.Error(), expr.Range())
	}

	// Assignment Calls are void
	return c.resolveLiteral(symbols.VOID)
}

func (c *Checker) evaluateCallExpression(expr *ast.CallExpression) *symbols.SymbolInfo {

	target := c.evaluateExpression(expr.Target)

	// Ensure Target is Callable
	if target.FuncDesc == nil || target.Type != symbols.FunctionSymbol {
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

	sym := symbols.NewSymbol(target.Name, symbols.FunctionSymbol)
	sym.SpecializedOf = target

	for i, arg := range expr.Arguments {
		provided := c.evaluateExpression(arg)
		expected := target.FuncDesc.Parameters[i].TypeDesc

		if expected == nil {
			panic("[CallExpression] Parameter Type Should not be nil")
		}

		if expected.Type == symbols.GenericTypeSymbol {
			// Generic, First Find Specialization
			v, ok := sym.Specializations.Get(expected)

			// If not found add
			if !ok {
				// Expected is a generic type, specialize
				err := c.specialize(sym.Specializations, expected, provided)
				if err != nil {
					c.addError(
						err.Error(),
						arg.Range(),
					)
					continue
				}
			} else {
				// specialization found, validate
				err := c.validate(v, provided, nil)

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
			err := c.validate(expected, provided, nil)

			if err != nil {
				c.addError(
					err.Error(),
					arg.Range(),
				)
			}

		}
	}

	if target.FuncDesc.ValidatedReturnType.Type == symbols.GenericTypeSymbol {
		v, ok := sym.Specializations.Get(target.FuncDesc.ValidatedReturnType)

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

func (c *Checker) evaluateCompositeLiteral(lit *ast.CompositeLiteral) *symbols.SymbolInfo {

	base, ok := c.find(lit.Identifier.Value)

	if base.Type != symbols.StructSymbol {
		c.addError(
			fmt.Sprintf("`%s` is not a struct", lit.Identifier.Value),
			lit.Identifier.Range(),
		)
		return unresolved
	}

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is not defined", lit.Identifier.Value),
			lit.Range(),
		)

		return unresolved
	}

	sym := symbols.NewSymbol(base.Name, symbols.StructSymbol)
	sym.SpecializedOf = base

	seen := make(map[string]bool)

	for _, pair := range lit.Pairs {
		key := pair.Key.Value

		expectedProperty, ok := base.Properties[key]

		// Property is not defined in struct
		if !ok {
			c.addError(
				fmt.Sprintf("`%s` is not a valid property", key),
				pair.Range(),
			)
			continue
		}

		expected := expectedProperty.TypeDesc

		// Ensure there is a provided type description
		if expected == nil {
			panic("[CompositeLiteral] Property Type Should not be nil")
		}

		_, ok = seen[key]

		// Property has already been evaluated
		if ok {
			c.addError(
				fmt.Sprintf("`%s` has already been provided", key),
				pair.Range(),
			)
			continue
		}

		provided := c.evaluateExpression(pair.Value)

		var err error
		if expected.Type == symbols.GenericTypeSymbol {
			_, ok := sym.Specializations[expected]
			if !ok {
				fmt.Println("[DEBUG] Struct Specialize")
				err = c.specialize(sym.Specializations, expected, provided)
			}
		} else {
			err = c.validate(expected, provided, sym.Specializations)
		}

		if err != nil {
			c.addError(
				err.Error(),
				pair.Range(),
			)
			continue
		}
		// Mark As seen
		seen[key] = true
	}

	// No Generic Arguments
	if len(base.GenericParams) == 0 {
		return base
	} else {
		// Ensure All Generic Params Have been Specialized
		sym.Specializations.Debug()
		fmt.Println(base.GenericParams)
		for _, param := range base.GenericParams {
			_, ok := sym.Specializations[param]

			if !ok {
				c.addError(
					fmt.Sprintf("unable to infer type of `%s`", param),
					lit.Range(),
				)
			}
		}

		return sym
	}
}
