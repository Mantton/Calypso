package irgen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ssa"
	"tinygo.org/x/go-llvm"
)

type LLIR struct {
	Data map[string]string
}

type Locals map[ssa.Value]llvm.Value

func init() {
	llvm.InitializeAllTargets()
	llvm.InitializeAllTargetMCs()
	llvm.InitializeAllTargetInfos()
	llvm.InitializeAllAsmParsers()
	llvm.InitializeAllAsmPrinters()
}

type compiler struct {
	context llvm.Context
	module  llvm.Module
	builder llvm.Builder
	ssaMod  *ssa.Module
}

func newCompiler(module *ssa.Module) *compiler {
	c := &compiler{
		context: llvm.NewContext(),
	}

	c.builder = c.context.NewBuilder()
	c.module = c.context.NewModule(module.Name)
	c.ssaMod = module

	return c
}

func Compile(s *ssa.Executable) *LLIR {
	ir := &LLIR{
		Data: make(map[string]string),
	}

	for _, mod := range s.Modules {
		ir.Data[mod.Name] = compileModule(mod)
	}

	return ir
}

func compileModule(m *ssa.Module) string {
	c := newCompiler(m)
	return c.Compile()
}

func (c *compiler) Compile() string {
	defer c.builder.Dispose()

	// Globals

	// Functions
	for _, fn := range c.ssaMod.Functions {
		c.emitFn(fn)
	}

	c.module.Dump()
	return "CAL"
}

func (c *compiler) emitFn(fn *ssa.Function) {
	fmt.Printf("[EMITTING FUNC] %s\n", fn.Symbol.Name)

	// Create LLVM Fn
	returnType := c.context.Int32Type()
	paramTypes := []llvm.Type{}
	fnType := llvm.FunctionType(returnType, paramTypes, false)

	// Params
	f := llvm.AddFunction(c.module, fn.Symbol.Name, fnType)
	locals := make(Locals)
	// Blocks

	for _, b := range fn.Blocks {
		c.emitBlock(f, b, locals)
	}

}

func (c *compiler) emitBlock(fn llvm.Value, b *ssa.Block, l Locals) {
	fmt.Printf("[EMITTING BLOCK] %d\n", b.Index)

	blk := llvm.AddBasicBlock(fn, "")
	c.builder.SetInsertPointAtEnd(blk)

	for _, i := range b.Instructions {
		c.emitInstruction(fn, blk, i, l)
	}
}

func (c *compiler) emitInstruction(fn llvm.Value, blk llvm.BasicBlock, i ssa.Instruction, l Locals) {
	fmt.Printf("[EMIT INSTR] %T\n", i)
	switch i := i.(type) {
	case ssa.Value:
		val := c.emitValue(fn, blk, i, l)
		l[i] = val // Add to Locals

	case *ssa.Return:
		v := c.emitValue(fn, blk, i.Result, l)
		c.builder.CreateRet(v)

	case *ssa.Store:
		v := c.emitValue(fn, blk, i.Value, l)
		a := l[i.Address]
		c.builder.CreateStore(v, a)

	default:
		panic("TODO: NOT IMPLMENTED")
	}

}

func (c *compiler) emitValue(fn llvm.Value, blk llvm.BasicBlock, v ssa.Value, l Locals) llvm.Value {
	fmt.Printf("[EMIT VALUE] %T\n", v)

	switch v := v.(type) {
	case *ssa.Constant:
		val := v.Value.(int)
		return llvm.ConstInt(c.context.Int32Type(), uint64(val), true)
	case *ssa.Allocate:
		// TODO: Head/Stack
		// TODO: Types
		addr := c.builder.CreateAlloca(c.context.Int32Type(), "")
		return addr

	case *ssa.Call:
		// TODO: Types

		target := c.module.NamedFunction(v.Target)

		// Function has not been declared
		if target.IsNil() {
			target = llvm.AddFunction(c.module, v.Target, llvm.FunctionType(c.context.Int32Type(), []llvm.Type{}, false))
		}
		t := llvm.FunctionType(c.context.Int32Type(), []llvm.Type{}, false)
		val := c.builder.CreateCall(t, target, []llvm.Value{}, "")
		return val
	case *ssa.Load:
		// TODO: Types
		addr := l[v.Address]
		val := c.builder.CreateLoad(c.context.Int32Type(), addr, "")
		return val

	default:
		panic("TODO: NOT IMPLMENTED")
	}
}
