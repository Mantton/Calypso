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
	case *ast.EnumStatement:
		c.checkEnumStatement(stmt)
	case *ast.SwitchStatement:
		c.checkSwitchStatement(stmt)
	case *ast.BreakStatement:
		return // nothing to TC on break
	case *ast.WhileStatement:
		c.checkWhileStatement(stmt)
	case *ast.TypeStatement:
		c.checkTypeStatement(stmt, nil)
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
	def := types.NewDefinedType(n.Identifier.Value, unresolved, nil, c.scope)
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

			def.AddTypeParameter(types.AsTypeParam(t))
			ok := def.Scope.Define(types.AsTypeParam(t))
			if !ok {
				c.addError(fmt.Sprintf("%s is already defined.", t), p.Identifier.Range())
				return
			}

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

func (c *Checker) checkEnumStatement(n *ast.EnumStatement) {
	name := n.Identifier.Value

	// 1 - Define
	def := types.NewDefinedType(name, unresolved, nil, c.scope)
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

			tP := t.(*types.TypeParam)
			def.AddTypeParameter(tP)
			ok := def.Scope.Define(tP)

			if !ok {
				c.addError(fmt.Sprintf("%s is already defined.", tP), p.Identifier.Range())
				return
			}
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

func (c *Checker) checkSwitchStatement(n *ast.SwitchStatement) {

	// 1 - Condition
	condition := c.evaluateExpression(n.Condition)

	// 2 - Cases

	seenDefault := false

	if len(n.Cases) == 0 {
		c.addError("expected at least one case", n.Range())
		return
	}

	for _, cs := range n.Cases {

		// Scope
		c.enterScope()

		defer func() {
			if len(c.scope.Symbols) != 0 {
				c.table.AddScope(cs, c.scope)
			}
			c.leaveScope()
		}()

		// Default Case
		if cs.IsDefault {
			// 1 - Check default has already been seen
			if seenDefault {
				c.addError("default case already added", cs.Range())
				continue
			}

			// 2 - Block
			c.checkBlockStatement(cs.Action)
			seenDefault = true
			continue
		}

		// 1 - Condition
		// For Tuple types, provide lhsType, which provides the generic specializations & correct fn signature when required
		c.lhsType = condition
		caseCondition := c.evaluateExpression(cs.Condition)
		c.lhsType = nil

		_, err := c.validate(condition, caseCondition)

		if err != nil {
			c.addError(err.Error(), cs.Condition.Range())
			continue
		}

		// 2 - Block
		c.checkBlockStatement(cs.Action)
	}
}

func (c *Checker) checkWhileStatement(n *ast.WhileStatement) {
	condition := c.evaluateExpression(n.Condition)

	_, err := c.validate(types.LookUp(types.Bool), condition)

	if err != nil {
		c.addError(err.Error(), n.Condition.Range())
		return
	}

	c.enterScope()
	c.table.AddScope(n, c.scope)
	defer c.leaveScope()
	c.checkBlockStatement(n.Action)
}

func (c *Checker) checkTypeStatement(n *ast.TypeStatement, standard *types.Standard) {
	name := n.Identifier.Value

	// 1 - Define
	alias := types.NewAlias(name, types.LookUp(types.Placeholder))
	ok := c.define(alias)
	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is already defined in context", name),
			n.Identifier.Range(),
		)
		return
	}

	// 2 - Add to Standard
	if standard != nil {
		standard.AddType(alias)
	}

	// 3 - Type Parameters

	hasError := false
	if n.GenericParams != nil {
		for _, p := range n.GenericParams.Parameters {
			t := c.evaluateGenericParameterExpression(p)
			if t == unresolved {
				continue
			}

			tP := t.(*types.TypeParam)
			alias.AddTypeParameter(tP)

			if !ok {
				c.addError(fmt.Sprintf("%s is already defined.", tP), p.Identifier.Range())
				hasError = true
			}
		}
	}

	if hasError {
		return
	}

	// 4 - RHS Type
	var RHS types.Type

	if n.Value != nil {
		fmt.Println("[DEBUG] Resolving RHS for Alias", n.Identifier.Value)
		RHS = c.evaluateTypeExpression(n.Value, alias.TypeParameters)
		alias.SetType(RHS)
	}
}
