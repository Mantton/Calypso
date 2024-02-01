package resolver

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lexer"
)

type Resolver struct {
	Errors  lexer.ErrorList
	symbols *SymbolTable
	depth   int
}

func New() *Resolver {
	return &Resolver{
		symbols: nil,
		depth:   0,
	}
}

// * Scopes
func (r *Resolver) enterScope() {
	r.depth += 1

	if r.depth > 1000 {
		panic("exceeded max scope depth") // TODO : Error
	}

	parent := r.symbols
	child := newSymbolTable(parent)
	r.symbols = child
}

func (r *Resolver) leaveScope(isParent bool) {
	parent := r.symbols.Parent

	if parent != nil {
		r.symbols = parent
		r.depth -= 1

	} else if !isParent {
		panic("cannot exit scope")
	}
}

func (r *Resolver) declare(s *SymbolInfo, e *ast.IdentifierExpression) {
	result := r.symbols.Declare(s)

	if !result {
		msg := fmt.Sprintf("`%s` is already declared in the current scope.", e.Value)
		panic(r.error(msg, e))
	}
}

func (r *Resolver) define(s *SymbolInfo, e *ast.IdentifierExpression) {

	result := r.symbols.Define(s)

	if !result {
		msg := fmt.Sprintf("`%s` is not declared in the current scope.", e.Value)
		panic(r.error(msg, e))
	}
}

func (r *Resolver) expect(e *ast.IdentifierExpression) *SymbolInfo {
	sym, res := r.symbols.Resolve(e.Value)

	if !res {
		msg := fmt.Sprintf("`%s` cannot be found in the current scope.", e.Value)
		panic(r.error(msg, e))
	}

	return sym
}

func (r *Resolver) error(message string, expr ast.Expression) lexer.Error {
	return lexer.Error{
		Range:   expr.Range(),
		Message: message,
	}
}
