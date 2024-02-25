package types

type Function struct {
	symbol
}

type FunctionSignature struct {
	Scope          *Scope
	TypeParameters []*TypeParam
	Parameters     []*Var
	ReturnType     Type
}

func (t *FunctionSignature) clyT()          {}
func (t *FunctionSignature) String() string { return "fn () -> unresolved" }

func NewFunctionSignature() *FunctionSignature {
	return &FunctionSignature{}
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
