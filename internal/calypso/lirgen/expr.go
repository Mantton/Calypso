package lirgen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (b *builder) evaluateExpression(n ast.Expression, fn *lir.Function) lir.Value {
	switch e := n.(type) {
	case *ast.BooleanLiteral:
		return lir.NewConst(e.Value, types.LookUp(types.Bool))
	case *ast.StringLiteral:
		// TODO: Global Composites
		panic("string literals")
	case *ast.CharLiteral:
		return lir.NewConst(e.Value, types.LookUp(types.Char))
	case *ast.IntegerLiteral:
		typ := b.Mod.TypeTable().GetNodeType(n)
		if typ == nil {
			panic("uknown integer type")
		}
		return lir.NewConst(e.Value, typ)
	case *ast.NilLiteral:
		typ := b.Mod.TypeTable().GetNodeType(n)
		if typ == nil {
			panic("unknown nullptr type")
		}

		return lir.NewConst(0, typ)

	default:
		panic(fmt.Sprintf("unknown expr %T\n", e))
	}
}
