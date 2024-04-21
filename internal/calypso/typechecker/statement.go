package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) checkStatement(stmt ast.Statement, ctx *NodeContext) {
	fmt.Printf(
		"Checking Statement: %T @ Line %d\n",
		stmt,
		stmt.Range().Start.Line,
	)
	switch stmt := stmt.(type) {
	case *ast.ExpressionStatement:
		c.checkExpression(stmt.Expr, ctx)
	case *ast.VariableStatement:
		// Global Constant has already been declared/registered
		if ctx.scope == c.ParentScope() {
			return
		}
		c.checkVariableStatement(stmt, ctx, false)
	case *ast.BlockStatement:
		panic("CALL `checkBlockStatement` DIRECTLY")
	case *ast.ReturnStatement:
		c.checkReturnStatement(stmt, ctx)
	case *ast.IfStatement:
		c.checkIfStatement(stmt, ctx)
	case *ast.StructStatement:
		c.checkStructStatement(stmt, ctx)
	case *ast.EnumStatement:
		c.checkEnumStatement(stmt, ctx)
	case *ast.SwitchStatement:
		c.checkSwitchStatement(stmt, ctx)
	case *ast.BreakStatement:
		return // nothing to TC on break
	case *ast.WhileStatement:
		c.checkWhileStatement(stmt, ctx)
	case *ast.TypeStatement:
		c.checkTypeStatement(stmt, ctx)
	default:
		msg := fmt.Sprintf("statement check not implemented, %T\n", stmt)
		panic(msg)
	}
}

func (c *Checker) checkBlockStatement(blk *ast.BlockStatement, ctx *NodeContext) {
	if len(blk.Statements) == 0 {
		return
	}

	for _, stmt := range blk.Statements {
		c.checkStatement(stmt, ctx)
	}
}

func (c *Checker) checkVariableStatement(stmt *ast.VariableStatement, ctx *NodeContext, global bool) {

	var def *types.Var
	if !global {
		def = types.NewVar(stmt.Identifier.Value, unresolved)
		def.Mutable = !stmt.IsConstant
		err := ctx.scope.Define(def)

		if err != nil {
			c.addError(
				fmt.Sprintf(err.Error(), def.Name()),
				stmt.Identifier.Range(),
			)
			return
		}
	} else {
		symbol := c.ParentScope().MustResolve(stmt.Identifier.Value)

		if symbol == nil {
			return
		}

		def = types.AsVar(symbol)
		if def == nil {
			return
		}
	}

	var annotation types.Type

	// Check Annotation
	if t := stmt.Identifier.AnnotatedType; t != nil {
		annotation = c.evaluateTypeExpression(t, nil, ctx)
		def.SetType(annotation)
	}

	initializer := c.evaluateExpression(stmt.Value, ctx)

	err := c.validateAssignment(def, initializer, stmt.Value, false)
	if err != nil {
		c.addError(
			err.Error(),
			stmt.Value.Range(),
		)
		return
	}

	// main context, ensure constant is known at compile time
	if ctx == c.ctx && !types.IsConstant(def.Type().Parent()) {
		c.addError(fmt.Sprintf("global constant \"%s\" must be a known compile-time constant", def.Name()),
			stmt.Range())
	}
}

func (c *Checker) checkReturnStatement(stmt *ast.ReturnStatement, ctx *NodeContext) {

	if ctx.sg == nil {
		c.addError(
			"top level return is not allowed",
			stmt.Value.Range(),
		)
		return
	}

	fn := ctx.sg
	provided := c.evaluateExpression(stmt.Value, ctx)

	// return type is already set, validate
	err := c.validateAssignment(fn.Result, provided, stmt.Value, false)

	if err != nil {
		c.addError(err.Error(), stmt.Range())

		return
	}
}

func (c *Checker) checkIfStatement(stmt *ast.IfStatement, ctx *NodeContext) {
	scope := types.NewScope(ctx.scope, "")
	newCtx := NewContext(scope, ctx.sg, nil)
	// 1 - Check Condition
	cond := c.evaluateExpression(stmt.Condition, newCtx)
	_, err := c.validate(types.LookUp(types.Bool), cond)

	if err != nil {
		c.addError(err.Error(), stmt.Condition.Range())
		return
	}

	// 2 - Check Action
	c.checkBlockStatement(stmt.Action, newCtx)

	// 3 - Check Alternative
	if stmt.Alternative != nil {
		c.checkBlockStatement(stmt.Alternative, newCtx)
	}
}

func (c *Checker) checkStructStatement(n *ast.StructStatement, ctx *NodeContext) {
	def := c.resolve(n.Identifier, n, ctx.scope)

	// already reported
	if def == nil {
		return
	}

	// 2  Parse Generic Params
	if n.GenericParams != nil {
		for i, p := range n.GenericParams.Parameters {
			tP := def.TypeParameters[i]
			c.evaluateTypeParamterStandards(p, tP, ctx)
		}

	}

	// 3 - Parse Fields
	var fields []*types.Var

	for _, f := range n.Fields {
		d := types.NewVar(f.Identifier.Value, unresolved)
		t := c.evaluateTypeExpression(f.Identifier.AnnotatedType, def.TypeParameters, ctx)
		d.SetType(t)
		fields = append(fields, d)
		err := def.GetScope().Define(d)

		if err != nil {
			c.addError(err.Error(), f.Identifier.Range())
		}
	}

	typ := types.NewStruct(fields)
	def.SetType(typ)
}

func (c *Checker) checkEnumStatement(n *ast.EnumStatement, ctx *NodeContext) {
	def := c.resolve(n.Identifier, n, ctx.scope)

	// already reported
	if def == nil {
		return
	}

	// 2  Parse Generic Params
	if n.GenericParams != nil {
		for i, p := range n.GenericParams.Parameters {
			tP := def.TypeParameters[i]
			c.evaluateTypeParamterStandards(p, tP, ctx)
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
		cases[name] = true

		// set fields
		fields := []*types.Var{}

		if v.Fields != nil {
			for _, f := range v.Fields.Fields {
				t := c.evaluateTypeExpression(f, def.TypeParameters, ctx)
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

	e := types.NewEnum(n.Identifier.Value, variants)
	def.SetType(e)
}

func (c *Checker) checkSwitchStatement(n *ast.SwitchStatement, ctx *NodeContext) {

	// 1 - Condition
	condition := c.evaluateExpression(n.Condition, ctx)

	// 2 - Cases

	seenDefault := false

	if len(n.Cases) == 0 {
		c.addError("expected at least one case", n.Range())
		return
	}

	for _, cs := range n.Cases {

		scope := types.NewScope(ctx.scope, "")
		// Default Case
		if cs.IsDefault {
			// 1 - Check default has already been seen
			if seenDefault {
				c.addError("default case already added", cs.Range())
				continue
			}

			// 2 - Block
			ctx := NewContext(scope, ctx.sg, nil)
			c.checkBlockStatement(cs.Action, ctx)
			seenDefault = true
			continue
		}

		// 1 - Condition
		ctx := NewContext(scope, ctx.sg, condition)
		// For Tuple types, provide lhsType, which provides the generic specializations & correct fn signature when required
		caseCondition := c.evaluateExpression(cs.Condition, ctx)

		_, err := c.validate(condition, caseCondition)

		if err != nil {
			c.addError(err.Error(), cs.Condition.Range())
			continue
		}

		// 2 - Block
		c.checkBlockStatement(cs.Action, ctx)
	}
}

func (c *Checker) checkWhileStatement(n *ast.WhileStatement, ctx *NodeContext) {
	condition := c.evaluateExpression(n.Condition, ctx)

	_, err := c.validate(types.LookUp(types.Bool), condition)

	if err != nil {
		c.addError(err.Error(), n.Condition.Range())
		return
	}

	scope := types.NewScope(ctx.scope, fmt.Sprintf("__if_Block__%v", n))
	newCtx := NewContext(scope, ctx.sg, ctx.lhs)
	c.checkBlockStatement(n.Action, newCtx)
}

func (c *Checker) checkTypeStatement(n *ast.TypeStatement, ctx *NodeContext) {

	// Fetch Alias
	var alias *types.Alias
	for alias == nil {

		v := c.table.GetNodeType(n)

		if v != nil {
			alias = types.AsAlias(v)
		} else {
			c.defineAlias(n, ctx)
		}
	}

	// 2 - Evaluate Type Paramters

	if n.GenericParams != nil {
		for i, p := range n.GenericParams.Parameters {
			tP := alias.TypeParameters[i]
			c.evaluateTypeParamterStandards(p, tP, ctx)
		}
	}

	// 3 - RHS Type
	var RHS types.Type

	if n.Value != nil {
		fmt.Println("[DEBUG] Resolving RHS for Alias", n.Identifier.Value)
		RHS = c.evaluateTypeExpression(n.Value, alias.TypeParameters, ctx)
		alias.SetType(RHS)
	}
}
