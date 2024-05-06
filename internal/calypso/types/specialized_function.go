package types

import "fmt"

type SpecializedFunctionSignature struct {
	InstanceOf *FunctionSignature
	Bounds     TypeList
	Spec       Specialization
	sg         *FunctionSignature
}

func (t *SpecializedFunctionSignature) Parent() Type { return t }

func (t *SpecializedFunctionSignature) String() string {

	f := ""

	f += "(%s) -> %s"
	params := ""
	spec := t.Specialization()
	ret := Instantiate(t.InstanceOf.Result.Type(), spec).String()

	for i, p := range t.InstanceOf.Parameters {
		a := ""

		pT := Instantiate(p.Type(), spec)

		if len(p.ParamLabel) != 0 {
			a += fmt.Sprintf("%s: ", p.ParamLabel)
		}

		a += pT.String()
		params += a

		if i != len(t.InstanceOf.Parameters)-1 {
			params += ", "
		}
	}

	return fmt.Sprintf(f, params, ret)
}

func (t *SpecializedFunctionSignature) Specialization() Specialization {
	return t.Spec
}

func NewSpecializedFunctionSignature(fn *FunctionSignature, sub Specialization) *SpecializedFunctionSignature {

	spec := &SpecializedFunctionSignature{
		InstanceOf: fn,
		Spec:       sub,
		Bounds:     makeBounds(fn.TypeParameters, sub),
	}

	preDef := fn.Function.FindSpec(spec.SymbolName())

	if preDef != nil {
		return preDef
	}

	fn.Function.AddSpec(spec.SymbolName(), spec)
	for fn := range fn.Function.CallGraph {
		Instantiate(fn, sub)
	}
	return spec
}

func (f *SpecializedFunctionSignature) ReturnType() Type {
	return Instantiate(f.InstanceOf.Result.Type(), f.Specialization())
}

func (f *SpecializedFunctionSignature) Sg() *FunctionSignature {

	if f.sg != nil {
		return f.sg
	}

	f.sg = NewFunctionSignature()

	for _, p := range f.InstanceOf.Parameters {
		v := NewVar(p.name, nil, p.mod)
		v.SetType(Instantiate(p.typ, f.Spec))
		v.ParamLabel = p.ParamLabel
		v.Mutable = p.Mutable
		f.sg.AddParameter(v)
	}

	f.sg.Result.SetType(Instantiate(f.InstanceOf.Result.typ, f.Spec))
	f.sg.Function = f.InstanceOf.Function
	return f.sg
}

func (f *SpecializedFunctionSignature) SymbolName() string {

	fn := f.InstanceOf.Function

	base := ""

	if fn.Self == nil {
		base += fn.symbol.SymbolName()
	} else {
		gen := Instantiate(fn.Self.Type(), f.Spec)
		base += SymbolName(gen) + "::" + fn.name
	}

	return SSym(base, f.Bounds)
}
