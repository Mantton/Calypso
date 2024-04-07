package lir

import "github.com/mantton/calypso/internal/calypso/types"

type TypeRef struct {
	Type types.Type
}

func (t *TypeRef) Yields() types.Type {
	return t.Type
}
