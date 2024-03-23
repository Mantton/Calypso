package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
	"github.com/mantton/calypso/internal/calypso/types"
)

// ---------------------- Checks ---------------------------------
func (c *Checker) checkExpression(expr ast.Expression) {

	fmt.Printf(
		"Checking Expression: %T @ Line %d\n",
		expr,
		expr.Range().Start.Line,
	)
	switch expr := expr.(type) {
	case *ast.FunctionExpression:
		c.checkFunctionExpression(expr)
	case *ast.AssignmentExpression:
		c.checkAssignmentExpression(expr)
	case *ast.CallExpression:
		c.checkCallExpression(expr)
	case *ast.ShorthandAssignmentExpression:
		c.CheckShorthandAssignment(expr)
	default:
		msg := fmt.Sprintf("expression check not implemented, %T", expr)
		panic(msg)
	}
}

func (c *Checker) checkFunctionExpression(e *ast.FunctionExpression) {
	c.evaluateFunctionExpression(e, nil)
}

func (c *Checker) checkAssignmentExpression(expr *ast.AssignmentExpression) {
	// TODO: mutability checks
	c.evaluateAssignmentExpression(expr)
}

func (c *Checker) CheckShorthandAssignment(expr *ast.ShorthandAssignmentExpression) {
	// TODO: Mutability checks
	c.evaluateShorthandAssignmentExpression(expr)
}

func (c *Checker) checkCallExpression(expr *ast.CallExpression) {

	retType := c.evaluateCallExpression(expr)

	if retType != types.LookUp(types.Void) || retType != unresolved {
		fmt.Println("[WARNING] Call Expression returning non void value is unused")
	}

	// check if function has not been resolved
	if t, ok := retType.(*types.FunctionSet); ok {
		c.addError(fmt.Sprintf("ambagious use of function \"%s\"", t.Name()), expr.Range())
	}
}

// ----------- Eval ------------------
func (c *Checker) evaluateExpression(expr ast.Expression) types.Type {
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
		return c.evaluateIdentifierExpression(expr)
	case *ast.GroupedExpression:
		return c.evaluateGroupedExpression(expr)
	case *ast.CallExpression:
		return c.evaluateCallExpression(expr)
	case *ast.UnaryExpression:
		return c.evaluateUnaryExpression(expr)
	case *ast.BinaryExpression:
		return c.evaluateBinaryExpression(expr)
	case *ast.AssignmentExpression:
		return c.evaluateAssignmentExpression(expr)
	case *ast.CompositeLiteral:
		return c.evaluateCompositeLiteral(expr)
	case *ast.FieldAccessExpression:
		return c.evaluatePropertyExpression(expr)
	case *ast.GenericSpecializationExpression:
		return c.evaluateGenericSpecializationExpression(expr)
	case *ast.ArrayLiteral:
		return c.evaluateArrayLiteral(expr)
	case *ast.MapLiteral:
		return c.evaluateMapLiteral(expr)
	case *ast.IndexExpression:
		return c.evaluateIndexExpression(expr)
	default:
		msg := fmt.Sprintf("expression evaluation not implemented, %T", expr)
		panic(msg)
	}
}

func (c *Checker) evaluateIdentifierExpression(expr *ast.IdentifierExpression) types.Type {

	s, ok := c.find(expr.Value)

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is not defined", expr.Value),
			expr.Range(),
		)

		return unresolved
	}

	return s.Type()
}

func (c *Checker) evaluateGroupedExpression(expr *ast.GroupedExpression) types.Type {
	return c.evaluateExpression(expr.Expr)
}

func (c *Checker) evaluateCallExpression(expr *ast.CallExpression) types.Type {
	typ := c.evaluateExpression(expr.Target)
	// already reported error
	if typ == unresolved {
		return typ
	}

	switch typ := typ.(type) {
	case *types.FunctionSignature:
		// check if invocable
		fn := typ

		// Enum Switch Spread
		if c.lhsType != nil {
			return c.evaluateEnumDestructure(c.lhsType, fn, expr)
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
				err := c.resolveVar(expected, arg.Value, specializations)

				if err != nil {
					c.addError(err.Error(), arg.Range())
				}
			}
			return fn.Result.Type()
		}

		// Function is generic, instantiate
		for i, arg := range expr.Arguments {
			expected := fn.Parameters[i]
			err := c.resolveVar(expected, arg.Value, specializations)

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
			vT := c.evaluateExpression(arg.Value)
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

func (c *Checker) evaluateUnaryExpression(expr *ast.UnaryExpression) types.Type {
	op := expr.Op
	rhs := c.evaluateExpression(expr.Expr)
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

func (c *Checker) evaluateBinaryExpression(e *ast.BinaryExpression) types.Type {
	lhs := c.evaluateExpression(e.Left)
	rhs := c.evaluateExpression(e.Right)

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

func (c *Checker) evaluateAssignmentExpression(expr *ast.AssignmentExpression) types.Type {

	lhs := c.evaluateExpression(expr.Target)
	rhs := c.evaluateExpression(expr.Value)

	_, err := c.validate(lhs, rhs)

	if err != nil {
		c.addError(err.Error(), expr.Range())
	}

	// assignment yield void
	return types.LookUp(types.Void)
}

func (c *Checker) evaluateShorthandAssignmentExpression(expr *ast.ShorthandAssignmentExpression) types.Type {
	lhs := c.evaluateExpression(expr.Target)
	rhs := c.evaluateExpression(expr.Right)

	_, err := c.validate(lhs, rhs)

	if err != nil {
		c.addError(err.Error(), expr.Range())
	}

	// assignment yield void
	return types.LookUp(types.Void)
}

func (c *Checker) evaluateCompositeLiteral(n *ast.CompositeLiteral) types.Type {

	// 1 - Find Defined Type

	target := c.evaluateExpression(n.Target)
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
		err := c.resolveVar(f, v, specializations)

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

func (c *Checker) resolveVar(f *types.Var, v ast.Expression, specializations Specializations) error {
	vT := c.evaluateExpression(v)

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

func (c *Checker) evaluatePropertyExpression(n *ast.FieldAccessExpression) types.Type {

	a := c.evaluateExpression(n.Target)

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

func (c *Checker) evaluateGenericSpecializationExpression(e *ast.GenericSpecializationExpression) types.Type {
	// 1- Find Target
	sym, ok := c.find(e.Identifier.Value)
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
		args = append(args, c.evaluateTypeExpression(t, nil))
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

func (c *Checker) evaluateFunctionExpression(e *ast.FunctionExpression, self *types.DefinedType) types.Type {
	// Create new function

	sg := types.NewFunctionSignature()
	def := types.NewFunction(e.Identifier.Value, sg)
	c.table.DefineFunction(e, def)

	prevFn := c.fn
	prevSc := c.scope
	defer func() {
		c.fn = prevFn
		c.scope = prevSc
	}()

	c.fn = sg

	// Enter Function Scope
	c.enterScope()
	sg.Scope = c.scope
	c.table.AddScope(e, c.scope)
	defer c.leaveScope()

	// Type/Generic Parameters
	hasError := false
	if e.GenericParams != nil {
		for _, p := range e.GenericParams.Parameters {
			t := c.evaluateGenericParameterExpression(p)
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
		t := c.evaluateTypeExpression(p.Type, sg.TypeParameters)

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
		c.scope.Define(v)

	}

	// Annotated Return Type
	if e.ReturnType != nil {
		t := c.evaluateTypeExpression(e.ReturnType, sg.TypeParameters)
		sg.Result = types.NewVar("result", t)
	} else {
		sg.Result = types.NewVar("result", types.LookUp(types.Void))
	}

	// At this point the signature has been constructed fully, add to scope
	err := prevSc.Define(def)

	if err != nil {
		c.addError(err.Error(), e.Identifier.Range())
		return unresolved
	}

	// inject `self`
	if self != nil {
		s := types.NewVar("self", self)
		c.scope.Parent = self.GetScope()
		c.scope.Define(s)
	}

	// Body
	c.checkBlockStatement(e.Body)

	// TODO:
	// Ensure All Generic Params are used
	// Ensure All Params are used

	return sg
}

func (c *Checker) evaluateArrayLiteral(n *ast.ArrayLiteral) types.Type {

	var element types.Type

	if len(n.Elements) == 0 {
		c.addError("empty literal", n.Range())
		return unresolved
	}

	hasError := false
	for _, node := range n.Elements {
		provided := c.evaluateExpression(node)

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

	sym, ok := c.find("Array")

	if !ok {
		c.addError("unable to find map type", n.Range())
		return unresolved
	}
	typ := types.AsDefined(sym.Type())
	instantiated := types.Instantiate(typ, []types.Type{element}, nil)
	return instantiated
}
func (c *Checker) evaluateMapLiteral(n *ast.MapLiteral) types.Type {

	var key types.Type
	var value types.Type

	if len(n.Pairs) == 0 {
		c.addError("empty literal", n.Range())
		return unresolved
	}

	hasError := false
	for _, node := range n.Pairs {
		providedKey := c.evaluateExpression(node.Key)
		providedValue := c.evaluateExpression(node.Value)

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

	sym, ok := c.find("Map")

	if !ok {
		c.addError("unable to find map type", n.Range())
		return unresolved
	}
	typ := types.AsDefined(sym.Type())
	instantiated := types.Instantiate(typ, []types.Type{key, value}, nil)
	return instantiated
}

func (c *Checker) evaluateIndexExpression(n *ast.IndexExpression) types.Type {

	// 1 - Eval Target
	target := c.evaluateExpression(n.Target)

	// 2 - Eval Subscript Standard
	symbol, ok := c.find("SubscriptStandard")

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
	index := c.evaluateExpression(n.Index)

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

func (c *Checker) evaluateEnumDestructure(inc types.Type, fn *types.FunctionSignature, expr *ast.CallExpression) types.Type {
	lhsTyp := types.AsDefined(inc)

	if lhsTyp == nil {
		c.addError(fmt.Sprintf("expected defined type, got %s, %s", c.lhsType, fn), expr.Range())
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

		err := c.define(types.NewVar(ident.Value, t.Type()))

		if err != nil {
			c.addError(err.Error(), arg.Range())
			return unresolved
		}
	}

	return sg.Result.Type()
}
