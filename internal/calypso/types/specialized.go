package types

type SpecializedType struct {
	Bounds     TypeList
	InstanceOf *DefinedType

	wrapped Type
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
	return t.wrapped
}
