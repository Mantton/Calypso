package lirgen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/types"
)

type builder struct {
	Mod            *lir.Module
	Functions      map[*ast.FunctionExpression]*lir.Function // maps fn expressions to their lir impl
	TFunctions     map[types.Type]*lir.Function              // maps fn types to their lir impl
	EnumFunctions  map[*types.EnumVariant]*lir.Function
	RFunctionEnums map[*lir.Function]*types.EnumVariant
	MP             *lir.Executable
	main           *lir.Function
}

func build(mod *lir.Module, mp *lir.Executable) error {
	b := &builder{
		Mod:            mod,
		Functions:      make(map[*ast.FunctionExpression]*lir.Function),
		TFunctions:     make(map[types.Type]*lir.Function),
		EnumFunctions:  make(map[*types.EnumVariant]*lir.Function),
		RFunctionEnums: make(map[*lir.Function]*types.EnumVariant),
		MP:             mp,
	}

	b.pass()
	b.debugPrint()
	return nil
}

func (b *builder) pass() {
	files := b.Mod.FileSet().Files
	passes := []func(*ast.File){
		b.passN,
		b.pass0,
		b.pass1,
		b.pass2,
		b.pass3,
	}

	for _, fn := range passes {
		for _, file := range files {
			fn(file)
		}
	}

	for _, fn := range b.EnumFunctions {
		b.Mod.Functions[fn.Name] = fn
	}

	b.mono()

	if b.Mod.IsMainTarget() {
		b.entry()
	}
}

func (b *builder) entry() {
	// Build Function
	sg := types.NewFunctionSignature()
	sg.Result.SetType(types.LookUp(types.Int8))
	tfn := types.NewFunction("main", sg, nil)
	tfn.Target = &types.FunctionTarget{
		Target: "calypso",
	}
	fn := lir.NewFunction(tfn)
	fn.Name = "main"
	b.Mod.Functions["main"] = fn

	if b.main != nil {
		fn.Emit(&lir.Call{
			Target: b.main,
		})
	}

	fn.Emit(&lir.Return{
		Result: lir.NewConst(int64(0), types.LookUp(types.Int8)),
	})

	fmt.Println("done", b.Mod.Name())
}
