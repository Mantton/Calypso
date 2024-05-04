package types

import "fmt"

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
		for _, b := range t.Bounds {
			if IsGeneric(b) {
				return true
			}
		}
		return false
	case *SpecializedFunctionSignature:
		for _, b := range t.Bounds {
			if IsGeneric(b) {
				return true
			}
		}

		return IsGeneric(t.Sg())
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

func ResolveMethod(t Type, n string) (Type, error) {
	switch a := t.(type) {
	case *DefinedType:
		field := a.ResolveField(n)
		if field != nil {
			return field, nil
		}

	case *SpecializedType:
		field := a.ResolveField(n)
		if field != nil {
			return field, nil
		}
	}
	return nil, fmt.Errorf("unknown function, type or field: \"%s\" on type \"%s\"", n, t)

}

func ResolveType(t Type, n string) Type {
	switch a := t.(type) {
	case *DefinedType:
		field := a.ResolveType(n)
		if field != nil {
			return field
		}

	case *SpecializedType:
		field := a.ResolveType(n)
		if field != nil {
			return field
		}
	}

	return nil
}

func IsUnresolved(t Type) bool {
	return ResolveAliases(t) == LookUp(Unresolved)
}

func ResolveSymbol(t Type, n string) (Symbol, Type) {

	switch a := t.(type) {
	case *Pointer:
		return ResolveSymbol(a.PointerTo, n)
	case *DefinedType:
		sym, typ := a.ResolveSymbol(n)

		if sym == nil {
			return nil, nil
		}
		return sym, typ

	case *SpecializedType:
		sym, typ := a.ResolveSymbol(n)

		if sym == nil {
			return nil, nil
		}
		return sym, typ

	case *Module:
		field := a.Scope.ResolveInCurrent(n)

		if field == nil {
			return nil, nil
		}

		return field, field.Type()
	}

	panic("cannot access field of type")
}
