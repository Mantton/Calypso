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

	lit := &ast.FunctionLiteral{
		Name: "main",
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

func (p *Parser) parseStatement() (ast.Statement, error) {

	switch p.current() {
	case token.CONST, token.LET:
		/**
		let x = `expr`;
		const y = `expr`;
		*/
		p.next()                          // Move to next token
		tok := p.expect(token.IDENTIFIER) // Parse Ident

		p.expect(token.ASSIGN)

		expr, err := p.parseExpression()

		if err != nil {
			return nil, err
		}

		if token.CONST == p.current() {
			// TODO: Const Statement
			return &ast.LetStatement{
				Ident: tok.Lit,
				Value: expr,
			}, nil
		} else {

			return &ast.LetStatement{
				Ident: tok.Lit,
				Value: expr,
			}, nil
		}

	}

	p.next()
	panic("bad statement")
}
