package lirgen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/types"
)

// Register All Functions
func (b *builder) pass2(f *ast.File) {
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

	// External
	for _, n := range f.Nodes.ExternalFunctions {
		for _, fn := range n.Signatures {
			b.registerFunction(fn.Func)
		}
	}
}

func (b *builder) registerFunction(n *ast.FunctionExpression) {
	sg := b.Mod.TModule.Table.Nodes[n].(*types.FunctionSignature)
	tFn := sg.Function

	if tFn == nil {
		panic("function node not type checked")
	}

	if types.IsGeneric(sg) {
		b.registerMonomorphicSpecializations(sg.Function)
		return
	}

	name := n.Identifier.Value

	fn := lir.NewFunction(tFn)
	fn.Name = tFn.SymbolName()      // set function name to symbol name
	b.Functions[n] = fn             // map node to function
	b.TFunctions[tFn.Sg()] = fn     // map sg to function
	b.Mod.Functions[fn.Name] = fn   // add function to module
	fn.External = tFn.Target != nil // mark target

	fmt.Println("<FUNCTION>", name, sg)
}

func (b *builder) registerMonomorphicSpecializations(fn *types.Function) {

	// 1 - Instantiate Generic Funciton
	gFn := lir.NewGenericFunction(fn)

	// 2 - Loop through generic instances & create new functions
	for _, ssg := range fn.AllSpecs() {
		fmt.Println("<FUNCTION>", fn.Name(), " , ", ssg.Sg())

		sFn := lir.NewFunction(fn)
		sFn.Spec = ssg
		sFn.Name = ssg.SymbolName()
		gFn.Specs[sFn.Name] = sFn
	}

	b.Mod.GFunctions[fn.SymbolName()] = gFn
}

// Populate Bodies?
func (b *builder) pass3(f *ast.File) {
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

	// External
	for _, n := range f.Nodes.ExternalFunctions {
		for _, fn := range n.Signatures {
			b.visitFunction(fn.Func)
		}
	}
}

func (b *builder) visitFunction(n *ast.FunctionExpression) {
	sg := b.Mod.TModule.Table.Nodes[n].(*types.FunctionSignature)
	if types.IsGeneric(sg) {
		b.walkMonomorphizations(sg.Function)
		return
	}

	// Non Generic
	fn := b.Functions[n]
	b.walkFunction(n, fn)
}

func (b *builder) walkFunction(n *ast.FunctionExpression, fn *lir.Function) {
	fn.AddSelf()

	// Parameters
	for _, p := range fn.Signature().Parameters {
		fn.AddParameter(p)
	}

	if fn.External {
		return
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

func (b *builder) walkMonomorphizations(fn *types.Function) {
	expr := fn.AST()
	gFn := b.Mod.GFunctions[fn.SymbolName()]

	fmt.Println(gFn)
	for _, ssg := range fn.AllSpecs() {
		sFn := gFn.Specs[ssg.SymbolName()]

		fmt.Println("<WALK_FUNCTION>", fn.Name(), ssg.Sg())
		b.walkFunction(expr, sFn)
	}
}
