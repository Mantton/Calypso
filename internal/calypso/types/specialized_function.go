package types

import "fmt"

type SpecializedFunctionSignature struct {
	Signature *FunctionSignature
	Bounds    TypeList
	Spec      Specialization
}

func (t *SpecializedFunctionSignature) Parent() Type { return t }

func (t *SpecializedFunctionSignature) String() string {

	f := ""

	f += "(%s) -> %s"
	params := ""
	spec := t.Specialization()
	ret := Instantiate(t.Signature.Result.Type(), spec).String()

	for i, p := range t.Signature.Parameters {
		a := ""

		pT := Instantiate(p.Type(), spec)

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

func NewSpecializedFunctionSignature(fn *FunctionSignature, sub Specialization) *SpecializedFunctionSignature {
	spec := &SpecializedFunctionSignature{
		Signature: fn,
		Spec:      sub,
	}

	// Collect Bounded Types
	for _, p := range fn.TypeParameters {
		arg, ok := sub[p]

		if !ok {
			fmt.Println("DEBUG - Unspecialized TypeParameter")
			return nil
		}

		spec.Bounds = append(spec.Bounds, arg)
	}

	return spec
}

func (f *SpecializedFunctionSignature) ReturnType() Type {
	return Instantiate(f.Signature.Result.Type(), f.Specialization())
}

func (f SpecializedFunctionSignature) Sg() *FunctionSignature {
	sg := NewFunctionSignature()

	for _, p := range f.Signature.Parameters {
		v := NewVar(p.name, nil)
		v.SetType(Instantiate(p.typ, f.Spec))
		v.ParamLabel = p.ParamLabel
		v.Mutable = p.Mutable
		sg.AddParameter(v)
	}

	sg.Result.SetType(Instantiate(f.Signature.Result.typ, f.Spec))
	sg.Function = f.Signature.Function
	return sg
}
