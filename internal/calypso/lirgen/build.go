package lirgen

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/types"
)

type builder struct {
	Mod            *lir.Module
	Functions      map[*ast.FunctionExpression]*lir.Function
	EnumFunctions  map[*types.EnumVariant]*lir.Function
	RFunctionEnums map[*lir.Function]*types.EnumVariant
	Refs           map[string]*lir.TypeRef
}

func build(mod *lir.Module) error {
	b := &builder{
		Mod:            mod,
		Functions:      make(map[*ast.FunctionExpression]*lir.Function),
		EnumFunctions:  make(map[*types.EnumVariant]*lir.Function),
		RFunctionEnums: make(map[*lir.Function]*types.EnumVariant),
		Refs:           make(map[string]*lir.TypeRef),
	}

	b.pass()
	b.debugPrint()
	return nil
}

func (b *builder) pass() {
	files := b.Mod.FileSet().Files
	passes := []func(*ast.File){
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
		b.Mod.Functions[fn.Name()] = fn
	}
}
