package typechecker

import (
	"fmt"
	"go/constant"
	"reflect"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) validate(expected types.Type, provided types.Type, expr ast.Expression) (types.Type, error) {
	fmt.Printf("Validating `%s`(provided) |> `%s`(expected)\n", provided, expected)

	_, isGeneric := expected.(*types.TypeParam)

	if isGeneric {
		// Check Constraints
		panic("generic specialization & checking not implemented")
	}

	match := reflect.TypeOf(expected) == reflect.TypeOf(provided)

	if !match {
		return nil, fmt.Errorf("expected `%s`, received `%s`", expected, provided)
	}

	switch expected := expected.(type) {
	case *types.Basic:
		return c.validateBasicTypes(expected, provided.(*types.Basic), expr)
	}
	return nil, fmt.Errorf("expected `%s`, received `%s`", expected, provided)
}

func (c *Checker) validateBasicTypes(expected *types.Basic, provided *types.Basic, expr ast.Expression) (types.Type, error) {

	if expected == types.LookUp(types.Any) {
		return expected, nil
	}

	// either side
	if types.IsGroupLiteral(provided) {
		switch {
		case provided.Literal == types.IntegerLiteral, types.IsNumeric(expected):
			fmt.Printf("[Validation] %T\n", expr)
			fmt.Println(expr)
			c.table.AddNode(expr, expected, constant.MakeBool(true), nil)
			return expected, nil
		case provided.Literal == types.FloatLiteral, types.IsFloatingPoint(expected):
			return expected, nil
		}
	}

	match := expected.Literal == provided.Literal
	if !match {
		return nil, fmt.Errorf("expected `%s`, received `%s`", expected, provided)
	}

	return expected, nil
}
