package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
)

func (t *TypeChecker) checkStatement(stmt ast.Statement) {
	switch stmt := stmt.(type) {
	case *ast.VariableStatement:
		t.checkVariableStatement(stmt)
	default:
		msg := fmt.Sprintf("statement check not implemented, %T", stmt)
		panic(msg)
	}
}

func (t *TypeChecker) checkVariableStatement(stmt *ast.VariableStatement) {
	// TODO: Handle situation where the type fails to resolve, simply set it to an UnknownExpr Type
	var annotation ExpressionType

	if stmt.TypeAnnotation != nil {
		annotation = t.evaluateTypeExpression(stmt.TypeAnnotation)
	}

	initializer := t.evaluateExpression(stmt.Value)

	// Infer Variable Type From Initializer if no annotation is provided
	if stmt.TypeAnnotation == nil {
		// Define In Scope
		t.define(stmt.Identifier, initializer)
		return
	}

	// Define In Scope
	t.mustValidate(initializer, annotation)
	t.define(stmt.Identifier, annotation)
}
