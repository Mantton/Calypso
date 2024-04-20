package lir

import (
	"github.com/mantton/calypso/internal/calypso/types"
)

// Size of a Type in Bytes
func SizeOf(t types.Type) uint64 {
	switch t := t.Parent().(type) {
	case *types.Basic:
		return sizeOfBasic(t)
	case *types.Pointer:
		return 8
	case *types.Struct:
		return sizeOfStruct(t)
	case *types.Enum:
		return sizeOfEnum(t)
	case *types.TypeParam:
		panic("unbound type")
	}
	return 0
}

func sizeOfBasic(t *types.Basic) uint64 {
	switch t.Literal {
	case types.Bool:
		return 1
	case types.Byte, types.UInt8, types.Int8:
		return 1
	case types.UInt16, types.Int16:
		return 2
	case types.UInt32, types.Int32, types.Char:
		return 4
	case types.Int64, types.UInt64:
		return 8
	case types.UInt, types.Int:
		return 8
	case types.Double:
		return 8
	case types.Float:
		return 4
	}

	return 0
}

func sizeOfStruct(t *types.Struct) uint64 {

	total := 0
	for _, f := range t.Fields {
		total += int(SizeOf(f.Type()))
	}

	return uint64(total)
}

func sizeOfEnum(t *types.Enum) uint64 {
	size := SizeOf(types.LookUp(types.Int8))
	maxUnionSize := uint64(0)
	for _, v := range t.Variants {
		maxUnionSize = max(maxUnionSize, sizeOfTaggedVariant(v))
	}

	return size + maxUnionSize
}

func sizeOfTaggedVariant(t *types.EnumVariant) uint64 {
	s := uint64(0)
	for _, f := range t.Fields {
		s += SizeOf(f.Type())
	}

	return s
}
