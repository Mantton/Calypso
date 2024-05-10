package resolver

import (
	"fmt"
	"strings"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/collections"
	"github.com/mantton/calypso/internal/calypso/fs"
	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/parser"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

type resolver struct {
	pacakges            map[string]*ast.Package    // maps packages to their key value
	errorList           lexer.ErrorList            // list of all errors found during parsing and resolution
	queuedActions       *collections.Stack[func()] // a FIFO stack of functions to be invoked
	programModuleGraph  *simple.DirectedGraph
	programPackageGraph *simple.DirectedGraph
	target              *ast.Package
}

type ResolvedData struct {
	Packages       map[string]*ast.Package
	OrderedModules []*ast.Module
}

func (r *resolver) addError(e error) {
	r.errorList.Add(e)
}

func getSTDPath() string {
	// TODO: STD path
	return "./dev/std"
}
func ParseAndResolve(path string) ([]*ast.Package, error) {
	// Collect paths
	r := &resolver{
		programModuleGraph:  simple.NewDirectedGraph(),
		programPackageGraph: simple.NewDirectedGraph(),
		pacakges:            make(map[string]*ast.Package),
		queuedActions:       &collections.Stack[func()]{},
	}

	// 1 -  Parse STD Package
	r.ParsePackage(getSTDPath(), false)

	// 2 - Parse Target Package
	r.ParsePackage(path, true)

	// Perform Queued actions
	for r.queuedActions.Length() != 0 {
		action, ok := r.queuedActions.Pop()

		if !ok {
			break
		}

		action()
	}

	if len(r.errorList) != 0 {
		return nil, lexer.CombinedErrors(r.errorList)
	}

	sorted, err := topo.Sort(r.programPackageGraph)

	if err != nil {
		return nil, err
	}

	var packages []*ast.Package = make([]*ast.Package, len(sorted))

	for i, node := range sorted {
		j := abs(i - len(sorted) + 1)
		packages[j] = node.(*ast.Package)
	}

	return packages, nil
}
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func (r *resolver) ParsePackage(p string, entry bool) *ast.Package {

	// Read Directory
	pkg, err := fs.CreatePackage(p)
	if err != nil {
		r.addError(err)
		return nil
	}

	astPackage := ast.NewPackage(pkg)
	r.programPackageGraph.AddNode(astPackage)

	// store
	if p == getSTDPath() {
		r.pacakges["std"] = astPackage
	} else {
		r.pacakges[astPackage.Key()] = astPackage
	}

	if entry {
		r.target = astPackage
	}

	for _, mod := range pkg.Modules {
		r.ParseModule(mod, astPackage)
	}

	r.queuedActions.Push(func() {
		r.ResolveImports(astPackage)
	})

	return astPackage
}

func (r *resolver) ParseModule(mod *fs.Module, pkg *ast.Package) {
	// 1 - Parse AST
	ast, err := parser.ParseModule(mod, pkg)

	if err != nil {
		r.addError(err)
		return
	}

	// add to graph
	r.programModuleGraph.AddNode(ast)
	// add top level mod to package
	if ast.ParentModule == nil {
		pkg.AddModule(ast)
	}

	for _, sub := range mod.SubModules {
		mod, err := parser.ParseModule(sub, pkg)

		if err != nil {
			r.addError(err)
			continue
		}

		ast.AddModule(mod)
	}
}

func (r *resolver) ResolveImports(pkg *ast.Package) {
	for _, mod := range pkg.Modules {
		for _, file := range mod.Set.Files {
			for _, node := range file.Nodes.Imports {
				r.ResolveDependency(node, file, mod)

			}
		}
	}
}

func (r *resolver) ResolveDependency(decl *ast.ImportDeclaration, file *ast.File, pMod *ast.Module) {
	pkg := pMod.Package
	// Path
	importPath := decl.Path.Value

	// Check length
	if len(importPath) == 0 {
		r.addError(lexer.NewError("empty import path", decl.Range(), file.LexerFile))
		return
	}

	// split path into <pkg>/<mod>/<mod>...
	splitPath := strings.Split(importPath, "/")
	if len(splitPath) < 1 {
		r.addError(lexer.NewError("expected module path, format: <package>/<module>", decl.Range(), file.LexerFile))
		return
	}

	p := splitPath[0]
	var targetPackage *ast.Package

	if p == "std" {
		targetPackage = r.pacakges["std"]
	} else if r.target != nil && r.target.Name() == p {
		targetPackage = r.target
	} else {

		// Resolve package
		dep := pkg.Info.Config.FindDependency(p)

		if dep == nil {
			r.addError(lexer.NewError(fmt.Sprintf("unable to locate package, \"%s\"", p), decl.Range(), file.LexerFile))
			return
		}

		pre, ok := r.pacakges[dep.ID()]

		// Package has already been parsed
		if ok {
			targetPackage = pre
		} else {

			// package has not been parsed, parse
			targetPackage = r.ParsePackage(dep.Path, false)
			r.queuedActions.Push(func() {
				r.ResolveImports(targetPackage)
			})
		}
	}

	if targetPackage == nil {
		r.addError(lexer.NewError(fmt.Sprintf("unable to locates package, \"%s\"", p), decl.Range(), file.LexerFile))
		return
	}

	// resolve target module
	paths := splitPath[1:]
	var mod *ast.Module

	// looking for base module, base module is module at base either named the package name or main
	if len(paths) == 0 {
		x, ok := targetPackage.Modules[targetPackage.Name()]
		if ok {
			mod = x
		} else {
			x, ok = targetPackage.Modules["main"]
			if ok {
				mod = x
			}
		}
	} else {

		// looking for non base module, loop till paths are exhasusted
		for _, path := range paths {

			if mod == nil {
				x, ok := targetPackage.Modules[path]
				// cannot locate
				if !ok {
					r.addError(lexer.NewError(fmt.Sprintf("unable to locate module, \"%s\"", path), decl.Range(), file.LexerFile))
					return
				} else {
					mod = x
				}
			} else {
				x, ok := mod.SubModules[path]

				if !ok {
					r.addError(lexer.NewError(fmt.Sprintf("unable to locate module, \"%s\"", path), decl.Range(), file.LexerFile))
					return
				} else {
					mod = x
				}
			}

		}
	}

	if pMod == mod {
		err := lexer.NewError("cannot import self", decl.Range(), file.LexerFile)
		r.addError(err)
		return
	}

	// cyclic package import
	if pMod.Package != mod.Package {
		cyclic := r.programPackageGraph.HasEdgeFromTo(mod.Package.ID(), pMod.Package.ID())

		if cyclic {
			r.addError(lexer.NewError(fmt.Sprintf("cyclic import between packages, %s & %s", pMod.Package.Name(), mod.Package.Name()), decl.Range(), file.LexerFile))
		}

		// add edge
		pEdge := r.programPackageGraph.NewEdge(pMod.Package, mod.Package)
		r.programPackageGraph.SetEdge(pEdge)

	} else {
		// of the same package, add edge between modules
		pMod.Package.SetEdge(pMod, mod)
	}

	// cyclic module import
	cyclic := r.programModuleGraph.HasEdgeFromTo(mod.ID(), pMod.ID())
	if cyclic {
		r.addError(lexer.NewError(fmt.Sprintf("cyclic import between modules, %s & %s", pMod.Name(), mod.Name()), decl.Range(), file.LexerFile))
	}

	// add edge
	mEdge := r.programModuleGraph.NewEdge(pMod, mod)
	r.programModuleGraph.SetEdge(mEdge)

	// poulate import id
	decl.ImportedModuleID = mod.ID()
}
