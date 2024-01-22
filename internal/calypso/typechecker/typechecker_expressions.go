package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
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
	case *ast.ArrayLiteral:
		return GenerateGenericType("ArrayLiteral", t.evaluateExpressionList(expr.Elements))
	case *ast.MapLiteral:
		k, v := t.evaluateExpressionPairs(expr.Pairs)
		return GenerateGenericType("MapLiteral", k, v)
	default:
		msg := fmt.Sprintf("expression evaluation not implemented, %T", expr)
		panic(msg)
	}
}

func (t *TypeChecker) evaluateExpressionList(exprs []ast.Expression) ExpressionType {

	if len(exprs) == 0 {
		panic("any list")
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
