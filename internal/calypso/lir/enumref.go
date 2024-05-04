package lir

import (
	"github.com/mantton/calypso/internal/calypso/types"
)

type EnumReference struct {
	Type types.Type
	Enum *types.Enum
}

func (t *EnumReference) Yields() types.Type {
	return t.Type
}

type GenericEnumReference struct {
	Type  *types.DefinedType
	Specs map[string]*EnumReference
}

func (t *GenericEnumReference) Yields() types.Type {
	return t.Type
}

type UnionTypeInlineCreation struct {
	Variant *types.EnumVariant
	Type    types.Symbol
}

func (t *UnionTypeInlineCreation) Yields() types.Type {
	panic("should never yeild")
}
