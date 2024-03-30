package lirgen

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lir"
)

type builder struct {
	Mod *lir.Module
}

func build(mod *lir.Module) error {
	b := &builder{
		Mod: mod,
	}

	b.pass()
	return nil
}

func (b *builder) pass() {
	files := b.Mod.FileSet().Files
	passes := []func(*ast.File){
		b.pass0,
		b.pass1,
	}

	for _, fn := range passes {
		for _, file := range files {
			fn(file)
		}
	}
}
