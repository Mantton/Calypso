package types

import "fmt"

type Alias struct {
	Name string
	RHS  Type
}

func (t *Alias) clyT()        {}
func (t *Alias) Parent() Type { return t.RHS.Parent() }

func (t *Alias) String() string { return fmt.Sprintf("%s(alias of %s)", t.Name, t.RHS.String()) }

func NewAlias(name string, RHS Type) *Alias {
	return &Alias{
		Name: name,
		RHS:  RHS,
	}
}
