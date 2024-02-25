package types

type Alias struct {
	Definition TypeDef
	RHS        Type
}

func (t *Alias) clyT()          {}
func (t *Alias) String() string { return t.Definition.name }
