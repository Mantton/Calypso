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
	case *lir.Global:
		return b.compiler.createConstant(v.Value)
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
	case *lir.AND:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateAnd(lhs, rhs, "")
	case *lir.OR:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateOr(lhs, rhs, "")
	case *lir.XOR:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateXor(lhs, rhs, "")
	case *lir.ShiftLeft:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateShl(lhs, rhs, "")
	case *lir.ArithmeticShiftRight:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateAShr(lhs, rhs, "")
	case *lir.LogicalShiftRight:
		lhs, rhs := b.getValue(v.Left), b.getValue(v.Right)
		return b.CreateLShr(lhs, rhs, "")
	case *lir.Call:
		return b.createCall(v)
	case *lir.PHI:
		return b.createPhi(v)
	case *lir.Load:
		return b.createLoad(v)
	case *lir.Allocate:
		return b.createAlloc(v)
	case *lir.GEP:
		return b.createGEP(v)
	case *lir.ExtractValue:
		return b.createExtractValue(v)
	default:
		msg := fmt.Sprintf("[LLIRGEN] Value not implemented, %T", v)
		panic(msg)
	}
}

func (b *builder) createPhi(v *lir.PHI) llvm.Value {
	// Values
	phi_vals := []llvm.Value{}
	for _, n := range v.Nodes {
		phi_vals = append(phi_vals, b.getValue(n.Value))

	}

	// Blocks
	phi_blocks := []llvm.BasicBlock{}

	for _, n := range v.Nodes {
		phi_blocks = append(phi_blocks, b.blocks[n.Block])
	}

	typ := b.getType(v.Yields())
	phi := b.CreatePHI(typ, "")
	phi.AddIncoming(phi_vals, phi_blocks)
	return phi
}

func (b *builder) createLoad(v *lir.Load) llvm.Value {
	addr := b.getValue(v.Address)
	typ := b.compiler.getType(v.Yields())
	val := b.CreateLoad(typ, addr, "")
	return val
}

func (b *builder) createAlloc(v *lir.Allocate) llvm.Value {
	// TODO: Heap/Stack
	typ := b.compiler.getType(v.Yields())
	addr := b.CreateAlloca(typ, "")
	return addr
}

func (b *builder) createCall(v *lir.Call) llvm.Value {
	lV, lT := b.getFunction(v.Target)
	var lA []llvm.Value

	for _, p := range v.Arguments {
		lA = append(lA, b.getValue(p))
	}

	r := b.CreateCall(lT, lV, lA, "")
	return r
}

func (b *builder) createGEP(v *lir.GEP) llvm.Value {
	addr := b.getValue(v.Address)

	indices := []llvm.Value{
		llvm.ConstInt(b.context.Int32Type(), 0, false),
		llvm.ConstInt(b.context.Int32Type(), uint64(v.Index), false),
	}

	elemT := b.getType(v.Composite.Type.Parent())
	// b.module.Dump()
	return b.CreateInBoundsGEP(elemT, addr, indices, "")
}

func (b *builder) createExtractValue(v *lir.ExtractValue) llvm.Value {
	addr := b.getValue(v.Address)
	fmt.Println("ADDR:", addr)
	return b.CreateExtractValue(addr, v.Index, "")
}
