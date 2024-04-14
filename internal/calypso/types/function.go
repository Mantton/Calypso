package types

import (
	"fmt"
	"strings"
)

type Function struct {
	symbol
	Target *FunctionTarget
}

type FunctionTarget struct {
	Target string
}

type FunctionSignature struct {
	Scope          *Scope
	TypeParameters TypeParams
	Parameters     []*Var
	Result         *Var
	Self           *Var
	IsAsync        bool
	IsStatic       bool
	IsMutating     bool
	Instances      map[string]*FunctionSignature
	ParentInstance *FunctionSignature
	Function       *Function
	InstanceHash   string
}

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

func NewFunction(name string, sg *FunctionSignature, mod *Module) *Function {
	fn := &Function{
		symbol: symbol{
			name: name,
			typ:  sg,
			mod:  mod,
		},
	}

	sg.Function = fn
	return fn
}

func (n *Function) SetSignature(sg *FunctionSignature) {
	n.SetType(sg)
	sg.Function = n
}

func (sg *FunctionSignature) AddTypeParameter(t *TypeParam) error {
	sg.TypeParameters = append(sg.TypeParameters, t)
	return sg.Scope.Define(t)

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

func IsFunctionSignature(t Type) bool {
	if t == nil {
		return false
	}

	_, ok := t.(*FunctionSignature)
	return ok
}

func (sg *FunctionSignature) ResolveParent() *FunctionSignature {
	if sg.ParentInstance == nil {
		return sg
	}
	return sg.ParentInstance
}

func (tg *FunctionSignature) AddInstance(t *FunctionSignature, m mappings) {

	sg := tg.ResolveParent()
	if sg.Instances == nil {
		sg.Instances = make(map[string]*FunctionSignature)
	}

	str := HashValue(m, sg.TypeParameters)

	key := strings.ReplaceAll(str, " ", "-")
	fmt.Println("HashValue:", key)

	sg.Instances[key] = t

	t.ParentInstance = sg
	t.InstanceHash = key
}
