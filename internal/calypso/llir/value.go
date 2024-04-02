package llir

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/lir"
	"tinygo.org/x/go-llvm"
)

func (b *builder) getValue(v lir.Value) llvm.Value {
	switch v := v.(type) {
	case *lir.Constant:
		return b.createConstant(v)
	default:
		lV, ok := b.locals[v]

		if !ok {
			lV = b.createValue(v)
			b.setValue(v, lV)
			return lV
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
		// TODO: Heap/Stack
		typ := b.compiler.getType(v.Yields())
		addr := b.CreateAlloca(typ, "")
		return addr
	case *lir.Load:
		addr := b.getValue(v.Address)
		typ := b.compiler.getType(v.Address.Yields())
		val := b.CreateLoad(typ, addr, "")
		return val
	case *lir.Add:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateAdd(lhs, rhs, "")
	case *lir.FAdd:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateFAdd(lhs, rhs, "")
	case *lir.Sub:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateSub(lhs, rhs, "")
	case *lir.FSub:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateFSub(lhs, rhs, "")
	case *lir.Mul:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateMul(lhs, rhs, "")
	case *lir.FMul:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateFMul(lhs, rhs, "")
	case *lir.UDiv:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateUDiv(lhs, rhs, "")
	case *lir.SDiv:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateSDiv(lhs, rhs, "")
	case *lir.FDiv:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateFDiv(lhs, rhs, "")
	case *lir.URem:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateURem(lhs, rhs, "")
	case *lir.SRem:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateSRem(lhs, rhs, "")
	case *lir.FRem:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateFRem(lhs, rhs, "")
	case *lir.INeg:
		rhs := b.getValue(v.Right)
		return b.CreateNeg(rhs, "")
	case *lir.FNeg:
		rhs := b.getValue(v.Right)
		return b.CreateFNeg(rhs, "")

	case *lir.ICmp:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		p := IntPredicateMap[v.Comparison]
		return b.CreateICmp(p, lhs, rhs, "")
	case *lir.XOR:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateXor(lhs, rhs, "")
	case *lir.Call:
		lV, lT := b.getFunction(v.Target.Type)
		var lA []llvm.Value

		for _, p := range v.Arguments {
			lA = append(lA, b.getValue(p))
		}

		r := b.CreateCall(lT, lV, lA, "")
		return r

	default:
		msg := fmt.Sprintf("[LLIRGEN] Value not implemented, %T", v)
		panic(msg)
	}
}
