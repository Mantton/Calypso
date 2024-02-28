package irgen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ssa"
	"github.com/mantton/calypso/internal/calypso/types"
	"tinygo.org/x/go-llvm"
)

type LLIR struct {
	Data map[string]string
}

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
	ssaMod  *ssa.Module
}

func newCompiler(module *ssa.Module) *compiler {
	c := &compiler{
		context: llvm.NewContext(),
	}

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
	builder := c.context.NewBuilder()
	defer builder.Dispose()

	// Globals

	// Declare Functions
	for _, fn := range c.ssaMod.Functions {
		b := newBuilder(fn, c, builder)
		b.buildFunction()
	}

	c.module.Dump()

	// Verify Module
	err := llvm.VerifyModule(c.module, llvm.ReturnStatusAction)

	if err != nil {
		panic(err)
	}

	trg, err := llvm.GetTargetFromTriple(llvm.DefaultTargetTriple())
	if err != nil {
		panic(err)
	}
	c.module.SetTarget(trg.Description())

	mt := trg.CreateTargetMachine(llvm.DefaultTargetTriple(), "", "", llvm.CodeGenLevelDefault, llvm.RelocDefault, llvm.CodeModelDefault)

	pbo := llvm.NewPassBuilderOptions()
	defer pbo.Dispose()

	pm := llvm.NewPassManager()
	mt.AddAnalysisPasses(pm)

	err = c.module.RunPasses("default<Os>", mt, pbo)

	if err != nil {
		panic(err)
	}

	fmt.Println("\n----------- OPTIMIZED")
	c.module.Dump()

	return "CAL"
}

func (c *compiler) createConstant(n *ssa.Constant) llvm.Value {
	t := n.Type().(*types.Basic)

	if t == nil {
		panic("invalid compile-time constant")
	}

	switch t.Literal {
	case types.Int, types.Int64, types.IntegerLiteral, types.UInt, types.UInt64:
		v := n.Value.(int64)
		return llvm.ConstInt(c.context.Int64Type(), uint64(v), true)
	case types.Char, types.Int32, types.UInt32:
		v := n.Value.(int64)
		return llvm.ConstInt(c.context.Int32Type(), uint64(v), true)
	case types.Int16, types.UInt16:
		v := n.Value.(int64)
		return llvm.ConstInt(c.context.Int16Type(), uint64(v), true)
	case types.Int8, types.UInt8:
		v := n.Value.(int64)
		return llvm.ConstInt(c.context.Int8Type(), uint64(v), true)
	case types.Bool:
		v := n.Value.(bool)
		o := 0
		if v {
			o = 1
		}
		return llvm.ConstInt(c.context.Int1Type(), uint64(o), true)
	case types.NilLiteral:
		panic("unreachable")
	default:
		panic("constant type has not been defined yet")
	}
}
