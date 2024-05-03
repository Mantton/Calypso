package types

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
)

type Function struct {
	symbol
	Target *FunctionTarget
	Self   *Var
	Scope  *Scope

	IsPublic bool

	IsAsync    bool
	IsStatic   bool
	IsMutating bool
	AST        *ast.FunctionExpression

	CallGraph map[Type]struct{}
}

func (t *Function) String() string {
	return t.name
}

type FunctionTarget struct {
	Target string
}

type FunctionSignature struct {
	TypeParameters TypeParams
	Parameters     []*Var
	Result         *Var
	Function       *Function
}

func (t *FunctionSignature) Parent() Type { return t }

func (t *FunctionSignature) String() string {

	f := ""

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

func NewFunction(name string, sg *FunctionSignature, mod *Module, expr *ast.FunctionExpression) *Function {
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
	return sg.Function.Scope.Define(t)

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

func (fn *Function) IsVisible(from *Module) bool {
	// Is Public or being accessed from current module
	if fn.IsPublic || from == fn.mod {
		return true
	}

	return false
}

func (fn *Function) AddCallEdge(t Type) {
	if fn.CallGraph == nil {
		fn.CallGraph = make(map[Type]struct{})
	}

	fn.CallGraph[t] = struct{}{}
}
