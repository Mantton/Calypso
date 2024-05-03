package types

import "fmt"

type SpecializedType struct {
	Bounds     TypeList
	Spec       Specialization
	InstanceOf *DefinedType
}

func NewSpecializedType(def *DefinedType, sub Specialization) *SpecializedType {
	bounds := makeBounds(def.TypeParameters, sub)
	symbolName := SpecializedSymbolName(def, bounds)

	preDef := def.FindSpec(symbolName)

	if preDef != nil {
		return preDef
	}

	spec := &SpecializedType{
		Spec:       sub,
		InstanceOf: def,
		Bounds:     bounds,
	}

	def.AddSpec(symbolName, spec)
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

func (f *SpecializedType) SymbolName() string {
	return SpecializedSymbolName(f.InstanceOf, f.Bounds)
}

func makeBounds(params TypeParams, ctx Specialization) TypeList {
	bounds := TypeList{}
	for _, p := range params {
		arg, ok := ctx[p]

		if !ok {
			fmt.Println("DEBUG - Unspecialized TypeParameter")
			return nil
		}

		bounds = append(bounds, arg)
	}

	return bounds
}
