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
	case *ast.FunctionCallExpression:
		c.checkCallExpression(expr)
	default:
		msg := fmt.Sprintf("expression check not implemented, %T", expr)
		panic(msg)
	}
}

func (c *Checker) checkFunctionExpression(e *ast.FunctionExpression) {
	def := types.NewFunction(e.Identifier.Value, nil)
	def.SetType(unresolved)
	ok := c.define(def)

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is already defined", def.Name()),
			e.Identifier.Range(),
		)
		return
	}
	sg := types.NewFunctionSignature()
	def.SetType(sg)
	t := c.evaluateFunctionExpression(e, nil, sg)
	def.SetType(t)

	c.table.DefineFunction(e, def)
}

func (c *Checker) checkAssignmentExpression(expr *ast.AssignmentExpression) {
	// TODO: mutability checks
	c.evaluateAssignmentExpression(expr)
}

func (c *Checker) checkCallExpression(expr *ast.FunctionCallExpression) {

	retType := c.evaluateCallExpression(expr)

	if retType != types.LookUp(types.Void) {
		fmt.Println("[WARNING] Call Expression returning non void value is unused")
	}

}

// ----------- Eval ------------------
func (c *Checker) evaluateExpression(expr ast.Expression) types.Type {
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
	case *ast.FunctionCallExpression:
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

	// case *ast.ArrayLiteral:
	// case *ast.MapLiteral:

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

func (c *Checker) evaluateCallExpression(expr *ast.FunctionCallExpression) types.Type {
	typ := c.evaluateExpression(expr.Target)
	switch fn := typ.(type) {
	case *types.FunctionSignature:

		if c.lhsType != nil {
			lhsTyp, ok := c.lhsType.(*types.DefinedType)

			if !ok {
				c.addError(fmt.Sprintf("expected defined type, got %s, %s", c.lhsType, fn), expr.Range())
				return unresolved
			}

			specializations := make(map[string]types.Type)

			for _, p := range lhsTyp.TypeParameters {
				specializations[p.Name()] = p.Bound
			}

			sg := apply(specializations, fn).(*types.FunctionSignature)

			fmt.Println("Instantiated: ", sg)

			if len(sg.Parameters) != len(expr.Arguments) {
				c.addError(fmt.Sprintf("expected %d variables, got %d", len(sg.Parameters), len(fn.Parameters)), expr.Range())
				return unresolved
			}

			for i, t := range sg.Parameters {
				arg := expr.Arguments[i]
				ident, ok := arg.(*ast.IdentifierExpression)

				if !ok {
					c.addError("expected identifier", arg.Range())
					return unresolved
				}

				ok = c.define(types.NewVar(ident.Value, t.Type()))

				if !ok {
					c.addError(fmt.Sprintf("\"%s\" already exists in current context", ident.Value), arg.Range())
					return unresolved
				}
			}

			return sg.Result.Type()
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
				err := c.resolveVar(expected, arg, specializations)

				if err != nil {
					c.addError(err.Error(), arg.Range())
				}
			}
			return fn.Result.Type()
		}

		// Function is generic, instantiate
		for i, arg := range expr.Arguments {
			expected := fn.Parameters[i]
			err := c.resolveVar(expected, arg, specializations)

			if err != nil {
				c.addError(err.Error(), arg.Range())
			}
		}

		t := apply(specializations, fn)
		fmt.Println("\nInstantiated Function:", t)
		fmt.Println("Original:", fn)
		fmt.Println()
		return t.(*types.FunctionSignature).Result.Type()
	}

	// already reported error
	if typ == unresolved {
		return typ
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

	case token.SUB:

		// if RHS is numeric, return RHS type as numberic types can be negated
		if types.IsNumeric(rhs) {
			return rhs
		}

		err = fmt.Errorf("unsupported negation operand on type `%s`", rhs)
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
	op := e.Op

	typ, err := c.validate(lhs, rhs)

	if err != nil {
		c.addError(err.Error(), e.Range())
		return unresolved
	}

	switch op {
	case token.ADD, token.SUB, token.QUO, token.MUL:
		if types.IsNumeric(typ) {
			return typ
		}
	case token.LSS, token.GTR, token.LEQ, token.GEQ:
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

func (c *Checker) evaluateCompositeLiteral(n *ast.CompositeLiteral) types.Type {

	// 1 - Find Defined Type
	name := n.Identifier.Value

	sym, ok := c.find(name)

	if !ok {
		if !ok {
			c.addError(
				fmt.Sprintf("`%s` is not defined", n.Identifier.Value),
				n.Range(),
			)

			return unresolved
		}
	}

	base, ok := sym.(*types.DefinedType)

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is not a type", n.Identifier.Value),
			n.Identifier.Range(),
		)
		return unresolved
	}

	// 2 - Ensure Defined Type is Struct
	if !types.IsStruct(base.Parent()) {
		c.addError(
			fmt.Sprintf("`%s` is not a struct", n.Identifier.Value),
			n.Identifier.Range(),
		)
		return unresolved
	}

	// 3 - Get Struct Signature of Defined Type

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

		_, ok = seen[k]

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
		_, ok := specializations[param.Name()]

		if !ok {
			panic(fmt.Errorf("failed to resolve generic parameter: %s in %s", param, specializations))
		}
	}

	return apply(specializations, base)
}

func (c *Checker) resolveVar(f *types.Var, v ast.Expression, specializations map[string]types.Type) error {
	vT := c.evaluateExpression(v)

	if vT == unresolved {
		return fmt.Errorf("unresolved type assigned for `%s`", f.Name())
	}

	// check constraints & specialize
	// can either be a type param or generic struct or a generic function

	switch gT := f.Type().(type) {
	case *types.TypeParam:
		alt, ok := specializations[gT.Name()]

		// has not already been specialized
		if !ok {

			// Ensure Conformance
			err := c.validateConformance(gT.Constraints, vT)
			if err != nil {
				return err
			}

			specializations[gT.Name()] = vT
			fmt.Println("Specialized Type Param", gT, ":", vT)

			return nil
		}

		// has been specialized, ensure strict match
		temp := types.NewVar("", alt)
		err := c.validateAssignment(temp, vT, v)

		if err != nil {
			return err
		}

		// no errors, type match
	case *types.FunctionSignature:
		panic("not implemented")
	case *types.DefinedType:
		if !types.IsGeneric(gT) {
			break
		}

		prev := vT
		vT := types.AsDefined(vT)

		if vT == nil {
			panic(fmt.Errorf("type is not a defined type : %T (%s)", prev, prev))
		}

		// TODO -> These two have to match
		for i, x := range gT.TypeParameters {
			specializations[x.Name()] = vT.TypeParameters[i]
		}

		fmt.Println("Specialized", gT, "as", vT, "with", specializations)
		return nil

	case *types.Pointer:
		base := gT.PointerTo

		// vT must be an instantiated struct of type fT
		iT, ok := vT.(*types.Pointer)

		if !ok {
			return fmt.Errorf("expected pointer to type %s, received %s", gT, vT)
		}
		alt, ok := specializations[base.String()]

		// has not already been specialized
		if !ok {
			fmt.Println("[DEBUG]specializing ptr type", base, ":", iT.PointerTo)
			specializations[base.String()] = iT.PointerTo
			return nil
		}

		// has been specialized, ensure strict match
		temp := types.NewVar("", alt)
		err := c.validateAssignment(temp, iT.PointerTo, v)

		if err != nil {
			return err
		}

	}

	err := c.validateAssignment(f, vT, v)
	if err != nil {
		return err
	}
	return nil
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

		// Access Field
		switch parent := a.Parent().(type) {
		case *types.Struct:
			for _, f := range parent.Fields {
				if field == f.Name() {
					return f.Type()
				}
			}

		case *types.Enum:
			for _, v := range parent.Variants {
				if v.Name == field {
					// Not tuple type, return parent type
					if len(v.Fields) == 0 {
						return a
					}

					// Tuple Type, Return Function Returning Parent Type
					sg := types.NewFunctionSignature()
					for _, p := range v.Fields {
						sg.AddParameter(p)
					}

					sg.Result.SetType(a)
					return sg
				}
			}
		default:
			fmt.Println(a)
			panic("TODO")
		}

		// Access Method
		method, ok := a.Methods[field]
		if ok {
			return method.Sg()
		}

	default:
		panic("TODO")
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

		err := c.validateConformance(param.Constraints, arg)

		if err != nil {
			c.addError(err.Error(), e.Clause.Arguments[i].Range())
			hasError = true
		}
	}

	if hasError {
		return unresolved
	}

	// 7 - Return instance of type
	instance := instantiate(sym.Type(), args, nil)
	return instance
}

func (c *Checker) evaluateFunctionExpression(e *ast.FunctionExpression, self *types.DefinedType, sg *types.FunctionSignature) types.Type {

	// Create new Signature
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

	// inject `self`
	if self != nil {
		s := types.NewVar("self", self)
		c.scope.Parent = self.Scope
		c.scope.Define(s)
	}

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
	for _, p := range e.Params {
		t := c.evaluateTypeExpression(p.AnnotatedType, sg.TypeParameters)
		v := types.NewVar(p.Value, t)
		c.scope.Define(v)
		sg.AddParameter(v)
	}

	// Annotated Return Type
	if e.ReturnType != nil {
		t := c.evaluateTypeExpression(e.ReturnType, sg.TypeParameters)
		sg.Result = types.NewVar("result", t)
	} else {
		sg.Result = types.NewVar("result", types.LookUp(types.Void))
	}

	// Body
	c.checkBlockStatement(e.Body)

	// TODO:
	// Ensure All Generic Params are used
	// Ensure All Params are used

	return sg
}
