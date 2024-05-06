package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
	"github.com/mantton/calypso/internal/calypso/types"
)

// ---------------------- Checks ---------------------------------
func (c *Checker) checkExpression(expr ast.Expression, ctx *NodeContext) {

	fmt.Printf(
		"Checking Expression: %T @ Line %d\n",
		expr,
		expr.Range().Start.Line,
	)
	switch expr := expr.(type) {
	case *ast.FunctionExpression:
		c.checkFunctionExpression(expr)
	case *ast.AssignmentExpression:
		c.checkAssignmentExpression(expr, ctx)
	case *ast.CallExpression:
		c.checkCallExpression(expr, ctx)
	case *ast.ShorthandAssignmentExpression:
		c.CheckShorthandAssignment(expr, ctx)
	default:
		msg := fmt.Sprintf("expression check not implemented, %T", expr)
		panic(msg)
	}
}

func (c *Checker) checkFunctionExpression(e *ast.FunctionExpression) {
	c.evaluateFunctionExpression(e)
}

func (c *Checker) checkAssignmentExpression(expr *ast.AssignmentExpression, ctx *NodeContext) {
	// TODO: mutability checks
	c.evaluateAssignmentExpression(expr, ctx)
}

func (c *Checker) CheckShorthandAssignment(expr *ast.ShorthandAssignmentExpression, ctx *NodeContext) {
	// TODO: Mutability checks
	c.evaluateShorthandAssignmentExpression(expr, ctx)
}

func (c *Checker) checkCallExpression(expr *ast.CallExpression, ctx *NodeContext) {

	retType := c.evaluateCallExpression(expr, ctx)

	if retType != types.LookUp(types.Void) || retType != unresolved {
		fmt.Println("[WARNING] Call Expression returning non void value is unused")
	}

	// check if function has not been resolved
	if t, ok := retType.(*types.FunctionSet); ok {
		c.addError(fmt.Sprintf("ambagious use of function \"%s\"", t.Name()), expr.Range())
	}
}

// ----------- Eval ------------------
func (c *Checker) evaluateExpression(expr ast.Expression, ctx *NodeContext) types.Type {
	fmt.Printf(
		"Evaluating Expression: %T @ Line %d\n",
		expr,
		expr.Range().Start.Line,
	)
	switch expr := expr.(type) {
	// Literals
	case *ast.IntegerLiteral:
		return types.LookUp(types.IntegerLiteral)
	case *ast.BooleanLiteral:
		return types.LookUp(types.Bool)
	case *ast.FloatLiteral:
		return types.LookUp(types.FloatLiteral)
	case *ast.StringLiteral:
		return types.LookUp(types.String)
	case *ast.CharLiteral:
		return types.LookUp(types.Char)
	case *ast.NilLiteral:
		return types.LookUp(types.NilLiteral)
	case *ast.VoidLiteral:
		return types.LookUp(types.Void)

	case *ast.IdentifierExpression:
		return c.evaluateIdentifierExpression(expr, ctx)
	case *ast.GroupedExpression:
		return c.evaluateGroupedExpression(expr, ctx)
	case *ast.CallExpression:
		return c.evaluateCallExpression(expr, ctx)
	case *ast.UnaryExpression:
		return c.evaluateUnaryExpression(expr, ctx)
	case *ast.BinaryExpression:
		return c.evaluateBinaryExpression(expr, ctx)
	case *ast.AssignmentExpression:
		return c.evaluateAssignmentExpression(expr, ctx)
	case *ast.CompositeLiteral:
		return c.evaluateCompositeLiteral(expr, ctx)
	case *ast.FieldAccessExpression:
		return c.evaluateFieldAccessExpression(expr, ctx)
	case *ast.SpecializationExpression:
		return c.evaluateSpecializationExpression(expr, ctx)
	case *ast.ArrayLiteral:
		return c.evaluateArrayLiteral(expr, ctx)
	case *ast.MapLiteral:
		return c.evaluateMapLiteral(expr, ctx)
	case *ast.IndexExpression:
		return c.evaluateIndexExpression(expr, ctx)
	default:
		msg := fmt.Sprintf("expression evaluation not implemented, %T", expr)
		panic(msg)
	}
}

func (c *Checker) evaluateIdentifierExpression(expr *ast.IdentifierExpression, ctx *NodeContext) types.Type {

	s, ok := ctx.scope.Resolve(expr.Value, c.ParentScope())

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is not defined", expr.Value),
			expr.Range(),
		)

		return unresolved
	}

	if !s.IsVisible(c.module) {
		c.addError(
			fmt.Sprintf("`%s` is not accessible in this context", expr.Value),
			expr.Range(),
		)
	}

	return s.Type()
}

func (c *Checker) evaluateGroupedExpression(expr *ast.GroupedExpression, ctx *NodeContext) types.Type {
	return c.evaluateExpression(expr.Expr, ctx)
}

func (c *Checker) evaluateCallExpression(expr *ast.CallExpression, ctx *NodeContext) types.Type {
	typ := c.evaluateExpression(expr.Target, ctx)
	// already reported error
	if types.IsUnresolved(typ) {
		return typ
	}

	switch typ := typ.(type) {
	case *types.SpecializedFunctionSignature:
		fn := typ
		if len(expr.Arguments) != len(typ.InstanceOf.Parameters) {
			c.addError(
				fmt.Sprintf("expected %d arguments, provided %d",
					len(fn.InstanceOf.Parameters),
					len(expr.Arguments)),
				expr.Range(),
			)

			return fn.ReturnType()
		}

		c.module.Table.SetNodeType(expr, typ)
		ctx.sg.Function.AddCallEdge(typ)
		specializations := make(types.Specialization)
		for i, arg := range expr.Arguments {
			param := fn.InstanceOf.Parameters[i]
			expected, err := c.instantiateWithSpecialization(param.Type(), fn.Specialization())

			if err != nil {
				c.addError(err.Error(), arg.Range())
				return fn.ReturnType()
			}

			if param.ParamLabel != arg.GetLabel() {
				c.addError(fmt.Sprintf("missing paramter label \"%s\"", param.ParamLabel), arg.Range())
			}

			I := types.NewVar(param.Name(), expected, c.module)
			err = c.resolveVar(I, arg.Value, specializations, ctx)

			if err != nil {
				c.addError(err.Error(), arg.Range())
			}
		}

		return fn.ReturnType()

	case *types.FunctionSignature:
		// check if invocable
		fn := typ

		// Enum Switch Spread
		if ctx.lhs != nil {
			return c.evaluateEnumDestructure(fn, expr, ctx)
		}

		isGeneric := types.IsGeneric(fn)
		// Guard Argument Count == Parameter Count
		if len(expr.Arguments) != len(fn.Parameters) {
			c.addError(
				fmt.Sprintf("expected %d arguments, provided %d",
					len(fn.Parameters),
					len(expr.Arguments)),
				expr.Range(),
			)

			// if is generic, return unresolved as we are unable to properly infer the type
			if isGeneric {
				return unresolved
			} else {
				return fn.Result.Type()
			}
		}

		// if not generic, simply check arguments & return correct function type regardless of error
		hasError := false
		specializations := make(types.Specialization)
		for i, arg := range expr.Arguments {
			expected := fn.Parameters[i]

			if expected.ParamLabel != arg.GetLabel() {
				c.addError(fmt.Sprintf("missing paramter label \"%s\"", expected.ParamLabel), arg.Range())
			}

			err := c.resolveVar(expected, arg.Value, specializations, ctx)

			if err != nil {
				c.addError(err.Error(), arg.Range())
				hasError = true
			}
		}

		if hasError {
			if isGeneric {
				return unresolved
			} else {
				return fn.Result.Type()
			}
		}

		// return signature if not generic
		if !isGeneric {
			c.module.Table.SetNodeType(expr, fn)
			ctx.sg.Function.AddCallEdge(fn)
			return fn.Result.Type()
		}

		// Function is generic, instantiate
		t, err := c.instantiateWithSpecialization(fn, specializations)

		if err != nil {
			c.addError(err.Error(), expr.Target.Range())
			return unresolved
		}

		c.module.Table.SetNodeType(expr, t)
		ctx.sg.Function.AddCallEdge(t)

		switch t := t.(type) {
		case *types.FunctionSignature:
			return t.Result.Type()
		case *types.SpecializedFunctionSignature:
			return t.ReturnType()
		}

	case *types.FunctionSet:

		// Overloaded Function | Function Set
		set := typ

		// build signature of call expression
		callSg := types.NewFunctionSignature()
		callSg.Result.SetType(types.LookUp(types.Placeholder)) // not going to be used, but stated as a safety mechanism

		for _, arg := range expr.Arguments {
			label := ""
			if arg.Label != nil {
				label = arg.Label.Value
			}
			vT := c.evaluateExpression(arg.Value, ctx)
			v := types.NewVar("", vT, c.module)
			v.ParamLabel = label
			callSg.AddParameter(v)
		}

		// Find Possible Options
		options := set.Find(callSg, false)

		// No Options found
		if options == nil {
			c.addError(("no exact match for overloaded function"), expr.Range())
			return unresolved
		}

		// check is there is an exact match
		if single, ok := options.GetAsSingle(); ok {
			fmt.Println("\t[OVERLOAD] Exact match", single.Sg())

			c.module.Table.SetNodeType(expr, single.Sg())

			return single.Sg().Result.Type()
		}
		mostSpec := options.MostSpecialized()
		// TODO: Call Edge
		c.module.Table.SetNodeType(expr, mostSpec)
		return mostSpec
	}

	c.addError(
		"expression is not invocable",
		expr.Target.Range(),
	)
	return unresolved
}

func (c *Checker) evaluateUnaryExpression(expr *ast.UnaryExpression, ctx *NodeContext) types.Type {
	op := expr.Op
	rhs := c.evaluateExpression(expr.Expr, ctx)
	var err error

	switch op {
	case token.NOT:

		// if RHS is boolean, return boolean type. as not inverts boolean value
		bl := types.LookUp(types.Bool)
		if rhs == bl {
			return bl
		}

		// RHS is not a boolean, ensure type conforms to NOT operan standard
		// TODO: ^^^
		panic("operand standards have not been implemented")

	case token.MINUS:

		// if RHS is numeric, return RHS type as numeric types can be negated
		if types.IsNumeric(rhs) {
			return rhs
		}

		err = fmt.Errorf("unsupported negation operand on type `%s`", rhs)
	case token.AMP:
		if types.IsGroupLiteral(rhs) {
			err = fmt.Errorf("cannot get reference of type \"%s\"", rhs)
			break
		}
		return types.NewPointer(rhs)

	case token.STAR:
		ptr, ok := rhs.(*types.Pointer)

		if !ok {
			err = fmt.Errorf("cannot dereference non-pointer type \"%s\"", rhs)
			break
		}

		return ptr.PointerTo
	default:
		err = fmt.Errorf("unsupported unary operand `%s`", token.LookUp(op))
	}

	if err == nil {
		panic("there should be an error here")
	}

	c.addError(err.Error(), expr.Range())

	return unresolved

}

func (c *Checker) evaluateBinaryExpression(e *ast.BinaryExpression, ctx *NodeContext) types.Type {
	lhs := c.evaluateExpression(e.Left, ctx)
	rhs := c.evaluateExpression(e.Right, ctx)

	if types.IsUnresolved(lhs) || types.IsUnresolved(rhs) {
		return unresolved
	}

	op := e.Op

	typ, err := c.validate(lhs, rhs)
	if err != nil {
		c.addError(err.Error(), e.Range())
		return unresolved
	}

	switch op {
	case token.PLUS, token.MINUS, token.QUO, token.STAR, token.PCT:
		if types.IsNumeric(typ) {
			return typ
		}
	case token.L_CHEVRON, token.R_CHEVRON, token.LEQ, token.GEQ:
		if types.IsNumeric(typ) {
			return types.LookUp(types.Bool)
		}
		panic("COMPARABLE STANDARD NOT IMPLEMENTED")

	case token.EQL, token.NEQ:
		if types.IsEquatable(typ) {
			return types.LookUp(types.Bool)
		}
		panic("EQUATABLE STANDARD NOT IMPLEMENTED")

	case token.DOUBLE_AMP, token.DOUBLE_BAR:
		if types.IsBoolean(typ) {
			return typ
		}

	case token.BIT_SHIFT_LEFT, token.BIT_SHIFT_RIGHT,
		token.BAR, token.AMP, token.CARET:
		if types.IsInteger(typ) {
			return typ
		}
	}

	// no matching operand
	err = fmt.Errorf("unsupported binary operand `%s` on `%s`", op, lhs)
	c.addError(err.Error(), e.Range())
	return unresolved

}

func (c *Checker) evaluateAssignmentExpression(expr *ast.AssignmentExpression, ctx *NodeContext) types.Type {

	lhs := c.evaluateExpression(expr.Target, ctx)
	rhs := c.evaluateExpression(expr.Value, ctx)

	_, err := c.validate(lhs, rhs)

	if err != nil {
		c.addError(err.Error(), expr.Range())
	}

	// assignment yield void
	return types.LookUp(types.Void)
}

func (c *Checker) evaluateShorthandAssignmentExpression(expr *ast.ShorthandAssignmentExpression, ctx *NodeContext) types.Type {
	lhs := c.evaluateExpression(expr.Target, ctx)
	rhs := c.evaluateExpression(expr.Right, ctx)

	_, err := c.validate(lhs, rhs)

	if err != nil {
		c.addError(err.Error(), expr.Range())
	}

	// assignment yield void
	return types.LookUp(types.Void)
}

func (c *Checker) evaluateCompositeLiteral(n *ast.CompositeLiteral, ctx *NodeContext) types.Type {

	// 1 - Find Type
	target := c.evaluateExpression(n.Target, ctx)

	if types.IsUnresolved(target) {
		return unresolved
	}
	base := target

	sg, ok := base.Parent().(*types.Struct)

	// 2 - Ensure Defined Type is Struct
	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is not a struct", target),
			n.Target.Range(),
		)
		return unresolved
	}

	// 3 - Get Struct Signature of Defined Type
	seen := make(map[string]ast.Expression)
	hasError := false

	// Collect Fields
	for _, p := range n.Body.Fields {
		k := p.Key.Value
		v := p.Value

		f := sg.FindField(k)
		// 1 - invalid field, report
		if f == nil {
			c.addError(
				fmt.Sprintf("`%s` is not a valid field", k),
				p.Range(),
			)
			hasError = true
			continue
		}

		_, ok := seen[k]

		// 2 - field has already been evaluated
		if ok {
			c.addError(
				fmt.Sprintf("`%s` has already been provided", k),
				p.Range(),
			)
			hasError = true
			continue
		}

		seen[k] = v
	}

	if len(seen) != len(sg.Fields) {
		c.addError(
			fmt.Sprintf("missing fields, %d", len(sg.Fields)),
			n.Range(),
		)
		hasError = true
	}

	if hasError {
		return unresolved
	}

	specializations := make(types.Specialization)

	for k, v := range seen {

		f := sg.FindField(k)
		err := c.resolveVar(f, v, specializations, ctx)

		if err != nil {
			c.addError(err.Error(), v.Range())
			hasError = true
		}
	}

	if hasError {
		return unresolved
	}

	tparams := types.GetTypeParams(target)
	if len(tparams) == 0 {
		c.module.Table.SetNodeType(n, base)
		return base
	}

	for _, param := range tparams {
		_, ok := specializations[param]
		if !ok {
			panic(fmt.Errorf("failed to resolve generic parameter: %s in %s", param, specializations))
		}
	}

	inst, err := c.instantiateWithSpecialization(base, specializations)

	if err != nil {
		c.addError(err.Error(), n.Range())
		return unresolved
	}

	c.module.Table.SetNodeType(n, inst)
	return inst
}

func (c *Checker) resolveVar(f *types.Var, v ast.Expression, specializations types.Specialization, ctx *NodeContext) error {
	vT := c.evaluateExpression(v, ctx)

	if types.IsUnresolved(vT) {
		return fmt.Errorf("unresolved type assigned for `%s`", f.Name())
	}

	fmt.Println("\n", "\t[Resolver] Variable Name", f.Name(), "\n", "\t[Resolver] Variable Type", f.Type(), "\n", "\t[Resolver] Provided Type", vT)
	// check constraints & specialize
	// can either be a type param or generic struct or a generic function
	fT := types.ResolveAliases(f.Type())

	if !types.IsGeneric(fT) {
		// resolve non generic types
		err := c.validateAssignment(f, vT, v, false)
		return err
	}

	fmt.Println("\t[Resolver] Specializing", fT, "with", vT)
	switch fT := fT.(type) {
	case *types.TypeParam:
		return c.specialize(specializations, fT, vT, v)

	case *types.Pointer:
		_, err := c.validate(fT, vT)
		if err != nil {
			return err
		}

		base := types.AsTypeParam(fT.PointerTo)

		if base == nil {
			panic("expected type parameter")
		}
		// vT must be an instantiated struct of type fT
		iT, ok := vT.(*types.Pointer)
		if !ok {
			return fmt.Errorf("expected pointer to type %s, received %s", fT, vT)
		}

		return c.specialize(specializations, base, iT, v)
	case *types.SpecializedType:
		// field type is a specialized instance of a type
		_, err := c.validate(fT, vT)

		if err != nil {
			return err
		}

		svT := vT.(*types.SpecializedType)

		for i, b := range fT.Bounds {

			tfT := types.AsTypeParam(b)

			if tfT == nil {
				continue
			}

			err := c.specialize(specializations, tfT, svT.Bounds[i], v)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Checker) evaluateFieldAccessExpression(n *ast.FieldAccessExpression, ctx *NodeContext) types.Type {

	a := c.evaluateExpression(n.Target, ctx)
	if types.IsUnresolved(a) {
		return unresolved
	}

	// Collect Property
	var field string
	switch p := n.Field.(type) {
	case *ast.IdentifierExpression:
		field = p.Value
	default:
		if a, ok := a.(*types.Module); ok {
			sc := a.Scope
			return c.evaluateExpression(p, NewContext(sc, ctx.sg, nil))
		}

		c.addError("invalid field key", n.Range())
		return unresolved
	}

	symbol, symbolType := types.ResolveSymbol(a, field)

	// DNE
	if symbol == nil {
		c.addError(fmt.Sprintf("unable to located '%s', field", field), n.Field.Range())
		return unresolved
	}

	// Private
	// if !symbol.IsVisible(c.module) {
	// 	c.addError(fmt.Sprintf("cannot access '%s' from this context", field), n.Field.Range())
	// 	return unresolved
	// }

	c.module.Table.SetNodeType(n, symbolType)
	return symbolType
}

func (c *Checker) evaluateSpecializationExpression(e *ast.SpecializationExpression, ctx *NodeContext) types.Type {

	// 1- Find Target
	target := c.evaluateExpression(e.Expression, ctx)

	if types.IsUnresolved(target) {
		return unresolved
	}

	// 2 - Collect Type Parameters
	args := []types.Type{}
	// 4 - Collect Args
	for _, t := range e.Clause.Arguments {
		args = append(args, c.evaluateTypeExpression(t, nil, ctx))
	}

	// 7 - Return instance of type
	instance, err := c.instantiateWithArguments(target, args, e.Clause)

	if err != nil {
		c.addError(err.Error(), e.Clause.Range())
		return unresolved
	}

	c.module.Table.SetNodeType(e, instance)
	return instance
}

func (c *Checker) evaluateFunctionExpression(e *ast.FunctionExpression) types.Type {

	sg := c.module.Table.GetNodeType(e).(*types.FunctionSignature)

	if sg == nil {
		panic("unregistered node")
	}

	fn := sg.Function

	if fn == nil {
		panic("passes missed function")
	}

	// Body
	newCtx := NewContext(sg.Function.Scope, sg, nil)
	c.checkBlockStatement(e.Body, newCtx)

	return sg
}

func (c *Checker) evaluateArrayLiteral(n *ast.ArrayLiteral, ctx *NodeContext) types.Type {

	var element types.Type

	if len(n.Elements) == 0 {
		c.addError("empty literal", n.Range())
		return unresolved
	}

	hasError := false
	for _, node := range n.Elements {
		provided := c.evaluateExpression(node, ctx)

		if element == nil && provided != unresolved {
			element = provided
			continue
		}

		_, err := c.validate(element, provided)

		if err != nil {
			c.addError(err.Error(), node.Range())
			hasError = true
		}
	}

	if hasError {
		return unresolved
	}

	sym, ok := ctx.scope.Resolve("Array", c.ParentScope())

	if !ok {
		c.addError("unable to find array type", n.Range())
		return unresolved
	}

	instance, err := c.instantiateWithArguments(sym.Type(), types.TypeList{element}, n)

	if err != nil {
		c.addError(err.Error(), n.Range())
		return unresolved
	}
	return instance
}
func (c *Checker) evaluateMapLiteral(n *ast.MapLiteral, ctx *NodeContext) types.Type {

	var key types.Type
	var value types.Type

	if len(n.Pairs) == 0 {
		c.addError("empty literal", n.Range())
		return unresolved
	}

	hasError := false
	for _, node := range n.Pairs {
		providedKey := c.evaluateExpression(node.Key, ctx)
		providedValue := c.evaluateExpression(node.Value, ctx)

		if key == nil && providedKey != unresolved {
			key = providedKey
		}

		if value == nil && providedValue != unresolved {
			value = providedValue
		}

		_, err := c.validate(key, providedKey)

		if err != nil {
			c.addError(err.Error(), node.Range())
			hasError = true
		}

		_, err = c.validate(value, providedValue)

		if err != nil {
			c.addError(err.Error(), node.Range())
			hasError = true
		}
	}

	if hasError {
		return unresolved
	}

	sym, ok := ctx.scope.Resolve("Map", c.ParentScope())

	if !ok {
		c.addError("unable to find map type", n.Range())
		return unresolved
	}
	instance, err := c.instantiateWithArguments(sym.Type(), types.TypeList{key, value}, n)

	if err != nil {
		c.addError(err.Error(), n.Range())
		return unresolved
	}
	return instance
}

func (c *Checker) evaluateIndexExpression(n *ast.IndexExpression, ctx *NodeContext) types.Type {

	// 1 - Eval Target
	target := c.evaluateExpression(n.Target, ctx)

	// 2 - Eval Subscript Standard
	symbol, ok := ctx.scope.Resolve("SubscriptStandard", c.ParentScope())

	if !ok {
		c.addError("unable to find subscript standard", n.Range())
		return unresolved
	}

	standard := types.AsStandard(symbol.Type().Parent())
	if standard == nil {
		c.addError("subscript is not a standard", n.Range())
		return unresolved
	}

	// 3 - Get Target Definition

	// 4 - Validate Conformance to Subscript Standard
	err := types.Conforms([]*types.Standard{standard}, target)

	if err != nil {
		c.addError(err.Error(), n.Target.Range())
		return unresolved
	}

	// 5 - Get Index Type

	indexType := types.ResolveType(target, "Index")

	if indexType == nil {
		c.addError("Unable to resolve type: \"Index\"", n.Target.Range())
		return unresolved
	}

	// 6 - Resolve & Validate Index Type
	index := c.evaluateExpression(n.Index, ctx)

	_, err = c.validate(indexType, index)

	if err != nil {
		c.addError(err.Error(), n.Index.Range())
		return unresolved
	}

	// - Validated At this point, return Element Type
	elementType := types.ResolveType(target, "Element")

	if elementType == nil {
		c.addError("Unable to locate Element Type", n.Target.Range())
		return unresolved
	}

	return elementType
}

func (c *Checker) evaluateEnumDestructure(fn types.Type, expr *ast.CallExpression, ctx *NodeContext) types.Type {

	var sg *types.FunctionSignature

	switch fn := fn.(type) {
	case *types.SpecializedFunctionSignature:
		sg = fn.Sg()
	case *types.FunctionSignature:
		sg = fn
	}

	switch lhs := ctx.lhs.(type) {
	case *types.SpecializedType:
		x, err := c.instantiateWithSpecialization(sg, lhs.Specialization())

		if err != nil {
			c.addError(err.Error(), expr.Range())
			return unresolved
		}

		sg = x.(*types.SpecializedFunctionSignature).Sg()
	}

	for i, t := range sg.Parameters {
		arg := expr.Arguments[i]

		if arg.Label != nil {
			c.addError("no labels allowed", arg.Range())
		}
		ident, ok := arg.Value.(*ast.IdentifierExpression)

		if !ok {
			c.addError("expected identifier", arg.Range())
			return unresolved
		}

		err := ctx.scope.Define(types.NewVar(ident.Value, t.Type(), c.module))

		if err != nil {
			c.addError(err.Error(), arg.Range())
			return unresolved
		}
	}

	return sg.Result.Type()
}
