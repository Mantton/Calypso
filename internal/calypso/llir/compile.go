package llir

import (
	"fmt"
	"os"

	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/types"
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
	context    llvm.Context
	module     llvm.Module
	lirMod     *lir.Module
	typesTable map[types.Type]llvm.Type
}

func newCompiler(module *lir.Module) *compiler {
	c := &compiler{
		context:    llvm.NewContext(),
		typesTable: make(map[types.Type]llvm.Type),
	}

	c.module = c.context.NewModule(module.Name())
	c.lirMod = module

	return c
}

func (c *compiler) compileModule() {
	// builder
	b := c.context.NewBuilder()
	defer b.Dispose()

	// Structs
	for _, cm := range c.lirMod.Composites {
		c.buildComposite(cm)
	}

	// Declare Functions
	for _, fn := range c.lirMod.Functions {
		if fn.External {

		} else {
			b := newBuilder(fn, c, b)
			b.buildFunction()
		}
	}

	err := llvm.VerifyModule(c.module, llvm.ReturnStatusAction)

	if err != nil {
		c.module.Dump()
		fmt.Printf("\n\n[Validation Error]\n%s\n", err)
		os.Exit(1)
	}

	c.module.Dump()
	fmt.Println("Valid Module")

	// trg, err := llvm.GetTargetFromTriple(llvm.DefaultTargetTriple())
	// if err != nil {
	// 	panic(err)
	// }
	// c.module.SetTarget(trg.Description())

	// mt := trg.CreateTargetMachine(llvm.DefaultTargetTriple(), "", "", llvm.CodeGenLevelDefault, llvm.RelocDefault, llvm.CodeModelDefault)

	// pbo := llvm.NewPassBuilderOptions()
	// defer pbo.Dispose()

	// pm := llvm.NewPassManager()
	// mt.AddAnalysisPasses(pm)

	// err = c.module.RunPasses("default<Os>", mt, pbo)

	// if err != nil {
	// 	panic(err)
	// }

}

func (c *compiler) buildComposite(cm *lir.Composite) llvm.Type {
	members := []llvm.Type{}

	for _, t := range cm.Members {
		typ := c.getType(t)
		members = append(members, typ)
	}

	llvmType := c.context.StructCreateNamed(cm.Name)
	c.typesTable[cm.Actual] = llvmType
	llvmType.StructSetBody(members, false)
	return llvmType

}
