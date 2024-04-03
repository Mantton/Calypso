package lir

import "github.com/mantton/calypso/internal/calypso/types"

// * 2
type Executable struct {
	Modules map[string]*Module
}
type Package struct {
	Modules map[string]*Module
}

type Composite struct {
	Members []types.Type
	Actual  types.Type
	Name    string
}
