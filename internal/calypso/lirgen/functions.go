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

	fmt.Println("\nGeneric Specializations")
	for k, n := range b.Mod.TModule.Table.SpecializedFunctions {
		// Do not generate generic
		if types.IsGeneric(n) {
			continue
		}

		b.registerMonomorphicSpecialization(k, n)
	}

	fmt.Println()

}

func (b *builder) registerFunction(n *ast.FunctionExpression) {
	sg := b.Mod.TModule.Table.Nodes[n].(*types.FunctionSignature)
	tFn := sg.Function

	if tFn == nil {
		panic("function node not type checked")
	}

	if types.IsGeneric(sg) {
		fmt.Println("DEBUG: Generic Function, skipping")
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

func (b *builder) registerMonomorphicSpecialization(k string, ssg *types.SpecializedFunctionSignature) {
	fn := lir.NewFunction(ssg.InstanceOf.Function)
	fn.Name = k
	b.Mod.Functions[fn.Name] = fn
	b.TFunctions[ssg] = fn
	fmt.Println("<FUNCTION>", fn.Name, " , ", ssg.Sg())
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

	// Monomorphization
	for _, n := range b.Mod.TModule.Table.SpecializedFunctions {
		// Do not generate generic
		if types.IsGeneric(n) {
			continue
		}
		b.walkMonomorphization(n)
	}

}

func (b *builder) visitFunction(n *ast.FunctionExpression) {

	sg := b.Mod.TModule.Table.Nodes[n].(*types.FunctionSignature)
	if types.IsGeneric(sg) {
		return
	}

	// Non Generic
	fn := b.Functions[n]
	b.walkFunction(n, fn)
}

func (b *builder) walkFunction(n *ast.FunctionExpression, fn *lir.Function) {
	fn.AddSelf()

	// Parameters
	fmt.Println(fn.Signature(), fn.TFunction.SymbolName())
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

func (b *builder) walkMonomorphization(ssg *types.SpecializedFunctionSignature) {
	fn := b.TFunctions[ssg]
	expr := ssg.InstanceOf.Function.AST()

	if expr == nil {
		panic("NIL")
	}
	b.walkFunction(expr, fn)
}
