package types

type TypeParam struct {
	Definition  *TypeDef
	Constraints []Type
}

func NewTypeParam(def *TypeDef, cns []Type) *TypeParam {
	return &TypeParam{
		Definition:  def,
		Constraints: cns,
	}
}

type TypeParams []*TypeParam

func (t *TypeParam) clyT()          {}
func (t *TypeParam) String() string { return t.Definition.name }
