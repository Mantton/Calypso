package types

type TypeParam struct {
	Name        string      // the name of the type parameter, e.g T, V
	Constraints []*Standard // standards, this param is constrained to
	Bound       Type        // the type this param is bound to after initialization
}

func NewTypeParam(n string, cns []*Standard, b Type) *TypeParam {
	return &TypeParam{
		Name:        n,
		Constraints: cns,
		Bound:       b,
	}
}

type TypeParams []*TypeParam

func (t *TypeParam) clyT()          {}
func (t *TypeParam) String() string { return t.Name }
func (t *TypeParam) Parent() Type   { return t }

func (n *TypeParam) AddConstraint(s *Standard) {
	n.Constraints = append(n.Constraints, s)
}

func AsTypeParam(t Type) *TypeParam {
	if a, ok := t.(*TypeParam); ok {
		return a
	}
	return nil

}

func (n *TypeParam) Unwrapped() Type {
	if n.Bound != nil {
		return n.Bound
	}
	return n
}
