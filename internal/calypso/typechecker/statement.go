package t

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) checkStatement(stmt ast.Statement) {
	fmt.Printf(
		"Checking Statement: %T @ Line %d\n",
		stmt,
		stmt.Range().Start.Line,
	)
	switch stmt := stmt.(type) {
	case *ast.ExpressionStatement:
		c.checkExpression(stmt.Expr)
	case *ast.VariableStatement:
		c.checkVariableStatement(stmt)
	case *ast.BlockStatement:
		panic("CALL `checkBlockStatement` DIRECTLY")
	case *ast.ReturnStatement:
		panic("CALL `checkReturnStatement` DIRECTLY")
	// case *ast.AliasStatement:
	// case *ast.StructStatement:
	// case *ast.IfStatement:
	default:
		msg := fmt.Sprintf("statement check not implemented, %T\n", stmt)
		panic(msg)
	}
}

func (c *Checker) checkBlockStatement(blk *ast.BlockStatement, fn *types.Function) {
	if len(blk.Statements) == 0 {
		return
	}

	for _, stmt := range blk.Statements {
		switch stmt := stmt.(type) {
		case *ast.ReturnStatement:
			c.checkReturnStatement(stmt, fn.Type().(*types.FunctionSignature))
		default:
			c.checkStatement(stmt)
		}
	}
}

func (c *Checker) checkVariableStatement(stmt *ast.VariableStatement) {

	def := types.NewVar(stmt.Identifier.Value, unresolved)
	def.Mutable = !stmt.IsConstant
	ok := c.define(def)

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is already defined", def.Name()),
			stmt.Identifier.Range(),
		)
		return
	}

	var annotation types.Type

	// Check Annotation
	if t := stmt.Identifier.AnnotatedType; t != nil {
		annotation = c.evaluateTypeExpression(t)
	}

	initializer := c.evaluateExpression(stmt.Value)

	if annotation == nil {
		def.SetType(initializer)
		return
	}

	// Annotation Present, Ensure Annotated Type Matches the provided Type
	annotation, err := c.validate(annotation, initializer)

	if err != nil {
		c.addError(
			err.Error(),
			stmt.Value.Range(),
		)
		return
	}

	def.SetType(annotation)
}

func (c *Checker) checkReturnStatement(stmt *ast.ReturnStatement, fn *types.FunctionSignature) {

	provided := c.evaluateExpression(stmt.Value)

	// no return type set, infer
	if fn.ReturnType == nil {
		fn.ReturnType = provided
		return
	}

	// return type is already set, validate
	t, err := c.validate(fn.ReturnType, provided)

	if err != nil {
		if err != nil {
			c.addError(err.Error(), stmt.Range())
		}

		return
	}

	fn.ReturnType = t
}
