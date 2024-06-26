package ast

import (
	"fmt"
	"sync/atomic"

	"github.com/mantton/calypso/internal/calypso/fs"
	"github.com/mantton/calypso/internal/calypso/lexer"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
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
	Info         *fs.Module
	Package      *Package
	Visibility   Visibility
	id           int64
}

func NewModule(i *fs.Module, pkg *Package) *Module {
	newID := atomic.AddInt64(&mTick, 1)

	return &Module{
		Info:    i,
		Package: pkg,
		id:      newID,
	}
}

var mTick int64
var pTick int64

func (m *Module) ID() int64 {
	return m.id
}

func (m *Module) Name() string {
	return m.Set.ModuleName
}

func (m *Module) Key() string {
	return fmt.Sprintf("%s::%s", m.Package.Key(), m.Set.ModuleName)
}

func (m *Module) AddModule(s *Module) {
	m.SubModules[s.Name()] = s
}

func (m *Module) IsSTD() bool {
	return m.Package.Info.IsSTD()
}

// * Package

type Package struct {
	Modules  map[string]*Module
	Info     *fs.Package
	graph    *simple.DirectedGraph
	id       int64
	IsTarget bool
}

func (p *Package) Name() string {
	return p.Info.Config.Package.Name
}

func NewPackage(fs *fs.Package) *Package {
	newID := atomic.AddInt64(&pTick, 1)

	return &Package{
		Info: fs,
		id:   newID,
	}
}

func (p *Package) Key() string {
	return p.Info.ID()
}

func (p *Package) ID() int64 {
	return p.id
}

func (p *Package) AddModule(m *Module) {
	if p.Modules == nil {
		p.Modules = make(map[string]*Module)
	}

	p.Modules[m.Name()] = m

	if p.graph == nil {
		p.graph = simple.NewDirectedGraph()
	}

	p.graph.AddNode(m)
}

func (p *Package) SetEdge(m1, m2 *Module) {
	e := p.graph.NewEdge(m1, m2)
	p.graph.SetEdge(e)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func (p *Package) PerformInOrder(fn func(*Module) error) error {

	errs := []error{}
	nodes, err := topo.Sort(p.graph)

	if err != nil {
		return err
	}
	for i := range nodes {
		j := abs(i - len(nodes) + 1)
		mod := nodes[j].(*Module)

		err := fn(mod)

		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return lexer.CombinedErrors(errs)
	}

	return nil
}
