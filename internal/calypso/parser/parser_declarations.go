package parser

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (p *Parser) parseDeclaration() ast.Declaration {

	switch p.current() {
	case token.CONST:
		stmt, err := p.parseVariableStatement()

		if err != nil {
			panic(err)
		}

		p.expect(token.SEMICOLON)
		return &ast.ConstantDeclaration{
			Stmt: stmt,
		}

	case token.FUNC:
		return p.parseFunctionDeclaration()
	}

	panic("expected declaration")
}

func (p *Parser) parseFunctionDeclaration() *ast.FunctionDeclaration {

	p.expect(token.FUNC) // Expect current to be `func`, consume

	// Name
	name := p.expect(token.IDENTIFIER).Lit // Function Name

	// Parameters

	p.expect(token.LPAREN)
	// TODO: Parse Parameters
	p.expect(token.RPAREN)

	// Body
	body := p.parseFunctionBody()

	lit := &ast.FunctionLiteral{
		Name: name,
		Body: body,
	}

	return &ast.FunctionDeclaration{
		Func: lit,
	}
}

func (p *Parser) parseFunctionBody() *ast.BlockStatement {
	// Opening
	p.expect(token.LBRACE)
	statements := p.parseStatementList()
	// Closing
	p.expect(token.RBRACE)

	return &ast.BlockStatement{
		Statements: statements,
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
					p.advance(token.IsStatement)
				}
			}()

			statement, err := p.parseStatement()

			if err != nil {
				panic(err)
			}
			p.expect(token.SEMICOLON)

			list = append(list, statement)
		}()

	}

	return list

}
