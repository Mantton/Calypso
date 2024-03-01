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
	case *ast.StructStatement:
		c.checkStructStatement(stmt)

	// case *ast.AliasStatement:
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
		annotation = c.evaluateTypeExpression(t, nil)
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
		c.addError(err.Error(), stmt.Range())

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

func (c *Checker) checkStructStatement(n *ast.StructStatement) {

	// 1 - Define
	def := types.NewDefinedType(n.Identifier.Value, unresolved, nil)
	ok := c.define(def)

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is already defined", def.Name()),
			n.Identifier.Range(),
		)
		return
	}

	c.enterScope()
	defer c.leaveScope()

	// 2 - TODO: Parse Generic Params
	if n.GenericParams != nil {
		for _, p := range n.GenericParams.Parameters {
			// TODO: Standards
			d := types.NewTypeParam(p.Identifier.Value, nil)

			for _, eI := range p.Standards {
				sym, ok := c.find(eI.Value)

				if !ok {
					c.addError(
						fmt.Sprintf("`%s` cannot be found", eI.Value),
						p.Identifier.Range(),
					)
					return
				}

				s, ok := sym.Type().Parent().(*types.Standard)

				if !ok {
					if !ok {
						c.addError(
							fmt.Sprintf("`%s` is not a standard", eI.Value),
							p.Identifier.Range(),
						)
						return
					}
				}

				d.AddConstraint(s)

			}
			def.AddTypeParameter(d)
		}

	}

	// 3 - Parse Fields
	var fields []*types.Var

	for _, f := range n.Fields {
		d := types.NewVar(f.Value, unresolved)
		t := c.evaluateTypeExpression(f.AnnotatedType, def.TypeParameters)
		d.SetType(t)
		fields = append(fields, d)
	}

	typ := types.NewStruct(fields)
	def.SetType(typ)
}
