package types

import "fmt"

type SpecializedFunctionSignature struct {
	InstanceOf *FunctionSignature
	Bounds     TypeList
	Spec       Specialization
	Module     *Module
	sg         *FunctionSignature
}

func (t *SpecializedFunctionSignature) Parent() Type { return t }

func (t *SpecializedFunctionSignature) String() string {

	f := ""

	f += "(%s) -> %s"
	params := ""
	spec := t.Specialization()
	ret := Instantiate(t.InstanceOf.Result.Type(), spec, t.Module).String()

	for i, p := range t.InstanceOf.Parameters {
		a := ""

		pT := Instantiate(p.Type(), spec, t.Module)

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

func NewSpecializedFunctionSignature(fn *FunctionSignature, sub Specialization, inMod *Module) *SpecializedFunctionSignature {

	bounds := makeBounds(fn.TypeParameters, sub)
	symbolName := SpecializedSymbolName(fn.Function, bounds)

	preDef := inMod.FindSpecializedFn(symbolName)

	if preDef != nil {
		return preDef
	}

	spec := &SpecializedFunctionSignature{
		InstanceOf: fn,
		Spec:       sub,
		Module:     inMod,
		Bounds:     makeBounds(fn.TypeParameters, sub),
	}

	inMod.Table.SpecializedFunctions[symbolName] = spec
	for fn := range fn.Function.CallGraph {
		Instantiate(fn, sub, inMod)
	}
	return spec
}

func (f *SpecializedFunctionSignature) ReturnType() Type {
	return Instantiate(f.InstanceOf.Result.Type(), f.Specialization(), f.Module)
}

func (f SpecializedFunctionSignature) Sg() *FunctionSignature {

	if f.sg != nil {
		return f.sg
	}

	f.sg = NewFunctionSignature()

	for _, p := range f.InstanceOf.Parameters {
		v := NewVar(p.name, nil)
		v.SetType(Instantiate(p.typ, f.Spec, f.Module))
		v.ParamLabel = p.ParamLabel
		v.Mutable = p.Mutable
		f.sg.AddParameter(v)
	}

	f.sg.Result.SetType(Instantiate(f.InstanceOf.Result.typ, f.Spec, f.Module))
	f.sg.Function = f.InstanceOf.Function
	return f.sg
}
