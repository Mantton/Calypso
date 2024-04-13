package lirgen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lir"
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
	def := b.Mod.TypeTable().GetNodeType(n)
	t, ok := def.Parent().(*types.Enum)

	if !ok {
		panic("node is not enum type")
	}

	b.Refs[t.Name] = &lir.TypeRef{
		Type: def,
	}

	defer fmt.Println("<ENUM>", t)

	if !t.IsUnion() {
		return
	}

	b.genTaggedUnion(t, n.Identifier.Value)

}
func (b *builder) genStruct(n *ast.StructStatement) {
	t, ok := b.Mod.TypeTable().GetNodeType(n).Parent().(*types.Struct)

	if !ok {
		panic("node is not struct type")
	}

	fmt.Println("<STRUCT>", t)

	c := &lir.Composite{
		Actual: t,
		Name:   n.Identifier.Value,
	}
	for _, f := range t.Fields {
		c.Members = append(c.Members, f.Type())
	}

	b.Mod.Composites[t] = c
}

func (b *builder) genTaggedUnion(n *types.Enum, name string) {

	// Take
	/*
		Take

		enum Foo {
			ABool(bool),
			ADouble(double),
			AInt(int),
		}

	*/

	// 1 - Generate Base Composite
	byt := types.LookUp(types.Int8)
	size := lir.SizeOf(n) // Get Size of Dicrimimant + Max Tagged Union Size

	discrimimantSize := lir.SizeOf(byt) // Get Size of Discrimimant Size
	maxUnionSize := size - discrimimantSize

	// 2 - Base Composite can simply be 1X i8 (Discriminant) + nX i8 (Max Union)
	baseComposite := &lir.Composite{
		Actual: n,
		Name:   name,
		Members: []types.Type{
			byt,
			&lir.StaticArray{
				OfType: byt,
				Count:  int(maxUnionSize),
			},
		},
	}
	b.Mod.Composites[n] = baseComposite

	// 3 - Generate Composite Types for each tagged union
	for _, variant := range n.Variants {

		paddingSize := maxUnionSize
		ts := []types.Type{}
		for _, field := range variant.Fields {
			paddingSize -= lir.SizeOf(field.Type())
			ts = append(ts, field.Type())
		}

		members := []types.Type{
			byt, // Discriminant
		}

		if paddingSize != 0 {
			members = append(members, &lir.StaticArray{
				OfType: byt,
				Count:  int(paddingSize),
			})
		}

		members = append(members, ts...)
		composite := &lir.Composite{
			Actual:    variant,
			Name:      name + "." + variant.Name,
			Members:   members,
			IsAligned: paddingSize != 0,
		}

		b.Mod.Composites[variant] = composite
		composite.EnumParent = baseComposite
	}
}
