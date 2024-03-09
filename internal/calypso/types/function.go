package types

import "fmt"

type Function struct {
	symbol
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
		params += p.Type().String()

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

func (sg *FunctionSignature) AddTypeParameter(t *TypeParam) {
	sg.TypeParameters = append(sg.TypeParameters, t)
}

func (sg *FunctionSignature) AddParameter(t *Var) {
	sg.Parameters = append(sg.Parameters, t)
}

func (n *Function) Sg() *FunctionSignature {
	return n.typ.(*FunctionSignature)
}
