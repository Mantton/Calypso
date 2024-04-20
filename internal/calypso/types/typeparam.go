package types

type TypeParam struct {
	symbol
	Constraints []*Standard // standards, this param is constrained to
}

func NewTypeParam(n string, cns []*Standard) *TypeParam {
	return &TypeParam{
		symbol: symbol{
			name: n,
		},
		Constraints: cns,
	}
}

type TypeParams []*TypeParam

func (t *TypeParam) String() string {
	return t.name
}
func (t *TypeParam) Name() string { return t.name }
func (t *TypeParam) Type() Type   { return t }

func (t *TypeParam) Parent() Type { return t }

func (n *TypeParam) AddConstraint(s *Standard) {
	n.Constraints = append(n.Constraints, s)
}

func AsTypeParam(t Type) *TypeParam {
	if a, ok := t.(*TypeParam); ok {
		return a
	}
	return nil

}
