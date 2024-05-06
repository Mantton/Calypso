package lir

// *

type Package struct {
	Modules map[string]*Module
	Name    string
}

func NewPackage(n string) *Package {
	return &Package{
		Modules: map[string]*Module{},
		Name:    n,
	}
}
