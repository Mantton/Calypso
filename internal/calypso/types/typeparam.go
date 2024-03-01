package types

type TypeParam struct {
	Name        string
	Constraints []Type
}

func NewTypeParam(n string, cns []Type) *TypeParam {
	return &TypeParam{
		Name:        n,
		Constraints: cns,
	}
}

type TypeParams []*TypeParam

func (t *TypeParam) clyT()          {}
func (t *TypeParam) String() string { return t.Name }
func (t *TypeParam) Parent() Type   { return t }
