package types

import "fmt"

type SpecializedFunctionSignature struct {
	Signature *FunctionSignature
	Bounds    TypeList
	Spec      Specialization
	Module    *Module
}

func (t *SpecializedFunctionSignature) Parent() Type { return t }

func (t *SpecializedFunctionSignature) String() string {

	f := ""

	f += "(%s) -> %s"
	params := ""
	spec := t.Specialization()
	ret := Instantiate(t.Signature.Result.Type(), spec, t.Module).String()

	for i, p := range t.Signature.Parameters {
		a := ""

		pT := Instantiate(p.Type(), spec, t.Module)

		if len(p.ParamLabel) != 0 {
			a += fmt.Sprintf("%s: ", p.ParamLabel)
		}

		a += pT.String()
		params += a

		if i != len(t.Signature.Parameters)-1 {
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

	// TODO: Generics have naming clash e.g T & T
	preDef := inMod.FindSpecialized(symbolName)

	if preDef != nil {
		return preDef.(*SpecializedFunctionSignature)
	}
	spec := &SpecializedFunctionSignature{
		Signature: fn,
		Spec:      sub,
		Module:    inMod,
		Bounds:    makeBounds(fn.TypeParameters, sub),
	}

	inMod.Table.Specializations[symbolName] = spec

	return spec
}

func (f *SpecializedFunctionSignature) ReturnType() Type {
	return Instantiate(f.Signature.Result.Type(), f.Specialization(), f.Module)
}

func (f SpecializedFunctionSignature) Sg() *FunctionSignature {
	sg := NewFunctionSignature()

	for _, p := range f.Signature.Parameters {
		v := NewVar(p.name, nil)
		v.SetType(Instantiate(p.typ, f.Spec, f.Module))
		v.ParamLabel = p.ParamLabel
		v.Mutable = p.Mutable
		sg.AddParameter(v)
	}

	sg.Result.SetType(Instantiate(f.Signature.Result.typ, f.Spec, f.Module))
	sg.Function = f.Signature.Function
	return sg
}
