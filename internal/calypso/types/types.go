package types

type Type interface {
	clyT()
	String() string
	Parent() Type
}

func IsGeneric(t Type) bool {
	switch t := t.(type) {
	case *Pointer:
		return IsGeneric(t.PointerTo)
	case *TypeParam:
		return true
	case *DefinedType:
		return len(t.TypeParameters) != 0 || IsGeneric(t.Parent())
	case *FunctionSignature:
		return len(t.TypeParameters) != 0
	case *EnumInstance:
		for _, arg := range t.TypeArgs {
			if IsGeneric(arg) {
				return true
			}
		}

		return false
	default:
		return false
	}
}
