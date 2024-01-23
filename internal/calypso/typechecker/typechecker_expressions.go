package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (t *TypeChecker) evaluateExpression(expr ast.Expression) ExpressionType {
	switch expr := expr.(type) {

	case *ast.IntegerLiteral:
		return GenerateBaseType("IntegerLiteral")
	case *ast.FloatLiteral:
		return GenerateBaseType("FloatLiteral")
	case *ast.StringLiteral:
		return GenerateBaseType("StringLiteral")
	case *ast.BooleanLiteral:
		return GenerateBaseType("BooleanLiteral")
	case *ast.VoidLiteral:
		return GenerateBaseType("VoidLiteral")
	case *ast.NullLiteral:
		return GenerateBaseType("NullLiteral")
	case *ast.ArrayLiteral:
		return GenerateGenericType("ArrayLiteral", t.evaluateExpressionList(expr.Elements))
	case *ast.MapLiteral:
		k, v := t.evaluateExpressionPairs(expr.Pairs)
		return GenerateGenericType("MapLiteral", k, v)
	case *ast.FunctionExpression:
		return t.get(expr.Identifier)
	case *ast.IdentifierExpression:
		return t.get(expr)
	case *ast.UnaryExpression:
		return t.evaluateUnaryExpression(expr)
	case *ast.AssignmentExpression:
		return t.evaluateAssignmentExpression(expr)
	case *ast.GroupedExpression:
		return t.evaluateGroupedExpression(expr)
	case *ast.BinaryExpression:
		return t.evaluateBinaryExpression(expr)
	case *ast.CallExpression:
		return t.evaluateCallExpression(expr)
	case *ast.IndexExpression:
		return t.evaluateIndexExpression(expr)
	default:
		msg := fmt.Sprintf("expression evaluation not implemented, %T", expr)
		panic(msg)
	}
}

func (t *TypeChecker) evaluateUnaryExpression(expr *ast.UnaryExpression) ExpressionType {
	op := expr.Op
	provided := t.evaluateExpression(expr.Expr)
	switch op {
	case token.NOT:
		panic("TODO: NOT Operator")
		// Returns a boolean type for sure.
	case token.SUB:
		expected := GenerateBaseType("IntegerLiteral")
		if !t.validate(provided, expected) {
			msg := fmt.Sprintf("expected `%s`, received `%s`", expected, provided)
			panic(t.error(msg, expr.Expr))
		}

	default:
		panic("Bad Logic Path")
	}

	return provided
}

func (t *TypeChecker) evaluateAssignmentExpression(expr *ast.AssignmentExpression) ExpressionType {

	expected := t.evaluateExpression(expr.Target)
	provided := t.evaluateExpression(expr.Value)

	t.mustValidate(provided, expected, expr.Value)
	return GenerateBaseType("VoidLiteral")
}

func (t *TypeChecker) evaluateGroupedExpression(expr *ast.GroupedExpression) ExpressionType {
	return t.evaluateExpression(expr.Expr)
}

func (t *TypeChecker) evaluateBinaryExpression(expr *ast.BinaryExpression) ExpressionType {

	left := t.evaluateExpression(expr.Left)
	right := t.evaluateExpression(expr.Right)
	t.mustValidate(left, right, expr.Right)

	op := expr.Op
	// TODO: Further checks
	switch op {
	case token.ADD, token.SUB, token.MUL, token.REM:
		return left

	case token.LSS, token.GTR, token.LEQ, token.GEQ, token.EQL, token.NEQ:
		return GenerateBaseType("BooleanLiteral")
	}

	return left
}

func (t *TypeChecker) evaluateCallExpression(expr *ast.CallExpression) ExpressionType {
	target := t.evaluateExpression(expr.Target)

	switch target := target.(type) {
	case *FunctionType:
		// Check Arguments are of correct type
		expectedArgCount := len(target.Params)
		providedArgCount := len(expr.Arguments)

		if expectedArgCount != providedArgCount {
			msg := fmt.Sprintf("expected %d arguments, provided %d instead", expectedArgCount, providedArgCount)
			panic(t.error(msg, expr.Target))
		}

		for idx, arg := range expr.Arguments {
			provided := t.evaluateExpression(arg)
			expected := target.Params[idx]
			t.mustValidate(provided, expected, arg)
		}

		return target.Return

	default:
		panic("TODO: handle call expression")

	}
}

func (t *TypeChecker) evaluateIndexExpression(expr *ast.IndexExpression) ExpressionType {
	panic("TODO: standards")
}

func (t *TypeChecker) evaluateExpressionList(exprs []ast.Expression) ExpressionType {

	if len(exprs) == 0 {
		return GenerateBaseType("AnyLiteral")
	}

	var base ExpressionType

	for _, expr := range exprs {

		if base == nil {
			base = t.evaluateExpression(expr)
			continue
		}

		parsed := t.evaluateExpression(expr)

		if !t.validate(parsed, base) {
			return GenerateBaseType("AnyLiteral")
		}

	}

	return base
}

func (t *TypeChecker) evaluateExpressionPairs(pairs map[ast.Expression]ast.Expression) (ExpressionType, ExpressionType) {

	if pairs == nil {
		return GenerateBaseType("AnyLiteral"), GenerateBaseType("AnyLiteral")
	}
	var keyType ExpressionType
	var valueType ExpressionType

	for k, v := range pairs {

		if keyType == nil && valueType == nil {
			keyType = t.evaluateExpression(k)
			valueType = t.evaluateExpression(v)
			continue
		}

		pK := t.evaluateExpression(k)
		pV := t.evaluateExpression(v)

		if !t.validate(pK, keyType) {
			keyType = GenerateBaseType("AnyLiteral")
		}

		if !t.validate(pV, valueType) {
			valueType = GenerateBaseType("AnyLiteral")
		}
	}

	return keyType, valueType
}

func (t *TypeChecker) checkExpression(expr ast.Expression) {
	switch expr := expr.(type) {
	case *ast.FunctionExpression:
		t.checkFunctionExpression(expr)

	default:
		msg := fmt.Sprintf("expression check not implemented, %T", expr)
		panic(msg)
	}
}

func (t *TypeChecker) checkFunctionExpression(expr *ast.FunctionExpression) {
	prevCFS := t.cfs
	params := []ExpressionType{}
	// Enter Function Scope
	t.enterScope()
	t.cfs = &CurrentFunctionState{}

	// Declare & Define Parameters
	for _, param := range expr.Params {

		paramType := t.evaluateTypeExpression(param.AnnotatedType)
		t.register(param.Value, paramType)
		params = append(params, paramType)
	}

	// TypeCheck Return Type If Provided

	var retType ExpressionType

	if expr.ReturnType != nil {
		retType = t.evaluateTypeExpression(expr.ReturnType)
		t.cfs.AnnotatedReturnType = retType
	}

	// TypeCheck Function body
	t.checkStatement(expr.Body)

	// Check Function Return Type
	inferred := t.cfs.InferredReturnType
	annotated := t.cfs.AnnotatedReturnType

	if inferred == nil && annotated == nil {
		retType = GenerateBaseType("VoidLiteral")
	} else if inferred == nil && annotated != nil {
		// Function Annotated To return Void
		if t.validate(annotated, GenerateBaseType("VoidLiteral")) {
			retType = GenerateBaseType("VoidLiteral")
		} else {
			msg := fmt.Sprintf("function does not return type `%s`", annotated)
			panic(t.error(msg, expr.Identifier))
		}

	} else if inferred != nil && annotated == nil {
		retType = inferred
	} else if inferred != nil && annotated != nil {
		t.mustValidate(inferred, annotated, expr.Identifier)
	}

	// Resolution Complete, define &  leave function scope
	fnType := GenerateFunctionType(expr.Identifier.Value, retType, params...)
	t.leaveScope()
	t.define(expr.Identifier, fnType)
	t.cfs = prevCFS
}
