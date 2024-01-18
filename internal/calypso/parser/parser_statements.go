package parser

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (p *Parser) parseStatement() (ast.Statement, error) {

	switch p.current() {
	case token.CONST, token.LET:
		return p.parseVariableStatement()
	}

	p.next()
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

	return &ast.VariableStatement{
		Identifier: tok.Lit,
		Value:      expr,
		IsConstant: isConst,
	}, nil

}
