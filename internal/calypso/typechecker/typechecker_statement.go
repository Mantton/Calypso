package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
)

func (t *TypeChecker) checkStatement(stmt ast.Statement) {
	switch stmt := stmt.(type) {
	case *ast.VariableStatement:
		t.checkVariableStatement(stmt)
	case *ast.BlockStatement:
		t.checkBlockStatement(stmt)
	case *ast.ReturnStatement:
		t.checkReturnStatement(stmt)

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
	t.mustValidate(initializer, annotation, stmt.Identifier)
	t.define(stmt.Identifier, annotation)
}

func (t *TypeChecker) checkBlockStatement(blk *ast.BlockStatement) {

	if len(blk.Statements) == 0 {
		return
	}

	for _, stmt := range blk.Statements {
		t.checkStatement(stmt)
	}
}

func (t *TypeChecker) checkReturnStatement(stmt *ast.ReturnStatement) {
	if t.cfs == nil {
		panic(t.error("cannot use return statement outside function scope.", stmt))
	}

	retType := t.evaluateExpression(stmt.Value)

	// No Annotated Return Type, Infer Instead
	if t.cfs.AnnotatedReturnType == nil {

		// Currently No Inferred Type, Set Current
		if t.cfs.InferredReturnType == nil {
			t.cfs.InferredReturnType = retType
		} else {
			// Inferred Return Type, Validate or Mark As Any
			if !t.validate(retType, t.cfs.InferredReturnType) {
				t.cfs.InferredReturnType = GenerateBaseType("AnyLiteral")
			}
		}

		return
	}

	// Annotated Type Exists, Validate & Type Inferred
	t.mustValidate(retType, t.cfs.AnnotatedReturnType, stmt)
	if t.cfs.InferredReturnType == nil {
		t.cfs.InferredReturnType = retType
	}

}
