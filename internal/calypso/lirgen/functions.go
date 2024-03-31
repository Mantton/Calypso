package lirgen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/types"
)

// External Functions
func (b *builder) pass2(f *ast.File) {

	for _, decl := range f.Nodes.ExternalFunctions {
		for _, fn := range decl.Signatures {
			f := b.Mod.TModule.Table.GetFunction(fn.Func)

			if f == nil {
				panic("function not found")
			}

			b.addExternalFunction(f)

		}
	}
}

func (b *builder) addExternalFunction(fn *types.Function) {

	if b.Mod.ExternalFunctions == nil {
		b.Mod.ExternalFunctions = make(map[string]*types.FunctionSignature)
	}
	b.Mod.ExternalFunctions[fn.Name()] = fn.Sg()
}

// Register All Functions
func (b *builder) pass3(f *ast.File) {
	// General
	for _, fn := range f.Nodes.Functions {
		b.registerFunction(fn.Func)
	}

	// Extensions
	for _, n := range f.Nodes.Extensions {
		for _, fn := range n.Content {
			b.registerFunction(fn.Func)
		}
	}

	// Conformances
	for _, n := range f.Nodes.Conformances {
		for _, fn := range n.Signatures {
			b.registerFunction(fn.Func)
		}
	}
}

func (b *builder) registerFunction(n *ast.FunctionExpression) {
	tFn := b.Mod.TModule.Table.GetFunction(n)

	if tFn == nil {
		panic("function node not type checked")
	}

	fn := lir.NewFunction(tFn)
	b.Functions[n] = fn
	b.Mod.Functions[n.Identifier.Value] = fn
	fmt.Println("<FUNCTION>", tFn.Sg())
}

// Populate Bodies?
func (b *builder) pass4(f *ast.File) {
	// General
	for _, fn := range f.Nodes.Functions {
		b.visitFunction(fn.Func)
	}

	// Extensions
	for _, n := range f.Nodes.Extensions {
		for _, fn := range n.Content {
			b.visitFunction(fn.Func)
		}
	}

	// Conformances
	for _, n := range f.Nodes.Conformances {
		for _, fn := range n.Signatures {
			b.visitFunction(fn.Func)
		}
	}
}

func (b *builder) visitFunction(n *ast.FunctionExpression) {
	fn := b.Functions[n]

	// Parameters
	for _, p := range fn.Symbol.Parameters {
		fn.AddParameter(p)
	}

	// Body
	stmts := n.Body.Statements

	if len(stmts) == 0 {
		fn.Emit(&lir.ReturnVoid{})
		return
	}

	// Statements
	for _, stmt := range stmts {
		b.visitStatement(stmt, fn)
	}
}
