package types

type Type interface {
	clyT()
	String() string
	Parent() Type
}

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
		if len(t.TypeParameters) == 0 {
			return false
		}

		// All Type Parameters are bounded
		for _, param := range t.TypeParameters {
			if param.Bound == nil {
				return true
			}
		}

		return false

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
	default:
		return false
	}
}
