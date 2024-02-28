package typechecker

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
		c.checkReturnStatement(stmt)
	case *ast.IfStatement:
		c.checkIfStatement(stmt)
	// case *ast.AliasStatement:
	// case *ast.StructStatement:
	default:
		msg := fmt.Sprintf("statement check not implemented, %T\n", stmt)
		panic(msg)
	}
}

func (c *Checker) checkBlockStatement(blk *ast.BlockStatement) {
	if len(blk.Statements) == 0 {
		return
	}

	for _, stmt := range blk.Statements {
		c.checkStatement(stmt)
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
		def.SetType(annotation)
	}

	initializer := c.evaluateExpression(stmt.Value)

	err := c.validateAssignment(def, initializer, stmt.Value)
	if err != nil {
		c.addError(
			err.Error(),
			stmt.Value.Range(),
		)
		return
	}
}

func (c *Checker) checkReturnStatement(stmt *ast.ReturnStatement) {

	if c.fn == nil {
		c.addError(
			"top level return is not allowed",
			stmt.Value.Range(),
		)
		return
	}

	fn := c.fn
	provided := c.evaluateExpression(stmt.Value)

	// return type is already set, validate
	err := c.validateAssignment(fn.Result, provided, stmt.Value)

	if err != nil {
		if err != nil {
			c.addError(err.Error(), stmt.Range())
		}

		return
	}
}

func (c *Checker) checkIfStatement(stmt *ast.IfStatement) {

	// 1 - Check Condition
	cond := c.evaluateExpression(stmt.Condition)
	_, err := c.validate(types.LookUp(types.Bool), cond)

	if err != nil {
		c.addError(err.Error(), stmt.Condition.Range())
		return
	}

	// 2 - Check Action
	c.enterScope()
	c.checkBlockStatement(stmt.Action)
	c.leaveScope()

	// 3 - Check Alternative
	if stmt.Alternative != nil {
		c.enterScope()
		c.checkBlockStatement(stmt.Alternative)
		c.leaveScope()
	}
}
