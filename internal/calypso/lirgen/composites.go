package lirgen

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/types"
)

// Composites
func (b *builder) pass1(f *ast.File) {
	for _, n := range f.Nodes.Enums {
		b.registerEnum(n)
	}

	for _, n := range f.Nodes.Structs {
		b.registerStruct(n)
	}
}

func (b *builder) registerEnum(n *ast.EnumStatement) {
	nT := b.Mod.TModule.Table.GetNodeType(n)
	symbol := types.AsDefined(nT)

	if types.IsGeneric(symbol) {
		b.genGenericEnums(symbol)
		return
	}

	underlying := symbol.Type().Parent().(*types.Enum)
	b.genEnumComposite(symbol, underlying)

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

func (b *builder) genTaggedUnion(symbol types.Type, n *types.Enum) {

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

	baseSymbolName := types.SymbolName(symbol)

	// 2 - Base Composite can simply be 1X i8 (Discriminant) + nX i8 (Max Union)
	baseComposite := &lir.Composite{
		Type: symbol,
		Name: baseSymbolName,
		Members: []types.Type{
			byt,
			&lir.StaticArray{
				OfType: byt,
				Count:  int(maxUnionSize),
			},
		},
	}
	b.Mod.Composites[baseSymbolName] = baseComposite
	b.MP.Composites[symbol] = baseComposite

	// 3 - Generate Composite Types for each tagged union
	for _, variant := range n.Variants {

		if len(variant.Fields) == 0 {
			continue
		}

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

		symbolName := EnumVariantSymbolNameStr(variant, baseSymbolName)
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

func EnumVariantSymbolName(v *types.EnumVariant, sym types.Type) string {
	return types.SymbolName(sym) + "::_V::" + v.Name
}

func EnumVariantSymbolNameStr(v *types.EnumVariant, sym string) string {
	return sym + "::_V::" + v.Name
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

func (b *builder) genEnumComposite(t types.Type, underlying *types.Enum) *lir.EnumReference {
	ref := &lir.EnumReference{
		Type: t,
		Enum: underlying,
	}

	b.Mod.Enums[types.SymbolName(t)] = ref

	if underlying.IsUnion() {
		b.genTaggedUnion(t, underlying)
	}
	return ref
}

func (b *builder) genGenericEnums(symbol *types.DefinedType) {

	ref := &lir.GenericEnumReference{
		Type:  symbol,
		Specs: make(map[string]*lir.EnumReference),
	}
	for _, sT := range symbol.AllSpecs() {
		ref.Specs[types.SymbolName(sT)] = b.genEnumComposite(sT, sT.Parent().(*types.Enum))
	}

	b.Mod.GEnums[symbol.SymbolName()] = ref
}
