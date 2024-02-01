package resolver

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lexer"
)

func (r *Resolver) resolveDeclaration(decl ast.Declaration) {

	func() {
		defer func() {
			if err := recover(); err != nil {
				if err, y := err.(lexer.Error); y {
					r.Errors.Add(err)
				} else {
					panic(r)
				}
			}
		}()
		switch decl := decl.(type) {
		case *ast.ConstantDeclaration:
			stmt := decl.Stmt
			r.resolveStatement(stmt)
		case *ast.FunctionDeclaration:
			expr := decl.Func
			r.resolveExpression(expr)
		case *ast.StatementDeclaration:
			stmt := decl.Stmt
			r.resolveStatement(stmt)
		case *ast.StandardDeclaration:
			s := newSymbolInfo(decl.Identifier.Value, StandardSymbol)
			r.declare(s, decl.Identifier)
			r.define(s, decl.Identifier)
		case *ast.ExtensionDeclaration:
			s := r.expect(decl.Identifier)

			if s.Type != StandardSymbol && s.Type != TypeSymbol {
				msg := fmt.Sprintf("`%s` is not extendable (non-standard || non-type)", decl.Identifier.Value)
				panic(r.error(msg, decl.Identifier))
			}
		case *ast.ConformanceDeclaration:
			s := r.expect(decl.Standard)

			if s.Type != StandardSymbol {
				msg := fmt.Sprintf("`%s` is not a standard.", decl.Standard.Value)
				panic(r.error(msg, decl.Standard))
			}

			s = r.expect(decl.Target)

			if s.Type != StructSymbol && s.Type != TypeSymbol {
				msg := fmt.Sprintf("`%s` cannot be conformed to a standard (non-conformable)", decl.Standard.Value)
				panic(r.error(msg, decl.Standard))
			}

		case *ast.TypeDeclaration:
			s := newSymbolInfo(decl.Identifier.Value, TypeSymbol)
			r.declare(s, decl.Identifier)
			r.define(s, decl.Identifier)

		default:
			msg := fmt.Sprintf("expression declaration not implemented, %T", decl)
			panic(msg)
		}
	}()

}
