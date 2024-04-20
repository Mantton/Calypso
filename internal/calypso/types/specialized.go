package types

type SpecializedType struct {
	Bounds     TypeList
	InstanceOf *DefinedType
}

type SpecializedFunction struct {
	Fn     *Function
	Bounds TypeList
}

func NewSpecializedType(def *DefinedType, bounds TypeList) *SpecializedType {
	return &SpecializedType{
		Bounds:     bounds,
		InstanceOf: def,
	}
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

func (t *SpecializedType) ResolveMethod(f string) Type {

	field := t.InstanceOf.ResolveMethod(f)

	if field == nil {
		return nil
	}

	return Instantiate(field, t.Specialization())
}

func (t *SpecializedType) Specialization() Specialization {
	spec := make(Specialization)
	for i, p := range t.Bounds {
		spec[t.InstanceOf.TypeParameters[i]] = p
	}
	return spec
}
