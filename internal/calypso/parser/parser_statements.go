package parser

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (p *Parser) parseStatement() (ast.Statement, error) {
	fmt.Println("STMT:", p.currentScannedToken())

	switch p.current() {
	case token.CONST, token.LET:
		return p.parseVariableStatement()
	case token.IF:
		return p.parseIfStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.IDENTIFIER:
		return p.parseExpressionStatement()
	}

	panic("expected statement")
}

func (p *Parser) parseVariableStatement() (*ast.VariableStatement, error) {
	/**
	let x = `expr`;
	const y = `expr`;
	*/
	isConst := p.current() == token.CONST

	p.next()                          // Move to next token
	tok := p.expect(token.IDENTIFIER) // Parse Ident

	p.expect(token.ASSIGN)

	expr, err := p.parseExpression()

	if err != nil {
		return nil, err
	}

	p.expect(token.SEMICOLON)

	return &ast.VariableStatement{
		Identifier: tok.Lit,
		Value:      expr,
		IsConstant: isConst,
	}, nil

}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	/**
	   {
	  	let x = 10;
	  	print("hello");
	  }
	*/

	// Opening
	p.expect(token.LBRACE)
	statements := p.parseStatementList()
	// Closing
	p.expect(token.RBRACE)

	return &ast.BlockStatement{
		Statements: statements,
	}

}

func (p *Parser) parseIfStatement() (ast.Statement, error) {
	/**
	  if (true) {
		return false;
	  } else {
		return true;
	  }
	*/

	p.expect(token.IF)

	// Condition
	p.expect(token.LPAREN)

	stmt := &ast.IfStatement{}
	condition, err := p.parseExpression()

	if err != nil {
		return nil, err
	}

	stmt.Condition = condition

	p.expect(token.RPAREN)

	// Action Block
	block := p.parseBlockStatement()
	stmt.Action = block

	// Conditional Block

	if p.currentMatches(token.ELSE) {
		p.next()
		alt := p.parseBlockStatement()
		stmt.Alternative = alt
	}

	return stmt, nil

}

func (p *Parser) parseReturnStatement() (ast.Statement, error) {
	p.expect(token.RETURN)

	expr, err := p.parseExpression()

	if err != nil {
		return nil, err
	}

	p.expect(token.SEMICOLON)

	return &ast.ReturnStatement{
		Value: expr,
	}, nil

}

func (p *Parser) parseWhileStatement() (ast.Statement, error) {
	p.expect(token.WHILE)
	// Condition
	p.expect(token.LPAREN)

	stmt := &ast.WhileStatement{}
	condition, err := p.parseExpression()

	if err != nil {
		return nil, err
	}

	stmt.Condition = condition

	p.expect(token.RPAREN)

	// Action Block
	block := p.parseBlockStatement()
	stmt.Action = block

	return stmt, nil
}

func (p *Parser) parseExpressionStatement() (ast.Statement, error) {

	expr, err := p.parseExpression()

	if err != nil {
		return nil, err
	}

	p.expect(token.SEMICOLON)
	return &ast.ExpressionStatement{
		Expr: expr,
	}, nil
}
