package llir

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/token"
	"github.com/mantton/calypso/internal/calypso/types"
	"tinygo.org/x/go-llvm"
)

func (b *builder) getValue(v lir.Value) llvm.Value {
	switch v := v.(type) {
	case *lir.Constant:
		return b.createConstant(v)
	default:
		lV, ok := b.locals[v]

		if !ok {
			m := fmt.Sprintf("Val not found: %T", v)
			panic(m)
		}

		return lV
	}
}

func (b *builder) setValue(k lir.Value, v llvm.Value) {
	b.locals[k] = v
}

func (b *builder) createValue(v lir.Value) llvm.Value {
	switch v := v.(type) {
	case *lir.Constant:
		return b.compiler.createConstant(v)
	case *lir.Allocate:
		// TODO: Head/Stack
		typ := b.compiler.getType(v.Yields())
		addr := b.CreateAlloca(typ, "")
		return addr
	case *lir.Load:
		addr := b.getValue(v.Address)
		typ := b.compiler.getType(v.Address.Yields())
		val := b.CreateLoad(typ, addr, "")
		return val
	case *lir.Binary:

		lhs := b.getValue(v.Left)
		rhs := b.getValue(v.Right)
		op := v.Op
		typ := v.Left.Yields()

		if typ == nil {
			fmt.Printf("%T", v.Left)
			panic("type is nil")
		}

		switch typ := typ.Parent().(type) {
		case *types.Basic:
			switch typ.Literal {

			case types.Bool:
				switch op {
				case token.EQL:
					return b.CreateICmp(llvm.IntEQ, lhs, rhs, "")
				case token.NEQ:
					return b.CreateICmp(llvm.IntNE, lhs, rhs, "")
				}

			case types.Int, types.IntegerLiteral, types.Int64, types.Int32, types.Int16, types.Int8:
				switch op {
				case token.PLUS:
					return b.CreateAdd(lhs, rhs, "")
				case token.MINUS:
					return b.CreateSub(lhs, rhs, "")
				// Compare

				case token.L_CHEVRON:
					return b.CreateICmp(llvm.IntSLT, lhs, rhs, "")
				case token.R_CHEVRON:
					return b.CreateICmp(llvm.IntSGT, lhs, rhs, "")
				case token.GEQ:
					return b.CreateICmp(llvm.IntSGE, lhs, rhs, "")
				case token.LEQ:
					return b.CreateICmp(llvm.IntSLE, lhs, rhs, "")
				case token.EQL:
					return b.CreateICmp(llvm.IntEQ, lhs, rhs, "")
				case token.NEQ:
					return b.CreateICmp(llvm.IntNE, lhs, rhs, "")
				}
			}

		}

		fmt.Println(token.LookUp(op), typ)
		panic("not ready")
	case *lir.Call:
		lV, lT := b.getFunction(v.Target.Type)
		var lA []llvm.Value

		for _, p := range v.Arguments {
			lA = append(lA, b.getValue(p))
		}

		r := b.CreateCall(lT, lV, lA, "")
		return r

	default:
		msg := fmt.Sprintf("[LIRGEN] Value not implemented, %T", v)
		panic(msg)
	}
}
