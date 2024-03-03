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
	default:
		msg := fmt.Sprintf("expression check not implemented, %T", expr)
		panic(msg)
	}
}

func (c *Checker) checkFunctionExpression(e *ast.FunctionExpression) {

	prevFn := c.fn
	defer func() {
		c.fn = prevFn
	}()

	sg := types.NewFunctionSignature()
	def := types.NewFunction(e.Identifier.Value, sg)
	ok := c.define(def)
	defer func() {
		c.table.DefineFunction(e, def)
	}()

	// set current checking function to sg
	c.fn = sg

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is already defined", def.Name()),
			e.Identifier.Range(),
		)
		return
	}

	c.enterScope()
	sg.Scope = c.scope
	c.table.AddScope(e, c.scope)
	defer c.leaveScope()

	hasError := false
	// Type/Generic Parameters
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
		return
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

	// Ensure All Generic Params are used
	// Ensure All Params are used
}

func (c *Checker) checkAssignmentExpression(expr *ast.AssignmentExpression) {
	// TODO: mutability checks
	c.evaluateAssignmentExpression(expr)
}

func (c *Checker) checkCallExpression(expr *ast.CallExpression) {

	retType := c.evaluateCallExpression(expr)

	if retType.Parent() != types.LookUp(types.Void) {
		fmt.Println("Add Warning here")

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

func (c *Checker) evaluateCallExpression(expr *ast.CallExpression) types.Type {
	t := c.evaluateExpression(expr.Target)

	fn, ok := t.(*types.FunctionSignature)

	// Ensure Target is callable
	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is not invocable", expr.Target),
			expr.Target.Range(),
		)
		return unresolved
	}

	// Guard Argument Count == Parameter Count
	if len(expr.Arguments) != len(fn.Parameters) {
		c.addError(
			fmt.Sprintf("expected %d arguments, provided %d",
				len(fn.Parameters),
				len(expr.Arguments)),
			expr.Range(),
		)
		return fn.Result.Type()
	}

	// Check Arguments
	specializations := make(map[types.Type]types.Type)
	hasError := false

	for i, arg := range expr.Arguments {
		expected := fn.Parameters[i]
		err := c.resolveVar(expected, arg, specializations)

		if err != nil {
			hasError = true
			c.addError(err.Error(), arg.Range())
		}
	}

	if hasError {
		return unresolved
	}

	if len(fn.TypeParameters) == 0 {
		return fn.Result.Type()
	}

	fmt.Println("specializations :", specializations)

	var args []types.Type

	for _, param := range fn.TypeParameters {
		v, ok := specializations[param]

		if !ok {
			panic("failed to resolve generic parameter")
		}

		args = append(args, v)
	}

	if types.IsGeneric(fn.Result.Type()) {
		// Record Function Instance
		x := types.NewFunctionInstance(fn, args)
		c.table.SetFunctionInstance(expr, x)
		return specializations[fn.Result.Type()]
	} else {
		return fn.Result.Type()
	}
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
	err = fmt.Errorf("unsupported binary operand `%s`", token.LookUp(op))
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
	for _, p := range n.Pairs {
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

	specializations := make(map[types.Type]types.Type)

	for k, v := range seen {

		f := sg.FindField(k)
		err := c.resolveVar(f, v, specializations)

		if err != nil {
			c.addError(err.Error(), v.Range())
			hasError = true
		}
	}

	fmt.Println("specs", specializations)
	if hasError {
		return unresolved
	}

	if len(base.TypeParameters) == 0 {
		return base
	}

	var args []types.Type

	for _, param := range base.TypeParameters {
		v, ok := specializations[param]

		if !ok {
			panic("failed to resolve generic parameter")
		}

		args = append(args, v)
	}

	return types.NewStructInstance(base, args)
}

func (c *Checker) resolveVar(f *types.Var, v ast.Expression, specializations map[types.Type]types.Type) error {
	fT := f.Type()
	vT := c.evaluateExpression(v)
	if !types.IsGeneric(fT) {
		fmt.Println("Skipping non generic", fT)
		err := c.validateAssignment(f, vT, v)
		if err != nil {
			return err
		}
		return nil
	}

	// check constraints & specialize
	// can either be a type param or generic struct or a generic function

	switch gT := fT.(type) {
	case *types.TypeParam:
		alt, ok := specializations[gT]

		// has not already been specialized
		if !ok {
			fmt.Println("specializing", gT, ":", vT)

			specializations[gT] = vT
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
		panic("bad path")
	case *types.StructInstance:
		// vT must be an instantiated struct of type fT
		iT, ok := vT.(*types.StructInstance)
		if !ok || iT.Type != gT.Type {
			return fmt.Errorf("expected instance of type %s, received %s", gT, vT)
		}

		for i, a := range iT.TypeArgs {
			p := gT.TypeArgs[i]
			if !types.IsGeneric(p) {
				continue
			}

			alt, ok := specializations[p]

			if !ok {
				specializations[p] = a
			} else {
				temp := types.NewVar("", alt)

				err := c.validateAssignment(temp, a, v)
				if err != nil {
					return err
				}

			}
		}

	case *types.Pointer:
		base := gT.PointerTo

		// vT must be an instantiated struct of type fT
		iT, ok := vT.(*types.Pointer)

		if !ok {
			return fmt.Errorf("expected pointer to type %s, received %s", gT, vT)
		}
		alt, ok := specializations[base]

		// has not already been specialized
		if !ok {
			fmt.Println("specializing", base, ":", iT.PointerTo)

			specializations[base] = iT.PointerTo
			return nil
		}

		// has been specialized, ensure strict match
		temp := types.NewVar("", alt)
		err := c.validateAssignment(temp, iT.PointerTo, v)

		if err != nil {

			return err
		}

	default:
		panic("invalid type with generic param")
	}

	return nil
}
