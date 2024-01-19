package parser

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (p *Parser) parseDeclaration() ast.Declaration {
	fmt.Println("DECL:", p.currentScannedToken())

	switch p.current() {
	case token.CONST:
		stmt, err := p.parseVariableStatement()

		if err != nil {
			panic(err)
		}

		return &ast.ConstantDeclaration{
			Stmt: stmt,
		}

	case token.FUNC:
		return p.parseFunctionDeclaration()
	}

	panic("expected declaration")
}

func (p *Parser) parseFunctionDeclaration() *ast.FunctionDeclaration {

	fn, err := p.parseFunctionLiteral()

	if err != nil {
		panic(err)
	}

	return &ast.FunctionDeclaration{
		Func: fn,
	}
}

func (p *Parser) parseStatementList() []ast.Statement {

	var list = []ast.Statement{}
	cancel := false
	for p.current() != token.RBRACE && p.current() != token.EOF && !cancel {

		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("STMT ERROR: ", r)
					hasMoved := p.advance(token.IsStatement)

					// avoid infinite loop
					if !hasMoved {
						p.next()
					}
				}
			}()

			statement, err := p.parseStatement()

			if err != nil {
				panic(err)
			}

			list = append(list, statement)
		}()

	}

	return list

}
