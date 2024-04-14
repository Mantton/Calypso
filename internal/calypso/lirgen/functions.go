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
	tFn := b.Mod.TModule.Table.GetFunction(n)

	if tFn == nil {
		panic("function node not type checked")
	}

	sg := tFn.Sg()
	if types.IsGeneric(sg) {
		instances := sg.Instances
		for _, v := range instances {
			name := tFn.SymbolName() + "::" + v.InstanceHash
			// fmt.Println("Instance", v, "\n", v.InstanceHash)
			fn := lir.NewFunction(tFn)
			fn.Name = name
			fn.SetSignature(v)
			b.TFunctions[v] = fn
			b.Mod.Functions[name] = fn
			fmt.Println("<FUNCTION>", name, sg)
		}
		return
	}

	name := n.Identifier.Value

	fn := lir.NewFunction(tFn)
	fn.Name = tFn.SymbolName()
	b.Functions[n] = fn
	b.TFunctions[sg] = fn
	b.Mod.Functions[fn.Name] = fn
	fn.External = tFn.Target != nil

	fmt.Println("<FUNCTION>", name, sg)
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

	tFn := b.Mod.TModule.Table.GetFunction(n)

	sg := tFn.Sg()
	if types.IsGeneric(sg) {
		for _, instance := range sg.Instances {
			fn := b.TFunctions[instance].(*lir.Function)
			b.walkFunction(n, fn)
		}
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
