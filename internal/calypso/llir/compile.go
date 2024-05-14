package llir

import (
	"os"
	"os/exec"

	"github.com/mantton/calypso/internal/calypso/lexer"
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

type GCompiler struct {
	modules map[int64]llvm.Module
}

func Compile(s *lir.Executable) error {

	gc := &GCompiler{
		modules: make(map[int64]llvm.Module),
	}
	errs := []error{}
	ctx := llvm.NewContext()

	files := []string{}

	// Generate LLVM Modules
	for _, pkg := range s.Packages {
		for _, mod := range pkg.Modules {
			c := newCompiler(mod, s, ctx)
			lMod, err := c.compileModule()
			if err != nil {
				errs = append(errs, err)
			}
			gc.modules[mod.ID()] = lMod
		}
	}

	if len(errs) != 0 {
		return lexer.CombinedErrors(errs)
	}

	// Combine
	for _, pkg := range s.Packages {

		base := ctx.NewModule(pkg.AST.Name())

		for _, mod := range pkg.Modules {
			llvmMod := gc.modules[mod.ID()]
			err := llvm.LinkModules(base, llvmMod)

			if err != nil {
				return err
			}

		}

		err := llvm.VerifyModule(base, llvm.ReturnStatusAction)

		if err != nil {
			return err
		}

		trg, err := llvm.GetTargetFromTriple(llvm.DefaultTargetTriple())
		if err != nil {
			errs = append(errs, err)
		}
		base.SetTarget(trg.Description())

		// mt := trg.CreateTargetMachine(llvm.DefaultTargetTriple(), "", "", llvm.CodeGenLevelDefault, llvm.RelocDefault, llvm.CodeModelDefault)

		// pbo := llvm.NewPassBuilderOptions()
		// defer pbo.Dispose()

		// pm := llvm.NewPassManager()
		// mt.AddAnalysisPasses(pm)

		// err = base.RunPasses("default<Os>", mt, pbo)

		// if err != nil {
		// 	return err
		// }

		// fmt.Println("\n\n")
		// base.Dump()

		f, err := os.Create("./bin/" + pkg.AST.Name() + ".bc")
		if err != nil {
			return err
		}

		err = llvm.WriteBitcodeToFile(base, f)

		if err != nil {
			return err
		}

		files = append(files, f.Name())
	}

	// Link Packages
	combined, err := os.Create("./bin/combined.bc")

	if err != nil {
		return err
	}

	args := []string{}
	args = append(args, files...)
	args = append(args, "-o")
	args = append(args, combined.Name())

	cmd := exec.Command("llvm-link", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	if err != nil {
		return err
	}

	// Create Executable
	cmd = exec.Command("clang", combined.Name(), "-o", "./bin/exec")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	if err != nil {
		return err
	}

	return nil
}

type compiler struct {
	context    llvm.Context
	module     llvm.Module
	lirMod     *lir.Module
	exec       *lir.Executable
	typesTable map[types.Type]llvm.Type
}

func newCompiler(module *lir.Module, exec *lir.Executable, ctx llvm.Context) *compiler {
	c := &compiler{
		context:    ctx,
		typesTable: make(map[types.Type]llvm.Type),
		exec:       exec,
	}

	c.module = c.context.NewModule(module.Name())
	c.lirMod = module

	return c
}

func (c *compiler) compileModule() (llvm.Module, error) {
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
		return llvm.Module{}, err
	}

	return c.module, nil
}

func (c *compiler) buildComposite(cm *lir.Composite) llvm.Type {
	members := []llvm.Type{}

	for _, t := range cm.Members {
		typ := c.getType(t)
		members = append(members, typ)
	}

	llvmType := c.context.StructCreateNamed(cm.Name)
	c.typesTable[cm.Type] = llvmType
	llvmType.StructSetBody(members, false)
	return llvmType

}
