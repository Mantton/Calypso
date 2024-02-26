package irgen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ssa"
	"github.com/mantton/calypso/internal/calypso/token"
	"github.com/mantton/calypso/internal/calypso/types"
	"tinygo.org/x/go-llvm"
)

type Locals map[ssa.Value]llvm.Value

type builder struct {
	*compiler
	llvm.Builder
	fn           *ssa.Function
	llvmFn       llvm.Value
	llvmFnType   llvm.Type
	locals       Locals
	blocks       map[*ssa.Block]llvm.BasicBlock
	currentBlock llvm.BasicBlock
}

func newBuilder(fn *ssa.Function, c *compiler, b llvm.Builder) *builder {
	// Initialize Function
	f, fnType := c.getFunction(fn.Symbol)
	return &builder{
		compiler:   c,
		fn:         fn,
		Builder:    b,
		locals:     make(Locals),
		blocks:     make(map[*ssa.Block]llvm.BasicBlock),
		llvmFn:     f,
		llvmFnType: fnType,
	}
}

func (b *builder) buildFunction() {
	fmt.Printf("[EMITTING FUNC] %s\n", b.fn.Symbol.Name())

	// Create Blocks
	var entry llvm.BasicBlock
	for i, block := range b.fn.Blocks {
		llvmBlock := b.context.AddBasicBlock(b.llvmFn, "")
		b.blocks[block] = llvmBlock

		if i == 0 {
			entry = llvmBlock
		}
	}

	b.SetInsertPointAtEnd(entry)

	// TODO: load params
	for i, v := range b.fn.Parameters {
		b.setValue(v, b.llvmFn.Param(i))
	}

	// Fill Blocks
	for _, ptr := range b.fn.Blocks {
		blk := b.blocks[ptr]
		b.SetInsertPointAtEnd(blk)
		b.currentBlock = blk
		for _, i := range ptr.Instructions {
			b.createInstruction(i)
		}
	}

}

func (b *builder) createInstruction(i ssa.Instruction) {
	switch i := i.(type) {
	case ssa.Value:
		val := b.createValue(i)
		b.setValue(i, val)

	case *ssa.Return:
		v := b.getValue(i.Result)
		b.CreateRet(v)

	case *ssa.Store:
		v := b.getValue(i.Value)
		a := b.getValue(i.Address)
		b.CreateStore(v, a)
	case *ssa.Branch:
		v := b.getValue(i.Condition)
		x := b.blocks[i.Action]
		y := b.blocks[i.Alternative]

		b.CreateCondBr(v, x, y)
	case *ssa.Jump:
		x := b.blocks[i.Block]
		b.CreateBr(x)

	default:

		panic(fmt.Sprintf("TODO: NOT IMPLEMENTED: %T", i))
	}

}

func (b *builder) createValue(v ssa.Value) llvm.Value {
	fmt.Printf("[EMIT VALUE] %T\n", v)

	switch v := v.(type) {
	case *ssa.Constant:
		return b.compiler.createConstant(v)
	case *ssa.Allocate:
		// TODO: Head/Stack
		typ := b.compiler.getType(v.Type())
		addr := b.CreateAlloca(typ, "")
		return addr
	// case *ssa.Call:
	// 	// TODO: Types

	// 	target := c.module.NamedFunction(v.Target)

	// 	// Function has not been declared
	// 	if target.IsNil() {
	// 		target = llvm.AddFunction(c.module, v.Target, llvm.FunctionType(c.context.Int32Type(), []llvm.Type{}, false))
	// 	}
	// 	t := llvm.FunctionType(c.context.Int32Type(), []llvm.Type{}, false)
	// 	val := c.builder.CreateCall(t, target, []llvm.Value{}, "")
	// 	return val
	case *ssa.Load:
		addr := b.getValue(v.Address)
		typ := b.compiler.getType(v.Address.Type())
		val := b.CreateLoad(typ, addr, "")
		return val
	case *ssa.Binary:

		lhs := b.getValue(v.Left)
		rhs := b.getValue(v.Right)
		op := v.Op
		typ := v.Left.Type()

		if typ == nil {
			fmt.Printf("%T", v.Left)
			panic("type is nil")
		}

		switch typ := typ.(type) {
		case *types.Basic:
			switch typ.Literal {

			case types.Bool:
				switch op {
				case token.EQL:
					return b.CreateICmp(llvm.IntEQ, lhs, rhs, "")
				case token.NEQ:
					return b.CreateICmp(llvm.IntNE, lhs, rhs, "")
				}

			case types.Int:
				switch op {
				case token.ADD:
					return b.CreateAdd(lhs, rhs, "")
				case token.SUB:
					return b.CreateSub(lhs, rhs, "")
				// Compare

				case token.LSS:
					return b.CreateICmp(llvm.IntSLT, lhs, rhs, "")
				case token.GTR:
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
	case *ssa.Call:
		lV, lT := b.getFunction(v.Target.Symbol)
		var lA []llvm.Value

		for _, p := range v.Arguments {
			lA = append(lA, b.getValue(p))
		}

		r := b.CreateCall(lT, lV, lA, "")
		return r

	default:
		panic("TODO: NOT IMPLMENTED")
	}
}

func (b *builder) getValue(v ssa.Value) llvm.Value {

	switch v := v.(type) {
	case *ssa.Constant:
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

func (b *builder) setValue(k ssa.Value, v llvm.Value) {
	b.locals[k] = v
}
