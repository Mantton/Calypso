package typechecker

import "github.com/mantton/calypso/internal/calypso/types"

type NodeContext struct {
	scope *types.Scope
	sg    *types.FunctionSignature
	lhs   types.Type
}

func NewContext(s *types.Scope, sg *types.FunctionSignature, l types.Type) *NodeContext {
	return &NodeContext{
		scope: s,
		sg:    sg,
		lhs:   l,
	}
}
