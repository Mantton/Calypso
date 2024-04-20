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

func ResolveField(t Type, f string, m *Module) (Type, error) {

	switch a := t.(type) {
	case *DefinedType:
		field := a.ResolveField(f)
		if field != nil {
			return field, nil
		}

	case *SpecializedType:
		field := a.ResolveField(f)
		if field != nil {
			return field, nil
		}

	case *Module:
		field := a.Table.Main.ResolveInCurrent(f)
		if !field.IsVisible(m) {
			return nil, fmt.Errorf("`%s` is not accessible in this context", f)
		}

		if field != nil {
			return field.Type(), nil
		}
	}

	return nil, fmt.Errorf("unknown function, type or field: \"%s\" on type \"%s\"", f, t)
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

func IsUnresolved(t Type) bool {
	return ResolveAliases(t) == LookUp(Unresolved)
}
