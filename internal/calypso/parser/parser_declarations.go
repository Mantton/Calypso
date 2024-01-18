package parser

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (p *Parser) parseDeclaration() ast.Declaration {

	switch p.current() {
	case token.CONST:
		panic("parse constant")
	case token.LET:
		panic("parse variable")
	case token.FUNC:
		return p.parseFunctionDeclaration()
	}

	panic("expected declaration")
}

func (p *Parser) parseFunctionDeclaration() *ast.FunctionDeclaration {

	p.expect(token.FUNC) // Expect current to be `func`, consume

	// Name
	p.expect(token.IDENTIFIER) // TODO: Parse Function Name

	// Parameters

	p.expect(token.LPAREN)
	// TODO: Parse Parameters
	p.expect(token.RPAREN)

	// Body
	body := p.parseFunctionBody()

	return &ast.FunctionDeclaration{
		Name: "main",
		Body: body,
	}
}

func (p *Parser) parseFunctionBody() *ast.BlockStatement {
	// Opening
	p.expect(token.LBRACE)
	statements := p.parseStatementList()
	fmt.Println(statements)
	// Closing
	p.expect(token.RBRACE)

	return &ast.BlockStatement{
		Statements: statements,
	}

}

func (p *Parser) parseStatementList() []ast.Statement {

	var list = []ast.Statement{}
	for p.current() != token.RBRACE && p.current() != token.EOF {

		defer func() {
			if r := recover(); r != nil {
				fmt.Print("ERROR ", r)
			}
		}()

		statement, err := p.parseStatement()

		if err != nil {
			panic(err)
		}

		list = append(list, statement)
		p.expect(token.SEMICOLON)
	}

	return list

}

func (p *Parser) parseStatement() (ast.Statement, error) {

	switch p.current() {
	case token.CONST, token.LET:
		/**
		let x = `expr`;
		const y = `expr`;
		*/
		p.next()                   // Move to next token
		p.expect(token.IDENTIFIER) // TODO: Parse Ident
		p.expect(token.ASSIGN)

		expr, err := p.parseExpression()

		if err != nil {
			return nil, err
		}

		return &ast.LetStatement{
			Ident: "name",
			Value: expr,
		}, nil
	}

	panic("statement")
}
