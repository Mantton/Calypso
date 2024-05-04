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
		b.registerStruct(n)
	}
}

func (b *builder) genEnum(n *ast.EnumStatement) {
	symbol := b.Mod.TModule.Table.GetNodeType(n)

	if types.IsGeneric(symbol) {
		return
	}

	underlying, ok := symbol.Parent().(*types.Enum)
	if !ok {
		panic("node is not enum type")
	}

	b.Refs[underlying.Name] = &lir.TypeRef{
		Type: symbol,
	}

	defer fmt.Println("<ENUM>", underlying)

	if !underlying.IsUnion() {
		return
	}

	b.genTaggedUnion(symbol.(types.Symbol), underlying)

}
func (b *builder) registerStruct(n *ast.StructStatement) {

	nT := b.Mod.TModule.Table.GetNodeType(n)
	symbol := types.AsDefined(nT)

	if types.IsGeneric(symbol) {
		b.genGenericStruct(symbol)
		return
	}

	// Generate & Store Composite
	c := b.genStructComposite(symbol.SymbolName(), symbol.Parent().(*types.Struct), symbol)
	b.Mod.Composites[c.Name] = c
	b.MP.Composites[symbol] = c
}

func (b *builder) genTaggedUnion(symbol types.Symbol, n *types.Enum) {

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
		Type: n,
		Name: symbol.SymbolName(),
		Members: []types.Type{
			byt,
			&lir.StaticArray{
				OfType: byt,
				Count:  int(maxUnionSize),
			},
		},
	}
	b.Mod.Composites[symbol.SymbolName()] = baseComposite
	b.MP.Composites[symbol.Type()] = baseComposite

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

		symbolName := EnumVariantSymbolName(variant, symbol)
		members = append(members, ts...)
		composite := &lir.Composite{
			Type:      variant,
			Name:      symbolName,
			Members:   members,
			IsAligned: paddingSize != 0,
		}

		b.Mod.Composites[symbolName] = composite
		b.MP.Composites[variant] = composite
		composite.EnumParent = baseComposite
	}
}

func EnumVariantSymbolName(v *types.EnumVariant, sym types.Symbol) string {
	return sym.SymbolName() + "::_v::" + v.Name
}

// Structs
func (b *builder) genStructComposite(name string, underlying *types.Struct, t types.Type) *lir.Composite {
	c := &lir.Composite{
		Type: t,
		Name: name,
	}

	for _, f := range underlying.Fields {
		c.Members = append(c.Members, f.Type())
	}

	fmt.Println("Generated Composite", c)
	return c
}

func (b *builder) genGenericStruct(symbol *types.DefinedType) {

	// create generic type
	t := lir.NewGenericType(symbol)
	b.Mod.GTypes[symbol.SymbolName()] = t

	// add specs to type
	for _, sT := range symbol.AllSpecs() {

		c := b.genStructComposite(sT.SymbolName(), sT.Parent().(*types.Struct), sT)
		t.Specs[c.Name] = c
		b.MP.Composites[sT] = c
	}
}
