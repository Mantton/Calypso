package types

type Type interface {
	String() string
	Parent() Type
}

type TypeList []Type

type TParam interface {
	TypeParameters() TypeParams
}

func IsGeneric(t Type) bool {
	t = ResolveAliases(t)
	switch t := t.(type) {
	case *Pointer:
		return IsGeneric(t.PointerTo)
	case *TypeParam:
		return true
	case *DefinedType:
		// No Type Parameters
		return len(t.TypeParameters) > 0
	case *FunctionSignature:
		if len(t.TypeParameters) != 0 {
			return true
		}

		for _, p := range t.Parameters {
			if IsGeneric(p.Type()) {
				return true
			}
		}

		if IsGeneric(t.Result.Type()) {
			return true
		}

		return false
	case *SpecializedType:
		return false
	default:
		return false
	}
}

func IsAssignable(t Type) bool {
	if _, ok := t.(*Module); ok {
		return false
	}

	return true
}
