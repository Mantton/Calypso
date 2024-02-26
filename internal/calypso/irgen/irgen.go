package irgen

import (
	"github.com/mantton/calypso/internal/calypso/ssa"
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

	// Emit Functino Values

	c.module.SetTarget("arm64")

	c.module.Dump()
	// Verify Module
	err := llvm.VerifyModule(c.module, llvm.PrintMessageAction)

	if err != nil {
		panic(err)
	}

	return "CAL"
}

func (c *compiler) createConstant(n *ssa.Constant) llvm.Value {
	switch n := n.Value.(type) {
	case int:
		return llvm.ConstInt(c.context.Int64Type(), uint64(n), true)
	case float64:
		return llvm.ConstFloat(c.context.FloatType(), n)
	case bool:
		v := 0
		if n {
			v = 1
		}
		return llvm.ConstInt(c.context.Int1Type(), uint64(v), true)

	default:
		panic("Invalid compile-time constant")
	}
}
