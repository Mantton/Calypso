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
	case *ast.AliasStatement:
		c.checkAliasStatement(stmt)
	case *ast.EnumStatement:
		c.checkEnumStatement(stmt)
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

	// 2  Parse Generic Params
	if n.GenericParams != nil {
		for _, p := range n.GenericParams.Parameters {
			t := c.evaluateGenericParameterExpression(p)
			if t == unresolved {
				continue
			}

			def.AddTypeParameter(t.(*types.TypeParam))
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

func (c *Checker) checkAliasStatement(n *ast.AliasStatement) {
	name := n.Identifier.Value

	// 1 - Define
	def := types.NewDefinedType(name, unresolved, nil)
	ok := c.define(def)

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is already defined", name),
			n.Identifier.Range(),
		)
		return
	}

	// Get Type Params
	hasError := false
	if n.GenericParams != nil {
		for _, p := range n.GenericParams.Parameters {
			t := c.evaluateGenericParameterExpression(p)
			if t == unresolved {
				hasError = true
				continue
			}

			def.AddTypeParameter(t.(*types.TypeParam))
		}
	}
	if hasError {
		return
	}

	// Target
	RHS := c.evaluateTypeExpression(n.Target, def.TypeParameters)

	if RHS == unresolved {
		// already reported
		return
	}

	a := types.NewAlias(name, RHS)
	def.SetType(a)
}

func (c *Checker) checkEnumStatement(n *ast.EnumStatement) {
	name := n.Identifier.Value

	// 1 - Define
	def := types.NewDefinedType(name, unresolved, nil)
	ok := c.define(def)

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is already defined", name),
			n.Identifier.Range(),
		)
		return
	}

	// 2  Parse Generic Params
	if n.GenericParams != nil {
		for _, p := range n.GenericParams.Parameters {
			t := c.evaluateGenericParameterExpression(p)
			if t == unresolved {
				continue
			}

			def.AddTypeParameter(t.(*types.TypeParam))
		}

	}

	cases := make(map[string]bool)
	discs := make(map[int]bool)
	gDisc := 0
	variants := []*types.EnumVariant{}
	// 3 Parse Variants
	for _, v := range n.Variants {

		name := v.Identifier.Value
		if _, ok := cases[name]; ok {
			c.addError(fmt.Sprintf("`%s` is already a variant", name), v.Identifier.Range())
			continue
		}

		// define
		cases[name] = ok

		// set fields
		fields := []*types.Var{}

		if v.Fields != nil {
			for _, f := range v.Fields.Fields {
				t := c.evaluateTypeExpression(f, def.TypeParameters)
				fields = append(fields, types.NewVar("", t))
			}

			if len(fields) == 0 {
				c.addError("tuple enum must provide at least 1 parameter", v.Identifier.Range())
				continue
			}
		}

		// set discriminant
		discriminant := gDisc

		if v.Discriminant != nil {
			d, ok := v.Discriminant.Value.(*ast.IntegerLiteral)
			if !ok {
				c.addError("discriminant must be an integer", v.Discriminant.Value.Range())
				continue
			}

			discriminant = int(d.Value)
			if _, ok := discs[discriminant]; ok {
				c.addError(fmt.Sprintf("`%d` is already assigned to a variant", discriminant), v.Identifier.Range())
				continue
			}
			gDisc = discriminant
		}

		v := types.NewEnumVariant(name, discriminant, fields)

		variants = append(variants, v)
		discs[discriminant] = true
		gDisc += 1

	}

	if len(variants) == 0 {
		c.addError("expected at least 1 enum variant", n.Identifier.Range())
	}

	e := types.NewEnum(name, variants)
	def.SetType(e)
}
