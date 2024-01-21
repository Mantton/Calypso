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

	default:
		msg := fmt.Sprintf("expression evaluation not implemented, %T", expr)
		panic(msg)
	}
}
