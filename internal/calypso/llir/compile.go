package llir

import (
	"github.com/mantton/calypso/internal/calypso/lir"
	"tinygo.org/x/go-llvm"
)

func init() {
	llvm.InitializeAllTargets()
	llvm.InitializeAllTargetMCs()
	llvm.InitializeAllTargetInfos()
	llvm.InitializeAllAsmParsers()
	llvm.InitializeAllAsmPrinters()
}

func Compile(s *lir.Executable) {

	for _, mod := range s.Modules {
		c := newCompiler(mod)
		c.compileModule()
	}
}

type compiler struct {
	context llvm.Context
	module  llvm.Module
	lirMod  *lir.Module
}

func newCompiler(module *lir.Module) *compiler {
	c := &compiler{
		context: llvm.NewContext(),
	}

	c.module = c.context.NewModule(module.Name())
	c.lirMod = module

	return c
}

func (c *compiler) compileModule() {
	// builder
	b := c.context.NewBuilder()
	defer b.Dispose()

	// Declare Functions
	for _, fn := range c.lirMod.Functions {
		b := newBuilder(fn, c, b)
		b.buildFunction()
	}

	c.module.Dump()

}
