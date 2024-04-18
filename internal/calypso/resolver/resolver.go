package resolver

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/fs"
	"github.com/mantton/calypso/internal/calypso/parser"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

type resolver struct {
	nodes       map[string]graph.Node
	seen        map[string]struct{}
	packages    map[string]*fs.LitePackage
	nodePackage map[string]*fs.LitePackage
	rnodes      map[graph.Node]string
	modules     map[string]*ast.Module

	dg *simple.DirectedGraph
}

type ResolvedData struct {
	Packages       map[string]*ast.Package
	OrderedModules []*ast.Module
}

func ParseAndResolve(pkg *fs.LitePackage) (*ResolvedData, error) {
	// Collect paths
	r := &resolver{
		nodes:       make(map[string]graph.Node),
		rnodes:      make(map[graph.Node]string),
		nodePackage: make(map[string]*fs.LitePackage),
		seen:        make(map[string]struct{}),
		packages:    make(map[string]*fs.LitePackage),
		modules:     make(map[string]*ast.Module),
		dg:          simple.NewDirectedGraph(),
	}

	// resolve first package
	err := r.resolvePackage(pkg)

	if err != nil {
		return nil, err
	}

	for len(r.seen) != len(r.nodes) {

		for x := range r.nodes {
			r.resolvePath(x)
		}
	}

	// sort & find cyclic imports
	sorted, err := topo.Sort(r.dg)

	if err != nil {
		return nil, err
	}

	l := []*ast.Module{}
	slices.Reverse(sorted) // this is the order we need

	for _, node := range sorted {
		p := r.rnodes[node]
		m := r.modules[p]
		l = append(l, m)
	}

	packages := make(map[string]*ast.Package)
	for k, mod := range r.modules {

		fsPkg := r.nodePackage[k]

		if v, ok := packages[fsPkg.Path]; ok {
			v.Modules[k] = mod
			mod.Package = v
			continue
		}

		// create pacakge
		astPkg := &ast.Package{
			FSPackage: fsPkg,
			Modules:   make(map[string]*ast.Module),
		}

		astPkg.Modules[k] = mod
		packages[fsPkg.Path] = astPkg
	}

	return &ResolvedData{
		OrderedModules: l,
		Packages:       packages,
	}, nil
}

func (r *resolver) resolvePackage(pkg *fs.LitePackage) error {
	r.packages[pkg.Path] = pkg

	// load base module
	src := filepath.Join(pkg.Path, "src")
	mod, err := fs.CollectModule(src, false)
	if err != nil {
		return err
	}

	err = r.addModuleNodes(mod, pkg)

	if err != nil {
		return err
	}

	return nil
}

func (r *resolver) resolvePath(path string) error {
	if _, ok := r.seen[path]; ok {
		return nil
	}

	pkg := r.nodePackage[path]

	mod, err := fs.CollectModule(path, false)
	if err != nil {
		return err
	}

	err = r.addModuleNodes(mod, pkg)

	return err
}

func (r *resolver) addModuleNodes(m *fs.Module, pkg *fs.LitePackage) error {
	if _, ok := r.seen[m.Path]; ok {
		return nil
	}

	// parse module
	mod, err := parser.ParseModule(m)

	if err != nil {
		return err
	}

	// Add current module
	if _, ok := r.nodes[m.Path]; !ok {
		node := r.dg.NewNode()
		r.dg.AddNode(node)
		r.nodes[m.Path] = node
		r.rnodes[node] = m.Path
	}

	r.seen[m.Path] = struct{}{}
	r.modules[m.Path] = mod
	r.nodePackage[m.Path] = pkg
	mod.FSMod = m

	// collect module deps
	err = r.resolveDependencies(mod, pkg)
	if err != nil {
		return err
	}

	return nil
}

func (r *resolver) addPathNode(path string) {

	if _, ok := r.nodes[path]; ok {
		return
	}
	node := r.dg.NewNode()
	r.dg.AddNode(node)
	r.nodes[path] = node
	r.rnodes[node] = path
}

func (r *resolver) addEdge(x, y string) {
	e := r.dg.NewEdge(r.nodes[x], r.nodes[y])
	r.dg.SetEdge(e)
}

func (r *resolver) resolveDependencies(m *ast.Module, pkg *fs.LitePackage) error {
	deps := []string{}
	for _, file := range m.Set.Files {
		for _, decl := range file.Nodes.Imports {

			dep, err := r.findDependency(decl, pkg)

			if err != nil {
				return err
			}

			deps = append(deps, dep)
		}
	}

	for _, dep := range deps {
		r.addPathNode(dep)
		r.addEdge(m.FSMod.Path, dep)
	}

	return nil

}

func (r *resolver) findDependency(decl *ast.ImportDeclaration, pkg *fs.LitePackage) (string, error) {

	importPath := decl.Path.Value
	if len(importPath) == 0 {
		return "", fmt.Errorf("unable to resolve dependency")
	}

	splitPath := strings.Split(importPath, "/")

	if len(splitPath) < 1 {
		return "", fmt.Errorf("unable to resolve dependency")
	}

	// First Element is always a package

	p := splitPath[0]

	var path string
	var err error
	if p == pkg.Config.Package.Name {
		// resolving local module
		path, err = r.findLocalDependency(splitPath[1:], pkg) // without package name

	} else {
		// resolving external dependency
		path, err = r.findExternalDependency(splitPath, pkg)
	}

	if err != nil {
		return path, err
	}

	// Map Node
	decl.PopulatedImportKey = path
	return path, err
}

func (r *resolver) findLocalDependency(paths []string, pkg *fs.LitePackage) (string, error) {
	// 1 - The `paths` param contains the split path without the Package Name
	packageBase := pkg.Path
	path := filepath.Join(packageBase, "src") // Points to the src folder of the current package

	for _, elem := range paths {
		path = filepath.Join(path, elem)
	}

	// 2 - Ensure Path Exists
	_, err := os.Stat(path)

	if err != nil {
		return "", err
	}

	r.nodePackage[path] = pkg
	// 3- Return Correct Path
	return path, nil
}

func (r *resolver) findExternalDependency(paths []string, pkg *fs.LitePackage) (string, error) {

	p := paths[0] // target package

	// Find In Config File
	cfg := pkg.Config
	dep := cfg.FindDependency(p)

	if dep == nil {
		return "", fmt.Errorf("unable to locate package, %s", p)
	}

	var tgt *fs.LitePackage
	if r.packages[dep.FilePath()] == nil {
		// add package
		p, err := fs.CreateLitePackage(dep.FilePath(), true)

		if err != nil {
			return "", nil
		}
		r.packages[dep.FilePath()] = p
		tgt = p
	} else {
		tgt = r.packages[dep.FilePath()]
	}
	// Now map to path
	path := filepath.Join(dep.FilePath(), "src") // Points to the src folder of the current package

	for _, elem := range paths[1:] {
		path = filepath.Join(path, elem)
	}

	// 2 - Ensure Path Exists
	_, err := os.Stat(path)

	if err != nil {
		return "", err
	}

	// 3- Return Correct Path
	r.nodePackage[path] = tgt
	return path, nil
}
