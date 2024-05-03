package types

import (
	"fmt"
	"sync/atomic"
)

var tick int64

type TypeParam struct {
	symbol
	Constraints []*Standard // standards, this param is constrained to
	ID          int64
}

func NewTypeParam(n string, cns []*Standard) *TypeParam {

	newID := atomic.AddInt64(&tick, 1)

	return &TypeParam{
		symbol: symbol{
			name: n,
		},
		ID:          newID,
		Constraints: cns,
	}
}

type TypeParams []*TypeParam

func (t *TypeParam) String() string {
	return fmt.Sprintf("%s_%d", t.name, t.ID)
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
