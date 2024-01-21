package parser

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (p *Parser) parseStatement() (ast.Statement, error) {
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

	panic(p.error("expected statement"))
}

func (p *Parser) parseVariableStatement() (*ast.VariableStatement, error) {
	/**
	let x = `expr`;
	const y = `expr`;
	const z :int = `expr`;
	*/
	isConst := p.current() == token.CONST

	p.next()                          // Move to next token
	tok := p.expect(token.IDENTIFIER) // Parse Ident

	// Parse Type Expression If Found
	t := p.parsePossibleTypeExpression()

	p.expect(token.ASSIGN)

	expr := p.parseExpression()

	p.expect(token.SEMICOLON)

	return &ast.VariableStatement{
		Identifier:     tok.Lit,
		Value:          expr,
		IsConstant:     isConst,
		TypeAnnotation: t,
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
	condition := p.parseExpression()

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

	expr := p.parseExpression()

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
	condition := p.parseExpression()

	stmt.Condition = condition

	p.expect(token.RPAREN)

	// Action Block
	block := p.parseBlockStatement()
	stmt.Action = block

	return stmt, nil
}

func (p *Parser) parseExpressionStatement() (ast.Statement, error) {

	expr := p.parseExpression()

	switch expr := expr.(type) {
	case *ast.AssignmentExpression, *ast.CallExpression:
		p.expect(token.SEMICOLON)
		return &ast.ExpressionStatement{
			Expr: expr,
		}, nil
	default:
		panic(p.error("expected statement, not expression"))
	}

}
