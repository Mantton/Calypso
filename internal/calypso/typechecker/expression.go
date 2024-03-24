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
		c.checkFunctionExpression(expr, ctx)
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

func (c *Checker) checkFunctionExpression(e *ast.FunctionExpression, ctx *NodeContext) {
	c.evaluateFunctionExpression(e, ctx, nil, true)
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
		// c.table.AddNode(expr, types.LookUp(types.IntegerLiteral), nil, nil)
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
		return c.evaluatePropertyExpression(expr, ctx)
	case *ast.GenericSpecializationExpression:
		return c.evaluateGenericSpecializationExpression(expr, ctx)
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

	s, ok := ctx.scope.Resolve(expr.Value)

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is not defined", expr.Value),
			expr.Range(),
		)

		return unresolved
	}

	return s.Type()
}

func (c *Checker) evaluateGroupedExpression(expr *ast.GroupedExpression, ctx *NodeContext) types.Type {
	return c.evaluateExpression(expr.Expr, ctx)
}

func (c *Checker) evaluateCallExpression(expr *ast.CallExpression, ctx *NodeContext) types.Type {
	typ := c.evaluateExpression(expr.Target, ctx)
	// already reported error
	if typ == unresolved {
		return typ
	}

	switch typ := typ.(type) {
	case *types.FunctionSignature:
		// check if invocable
		fn := typ

		// Enum Switch Spread
		if ctx.lhs != nil {
			return c.evaluateEnumDestructure(ctx.lhs, fn, expr, ctx)
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
		specializations := make(map[string]types.Type)
		if !isGeneric {
			for i, arg := range expr.Arguments {
				expected := fn.Parameters[i]

				if expected.ParamLabel != arg.GetLabel() {
					c.addError(fmt.Sprintf("missing paramter label \"%s\"", expected.ParamLabel), arg.Range())
				}

				err := c.resolveVar(expected, arg.Value, specializations, ctx)

				if err != nil {
					c.addError(err.Error(), arg.Range())
				}
			}
			return fn.Result.Type()
		}

		// Function is generic, instantiate
		for i, arg := range expr.Arguments {
			expected := fn.Parameters[i]

			if expected.ParamLabel != arg.GetLabel() {
				c.addError(fmt.Sprintf("missing paramter label \"%s\"", expected.ParamLabel), arg.Range())
			}

			err := c.resolveVar(expected, arg.Value, specializations, ctx)

			if err != nil {
				c.addError(err.Error(), arg.Range())
			}
		}

		t := types.Apply(specializations, fn)
		fmt.Println("\nInstantiated Function:", t)
		fmt.Println("Original:", fn)
		fmt.Println()
		return t.(*types.FunctionSignature).Result.Type()

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
			v := types.NewVar("", vT)
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
			return single.Sg().Result.Type()
		}

		return options.MostSpecialized()
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

	if lhs == unresolved || rhs == unresolved {
		return unresolved
	}

	op := e.Op

	typ, err := c.validate(lhs, rhs)

	if err != nil {
		c.addError(err.Error(), e.Range())
		return unresolved
	}

	switch op {
	case token.PLUS, token.MINUS, token.QUO, token.STAR:
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
	}

	// no matching operand
	err = fmt.Errorf("unsupported binary operand `%s` on `%s`", token.LookUp(op), lhs)
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

	// 1 - Find Defined Type

	target := c.evaluateExpression(n.Target, ctx)
	base := types.AsDefined(target)

	if base == nil {
		c.addError(
			fmt.Sprintf("`%s` is not a type", target),
			n.Target.Range(),
		)
		return unresolved
	}

	// 2 - Ensure Defined Type is Struct
	if !types.IsStruct(base.Parent()) {
		c.addError(
			fmt.Sprintf("`%s` is not a struct", target),
			n.Target.Range(),
		)
		return unresolved
	}

	// 3 - check for annotated specialization

	// 4 - Get Struct Signature of Defined Type

	sg := base.Parent().(*types.Struct)

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

	specializations := make(map[string]types.Type)

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

	if len(base.TypeParameters) == 0 {
		return base
	}

	for _, param := range base.TypeParameters {
		if param.Bound != nil {
			continue
		}

		_, ok := specializations[param.Name()]

		if !ok {
			panic(fmt.Errorf("failed to resolve generic parameter: %s in %s", param, specializations))
		}
	}

	return types.Apply(specializations, base)
}

func (c *Checker) resolveVar(f *types.Var, v ast.Expression, specializations Specializations, ctx *NodeContext) error {
	vT := c.evaluateExpression(v, ctx)

	fmt.Println("\n", "\t[Resolver] Variable Name", f.Name(), "\n", "\t[Resolver] Variable Type", f.Type(), "\n", "\t[Resolver] Provided Type", vT)
	if vT == unresolved {
		return fmt.Errorf("unresolved type assigned for `%s`", f.Name())
	}

	// check constraints & specialize
	// can either be a type param or generic struct or a generic function
	fT := types.ResolveAliases(f.Type())
	switch fT := fT.(type) {
	case *types.TypeParam:
		return specializations.specialize(fT, vT, c, v)
	case *types.FunctionSignature:
		if !types.IsGeneric(fT) {
			break
		}
		panic("not implemented")
	case *types.DefinedType:

		// Not Generic , nothing to do, use typical validation
		if !types.IsGeneric(fT) {
			break
		}
		// Convert to Defined type
		prev := vT
		vT := types.AsDefined(vT)
		if vT == nil {
			panic(fmt.Errorf("type is not a defined type : %T (%s)", prev, prev))
		}

		// validate vT can be passed to fT
		_, err := c.validate(fT, vT)

		if err != nil {
			fmt.Println("[DEBUG]", err)
			return err
		}

		// Now sides are of the same type & should be of the same parameter length
		// guard in situation where not
		if len(fT.TypeParameters) != len(vT.TypeParameters) {
			return fmt.Errorf("expected %d type parameter(s), got %d instead", len(fT.TypeParameters), len((vT.TypeParameters)))
		}

		// There for, Foo<T> == Foo<V>, we can infer T == V
		for i, fTParam := range fT.TypeParameters {
			vTParam := vT.TypeParameters[i]
			return specializations.specialize(fTParam, vTParam, c, v)
		}
		return nil
	case *types.Alias:
		panic("fT should be resolved, bad path")

	case *types.Pointer:

		// non generic pointer type
		if !types.IsGeneric(fT) {
			break
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

		return specializations.specialize(base, iT, c, v)
	}

	// resolve non generic types
	err := c.validateAssignment(f, vT, v)
	return err
}

func (c *Checker) evaluatePropertyExpression(n *ast.FieldAccessExpression, ctx *NodeContext) types.Type {

	a := c.evaluateExpression(n.Target, ctx)

	if a == unresolved {
		return unresolved
	}

	// Collect Property
	var field string
	switch p := n.Field.(type) {
	case *ast.IdentifierExpression:
		field = p.Value
	default:
		c.addError("invalid property key", n.Range())
		return unresolved
	}

	switch a := a.(type) {
	case *types.DefinedType:
		f := a.ResolveField(field)
		if f != nil {
			return f
		}
	}

	c.addError(fmt.Sprintf("unknown method or field: \"%s\" on type \"%s\"", field, a), n.Range())
	return unresolved
}

func (c *Checker) evaluateGenericSpecializationExpression(e *ast.GenericSpecializationExpression, ctx *NodeContext) types.Type {
	// 1- Find Target
	sym, ok := ctx.scope.Resolve(e.Identifier.Value)
	if !ok {
		msg := fmt.Sprintf("could not find `%s` in scope", e.Identifier.Value)
		c.addError(msg, e.Range())
		return unresolved
	}

	// 2 - Collect Type Parameters

	args := []types.Type{}
	params := []*types.TypeParam{}

	switch sym := sym.(type) {
	case *types.Function:
		sg := sym.Sg()
		params = sg.TypeParameters
	case *types.DefinedType:
		params = sym.TypeParameters
	}

	// 3 - Ensure Type is generic
	if len(params) == 0 {
		msg := fmt.Sprintf("`%s` cannot be specialized", e.Identifier.Value)
		c.addError(msg, e.Identifier.Range())
		return unresolved
	}

	// 4 - Collect Args
	for _, t := range e.Clause.Arguments {
		args = append(args, c.evaluateTypeExpression(t, nil, ctx))
	}

	// 5 - Ensure Length match
	if len(params) != len(args) {
		c.addError(fmt.Sprintf("expected %d arguments provided %d", len(params), len(args)), e.Range())
		return unresolved
	}

	// 6 - Ensure Conformance For Each Argument
	hasError := false
	for i, arg := range args {
		param := params[i]

		err := types.Conforms(param.Constraints, arg)

		if err != nil {
			c.addError(err.Error(), e.Clause.Arguments[i].Range())
			hasError = true
		}
	}

	if hasError {
		return unresolved
	}

	// 7 - Return instance of type
	instance := types.Instantiate(sym.Type(), args, nil)
	return instance
}

func (c *Checker) evaluateFunctionExpression(e *ast.FunctionExpression, ctx *NodeContext, self *types.DefinedType, define bool) types.Type {
	// Create new function

	sg := types.NewFunctionSignature()
	def := types.NewFunction(e.Identifier.Value, sg)
	c.table.DefineFunction(e, def)

	// Enter Function Scope
	sg.Scope = types.NewScope(ctx.scope)
	c.table.AddScope(e, sg.Scope)

	// inject `self`
	if self != nil {
		s := types.NewVar("self", self)
		sg.Scope.Define(s)
	}

	// Type/Generic Parameters
	hasError := false
	if e.GenericParams != nil {
		for _, p := range e.GenericParams.Parameters {
			t := c.evaluateGenericParameterExpression(p, ctx)
			if t == unresolved {
				hasError = true
				continue
			}

			sg.AddTypeParameter(t.(*types.TypeParam))
		}
	}

	if hasError {
		return unresolved
	}

	// Parameters
	for _, p := range e.Parameters {
		t := c.evaluateTypeExpression(p.Type, sg.TypeParameters, ctx)

		// Placeholder / Discard

		v := types.NewVar(p.Name.Value, t)

		// Parameter Has Required Label
		if p.Label.Value != "_" {
			v.ParamLabel = p.Label.Value
		}

		sg.AddParameter(v)

		if p.Name.Value == "_" {
			continue
		}
		err := sg.Scope.Define(v)

		if err != nil {
			c.addError(err.Error(), p.Range())
		}
	}

	// Annotated Return Type
	if e.ReturnType != nil {
		t := c.evaluateTypeExpression(e.ReturnType, sg.TypeParameters, ctx)
		sg.Result = types.NewVar("result", t)
	} else {
		sg.Result = types.NewVar("result", types.LookUp(types.Void))
	}

	if define {
		// At this point the signature has been constructed fully, add to scope
		err := ctx.scope.Define(def)

		if err != nil {
			c.addError(err.Error(), e.Identifier.Range())
			return unresolved
		}

	}

	// Body
	newCtx := NewContext(sg.Scope, sg, nil)
	c.checkBlockStatement(e.Body, newCtx)

	// TODO:
	// Ensure All Generic Params are used
	// Ensure All Params are used
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

	sym, ok := ctx.scope.Resolve("Array")

	if !ok {
		c.addError("unable to find array type", n.Range())
		return unresolved
	}
	typ := types.AsDefined(sym.Type())
	instantiated := types.Instantiate(typ, []types.Type{element}, nil)
	return instantiated
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

	sym, ok := ctx.scope.Resolve("Map")

	if !ok {
		c.addError("unable to find map type", n.Range())
		return unresolved
	}
	typ := types.AsDefined(sym.Type())
	instantiated := types.Instantiate(typ, []types.Type{key, value}, nil)
	return instantiated
}

func (c *Checker) evaluateIndexExpression(n *ast.IndexExpression, ctx *NodeContext) types.Type {

	// 1 - Eval Target
	target := c.evaluateExpression(n.Target, ctx)

	// 2 - Eval Subscript Standard
	symbol, ok := ctx.scope.Resolve("SubscriptStandard")

	if !ok {
		c.addError("unable to find subscript standard", n.Range())
		return unresolved
	}

	standardDefinition := types.AsDefined(symbol.Type())

	if standardDefinition == nil {
		c.addError("subscript is not a defined type", n.Range())
		return unresolved
	}

	standard := types.AsStandard(standardDefinition.Parent())
	if standard == nil {
		c.addError("subscript is not a standard", n.Range())
		return unresolved
	}

	// 3 - Get Target Definition

	definition := types.AsDefined(target)

	if definition == nil {
		c.addError(fmt.Sprintf("\"%s\" is not a defined type", target), n.Range())
		return unresolved
	}

	// 4 - Validate Conformance to Subscript Standard
	err := types.Conforms([]*types.Standard{standard}, target)

	if err != nil {
		c.addError(err.Error(), n.Target.Range())
		return unresolved
	}

	// 5 - Get Index Type

	indexSymbol := definition.ResolveType("Index")

	if indexSymbol == nil {
		c.addError("Unable to resolve type: \"Index\"", n.Target.Range())
		return unresolved
	}

	indexType := types.AsDefined(indexSymbol)

	if indexType == nil {
		c.addError(fmt.Sprintf("%s is not a defined type", indexSymbol), n.Target.Range())
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
	elementSymbol := definition.ResolveType("Element")

	if elementSymbol == nil {
		c.addError("Unable to locate Element Type", n.Target.Range())
		return unresolved
	}

	elementType := types.AsDefined(elementSymbol)

	if elementType == nil {
		c.addError(fmt.Sprintf("%s is not a defined type", elementSymbol), n.Target.Range())
		return unresolved
	}

	return elementType
}

func (c *Checker) evaluateEnumDestructure(inc types.Type, fn *types.FunctionSignature, expr *ast.CallExpression, ctx *NodeContext) types.Type {
	lhsTyp := types.AsDefined(inc)

	if lhsTyp == nil {
		c.addError(fmt.Sprintf("expected defined type, got %s, %s", lhsTyp, fn), expr.Range())
		return unresolved
	}

	specializations := make(map[string]types.Type)

	for _, p := range lhsTyp.TypeParameters {
		specializations[p.Name()] = p
	}

	sg := types.Apply(specializations, fn).(*types.FunctionSignature)

	fmt.Println("Instantiated: ", sg)

	if len(sg.Parameters) != len(expr.Arguments) {
		c.addError(fmt.Sprintf("expected %d arguments, got %d", len(sg.Parameters), len(fn.Parameters)), expr.Range())
		return unresolved
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

		err := ctx.scope.Define(types.NewVar(ident.Value, t.Type()))

		if err != nil {
			c.addError(err.Error(), arg.Range())
			return unresolved
		}
	}

	return sg.Result.Type()
}
