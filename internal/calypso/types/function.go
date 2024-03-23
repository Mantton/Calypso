package types

import "fmt"

type Function struct {
	symbol
	Target *FunctionTarget
}

type FunctionTarget struct {
	Target string
}

type FunctionSignature struct {
	Scope          *Scope
	TypeParameters []*TypeParam
	Parameters     []*Var
	Result         *Var
}

func (t *FunctionSignature) clyT()        {}
func (t *FunctionSignature) Parent() Type { return t }

func (t *FunctionSignature) String() string {

	f := "fn "

	if len(t.TypeParameters) != 0 {
		f += "<"

		for i, p := range t.TypeParameters {
			f += p.String()

			if i != len(t.TypeParameters)-1 {
				f += ", "
			}
		}

		f += ">"
	}

	f += "(%s) -> %s"
	params := ""
	ret := t.Result.Type().String()

	for i, p := range t.Parameters {
		a := ""

		if len(p.ParamLabel) != 0 {
			a += fmt.Sprintf("%s: ", p.ParamLabel)
		}
		a += p.Type().String()
		params += a

		if i != len(t.Parameters)-1 {
			params += ", "
		}
	}

	return fmt.Sprintf(f, params, ret)
}

func NewFunctionSignature() *FunctionSignature {
	return &FunctionSignature{
		Result: NewVar("", LookUp(Unresolved)),
	}
}

func NewFunction(name string, sg *FunctionSignature) *Function {
	return &Function{
		symbol: symbol{
			name: name,
			typ:  sg,
		},
	}
}

func (n *Function) SetSignature(sg *FunctionSignature) {
	n.SetType(sg)
}

func (sg *FunctionSignature) AddTypeParameter(t *TypeParam) {
	sg.TypeParameters = append(sg.TypeParameters, t)
}

func (sg *FunctionSignature) AddParameter(t *Var) {
	sg.Parameters = append(sg.Parameters, t)
}

func (n *Function) Sg() *FunctionSignature {
	sg, ok := n.typ.(*FunctionSignature)

	if !ok {
		return nil
	}

	return sg
}

func AsFunction(t Symbol) *Function {
	if t == nil {
		return nil
	}

	if a, ok := t.(*Function); ok {
		return a
	}

	return nil

}
