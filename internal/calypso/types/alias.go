package types

import "fmt"

type Alias struct {
	symbol
	TypeParameters TypeParams
	RHS            Type
}

func (t *Alias) clyT()        {}
func (t *Alias) Parent() Type { return t.RHS.Parent() }

func (t *Alias) String() string {
	if len(t.TypeParameters) == 0 {
		return fmt.Sprintf("%s / %s", t.Name(), t.RHS)
	} else {
		f := t.Name() + "<"

		for i, p := range t.TypeParameters {
			f += p.String()

			if i != len(t.TypeParameters)-1 {
				f += ", "
			}
		}
		f += ">"
		return f + " / " + t.RHS.String()
	}
}

func NewAlias(name string, RHS Type) *Alias {
	return &Alias{
		symbol: symbol{
			name: name,
			typ:  nil,
		},
		RHS: RHS,
	}
}

func (s *Alias) AddTypeParameter(t *TypeParam) {
	s.TypeParameters = append(s.TypeParameters, t)
}

func (a *Alias) SetType(t Type) {
	a.RHS = (t)
}

func (t *Alias) Type() Type {
	return t
}

func ResolveAliases(t Type) Type {
	switch t := t.(type) {
	case *Alias:
		return ResolveAliases(t.RHS)
	default:
		return t
	}
}

func UnwrapBounded(t Type) Type {
	switch t := t.(type) {
	case *TypeParam:
		return t.Unwrapped()
	default:
		return t
	}
}

func AsAlias(t Type) *Alias {

	if t == nil {
		return nil
	}

	if a, ok := t.(*Alias); ok {
		return a
	}

	return nil
}
