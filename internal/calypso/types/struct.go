package types

type Struct struct {
}

func (t *Struct) clyT()          {}
func (t *Struct) String() string { return "struct" }
