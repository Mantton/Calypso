package lirgen

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lir"
)

type builder struct {
	Mod       *lir.Module
	Functions map[*ast.FunctionExpression]*lir.Function
}

func build(mod *lir.Module) error {
	b := &builder{
		Mod:       mod,
		Functions: make(map[*ast.FunctionExpression]*lir.Function),
	}

	b.pass()
	// b.debugPrint()
	return nil
}

func (b *builder) pass() {
	files := b.Mod.FileSet().Files
	passes := []func(*ast.File){
		b.pass0,
		b.pass1,
		b.pass2,
		b.pass3,
		b.pass4,
	}

	for _, fn := range passes {
		for _, file := range files {
			fn(file)
		}
	}
}
