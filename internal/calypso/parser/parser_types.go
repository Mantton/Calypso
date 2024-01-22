package parser

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (p *Parser) parsePossibleTypeExpression() ast.TypeExpression {
	if !p.match(token.COLON) {
		return nil
	}
	// have matched colon, on type expression

	return p.parseTypeExpression()

}

func (p *Parser) parseTypeExpression() ast.TypeExpression {

	switch p.current() {
	case token.IDENTIFIER:
		return p.parseIdentifierTypeExpression()
	case token.LBRACKET:
		return p.parseArrayTypeExpression()
	}
	panic("expected type expression")
}

func (p *Parser) parseArrayTypeExpression() ast.TypeExpression {
	p.expect(token.LBRACKET)
	expr := p.parseTypeExpression()

	switch p.current() {
	case token.COLON:
		p.expect(token.COLON)
		value := p.parseTypeExpression()
		p.expect(token.RBRACKET)

		return &ast.MapTypeExpression{
			Key:   expr,
			Value: value,
		}

	default:
		p.expect(token.RBRACKET)
	}

	return &ast.ArrayTypeExpression{
		Element: expr,
	}
}

func (p *Parser) parseIdentifierTypeExpression() ast.TypeExpression {

	ident := p.parseIdentifier()
	args := []ast.TypeExpression{}
	if p.currentMatches(token.LSS) {
		args = p.parseGenericArgumentClauseExpression()

	}

	if len(args) != 0 {
		return &ast.GenericTypeExpression{
			Identifier: ident,
			Arguments:  args,
		}
	}

	return &ast.IdentifierTypeExpression{
		Identifier: ident,
	}

}

func (p *Parser) parseGenericArgumentClauseExpression() []ast.TypeExpression {

	args := []ast.TypeExpression{}
	p.expect(token.LSS)

	if p.match(token.GTR) {
		panic("expected at least 1 argument")
	}

	// First Argument
	expr := p.parseTypeExpression()
	args = append(args, expr)

	// Check For Others
	for p.match(token.COMMA) {

		if p.match(token.GTR) {
			panic("expected type expression")
		}

		expr := p.parseTypeExpression()

		args = append(args, expr)
	}

	p.expect(token.GTR)

	if len(args) == 0 {
		panic("expected arguments")
	}

	return args
}
