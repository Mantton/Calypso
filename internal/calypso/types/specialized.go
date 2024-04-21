package types

import "fmt"

type SpecializedType struct {
	Bounds     TypeList
	Spec       Specialization
	InstanceOf *DefinedType
}

func NewSpecializedType(def *DefinedType, sub Specialization) *SpecializedType {
	spec := &SpecializedType{
		Spec:       sub,
		InstanceOf: def,
	}

	for _, p := range def.TypeParameters {
		arg, ok := sub[p]

		if !ok {
			fmt.Println("DEBUG - Unspecialized TypeParameter", p)
			return nil
		}

		spec.Bounds = append(spec.Bounds, arg)
	}

	return spec

}

func (t *SpecializedType) String() string {
	f := t.InstanceOf.name + "<"

	for i, p := range t.Bounds {
		f += p.String()

		if i != len(t.Bounds)-1 {
			f += ", "
		}
	}
	f += ">"
	return f
}

func (t *SpecializedType) Parent() Type {
	return cloneWithSpecialization(t.InstanceOf.wrapped, t.Specialization())
}

func (t *SpecializedType) ResolveField(f string) Type {

	field := t.InstanceOf.ResolveField(f)

	if field == nil {
		return nil
	}

	return Instantiate(field, t.Specialization())
}

func (t *SpecializedType) ResolveType(f string) Type {

	field := t.InstanceOf.ResolveType(f)

	if field == nil {
		return nil
	}

	return Instantiate(field, t.Specialization())
}

func (t *SpecializedType) ResolveMethod(f string) Type {

	field := t.InstanceOf.ResolveMethod(f)

	if field == nil {
		return nil
	}

	return Instantiate(field, t.Specialization())
}

func (t *SpecializedType) Specialization() Specialization {
	return t.Spec
}
