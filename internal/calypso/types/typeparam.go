package types

type TypeParam struct {
	Name        string
	Constraints []*Standard
}

func NewTypeParam(n string, cns []*Standard) *TypeParam {
	return &TypeParam{
		Name:        n,
		Constraints: cns,
	}
}

type TypeParams []*TypeParam

func (t *TypeParam) clyT()          {}
func (t *TypeParam) String() string { return t.Name }
func (t *TypeParam) Parent() Type   { return t }

func (n *TypeParam) AddConstraint(s *Standard) {
	n.Constraints = append(n.Constraints, s)
}
