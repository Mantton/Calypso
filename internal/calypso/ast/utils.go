package ast

import "github.com/mantton/calypso/internal/calypso/lexer"

type File struct {
	ModuleName string
	Nodes      *Nodes
	Errors     lexer.ErrorList
	LexerFile  *lexer.File
}

type FileSet struct {
	ModuleName string
	Files      []*File
}

type Module struct {
	Set          *FileSet
	SubModules   map[string]*Module
	ParentModule *Module
}

type Package struct {
	Source *Module
}

func (m *Module) Name() string {
	return m.Set.ModuleName
}
