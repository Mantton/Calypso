package resolver

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
)

type Resolver struct {
	Errors []string
}

func New() *Resolver {
	return &Resolver{}
}

func (r *Resolver) ResolveFile(file *ast.File) {

	if len(file.Constants) != 0 {

		for _, decl := range file.Constants {
			r.resolveDeclaration(decl)
		}

	}

}
func (r *Resolver) resolveDeclaration(decl ast.Declaration) {
	switch decl := decl.(type) {
	case *ast.ConstantDeclaration:
		panic("resolve const decl")
	case *ast.FunctionDeclaration:
		panic("resolve func decl")
	default:
		fmt.Printf("\n%T", decl)
		panic("resolver: unknown declaration")

	}

}

func (r *Resolver) resolveStatements(statements []ast.Statement) {}
func (r *Resolver) resolveStatement(statement ast.Statement)     {}
func (r *Resolver) resolveExpression(expression ast.Expression)  {}
