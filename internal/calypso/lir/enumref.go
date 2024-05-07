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
	Type    types.Type
}

func (t *UnionTypeInlineCreation) Yields() types.Type {
	return t.Type
}

type EnumExpansionResult struct {
	Discriminant Value
	Emit         func(*Function, Value)
}

func (t *EnumExpansionResult) Yields() types.Type {
	return t.Discriminant.Yields()
}
