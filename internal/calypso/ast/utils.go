package ast

import (
	"github.com/mantton/calypso/internal/calypso/fs"
	"github.com/mantton/calypso/internal/calypso/lexer"
)

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
	FSMod        *fs.Module
	Package      *Package
	Visibility   Visibility
}

type Package struct {
	Modules   map[string]*Module
	FSPackage *fs.LitePackage
}

func (m *Module) Name() string {
	return m.Set.ModuleName
}

func (p *Package) Name() string {
	return p.FSPackage.Config.Package.Name
}
