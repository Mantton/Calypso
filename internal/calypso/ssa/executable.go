package ssa

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
)

func NewExecutable(file *ast.File) *Executable {
	return &Executable{
		IncludedFile: file,
		Modules:      make(map[string]*Module),
	}
}

type builder struct {
	Mod   *Module
	Fn    *Function
	Block *Block
}

func (n *Executable) Build() {
	b := &builder{
		Mod: NewModule(n.IncludedFile),
	}

	for _, decl := range n.IncludedFile.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDeclaration:
			b.resolveFunction(d.Func)
		}
	}
}

func (b builder) resolveFunction(d *ast.FunctionExpression) {

	fn := &Function{}
	b.Mod.Members[d.Identifier.Value] = fn
	b.Fn = fn
	b.Block = &Block{}
	b.resolveBlockStatement(d.Body, fn)
}

func (b builder) resolveStmt(n ast.Statement, fn *Function) {

	switch n := n.(type) {
	case *ast.VariableStatement:
		b.resolveVariableStmt(n, fn)
		return
	case *ast.IfStatement:
		return
	case *ast.ReturnStatement:
		return
	case *ast.BlockStatement:
		panic("CANNOT BE CALLED DIRECTLY")
	}

	panic(fmt.Sprintf("unknown statement %T\n", n))
}

func (b *builder) resolveVariableStmt(n *ast.VariableStatement, fn *Function) {

	// Alloc
	// Value
	val := b.resolveExpr(n.Value)

	fmt.Println(val)
}
func (b *builder) resolveBlockStatement(n *ast.BlockStatement, fn *Function) {

	for _, s := range n.Statements {
		b.resolveStmt(s, fn)
	}
}

func (b *builder) resolveExpr(n ast.Expression) Value {
	switch n := n.(type) {
	case *ast.IntegerLiteral:
		return &Constant{
			Value: n.Value,
		}
	}

	panic("expr")
}
