package lirgen

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lir"
)

func (b *builder) passN(f *ast.File) {

	for _, d := range f.Nodes.Imports {
		key := d.PopulatedImportKey

		mod := b.MP.Modules[key]

		// Define in scope
		name := mod.Name()
		if d.Alias != nil {
			name = d.Alias.Value
		}
		b.Mod.Imports[name] = mod
	}
}

// Global Constants
func (b *builder) pass0(f *ast.File) {
	for _, c := range f.Nodes.Constants {
		b.genConstant(c)
	}
}

func (b *builder) genConstant(c *ast.ConstantDeclaration) {
	ident := c.Stmt.Identifier.Value
	value, ok := b.evaluateExpression(c.Stmt.Value, nil, b.Mod).(*lir.Constant)

	if !ok {
		panic("not a constant value")
	}

	b.emitGlobalVar(b.Mod, value, ident)

}
