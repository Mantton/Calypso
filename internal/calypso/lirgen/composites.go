package lirgen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

// Composites
func (b *builder) pass1(f *ast.File) {
	for _, n := range f.Nodes.Enums {
		b.genEnum(n)
	}

	for _, n := range f.Nodes.Structs {
		b.genStruct(n)
	}
}

func (b *builder) genEnum(n *ast.EnumStatement) {
	t, ok := b.Mod.TypeTable().GetNodeType(n).Parent().(*types.Enum)

	if !ok {
		panic("node is not enum type")
	}

	fmt.Println("<ENUM>", t)
}
func (b *builder) genStruct(n *ast.StructStatement) {
	t, ok := b.Mod.TypeTable().GetNodeType(n).Parent().(*types.Struct)

	if !ok {
		panic("node is not struct type")
	}

	fmt.Println("<STRUCT>", t)
}
