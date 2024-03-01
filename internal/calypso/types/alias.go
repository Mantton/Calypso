package types

import "fmt"

type Alias struct {
	Name string
	RHS  DefinedType
}

func (t *Alias) clyT()        {}
func (t *Alias) Parent() Type { return t.RHS.Parent() }

func (t *Alias) String() string { return fmt.Sprintf("%s(alias of %s)", t.Name, t.RHS.Name()) }
