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
	start := p.expect(token.LBRACKET)
	expr := p.parseTypeExpression()
	var end token.TokenPosition
	switch p.current() {
	case token.COLON:
		p.expect(token.COLON)
		value := p.parseTypeExpression()
		end := p.expect(token.RBRACKET)

		return &ast.MapTypeExpression{
			Key:         expr,
			Value:       value,
			LBracketPos: start.Pos,
			RBracketPos: end.Pos,
		}

	default:
		end = p.expect(token.RBRACKET).Pos
	}

	return &ast.ArrayTypeExpression{
		Element:     expr,
		LBracketPos: start.Pos,
		RBracketPos: end,
	}
}

func (p *Parser) parseIdentifierTypeExpression() ast.TypeExpression {

	ident := p.parseIdentifier()
	args := []ast.TypeExpression{}
	var start, end token.TokenPosition
	if p.currentMatches(token.LSS) {
		args, start, end = p.parseGenericArgumentClauseExpression()
	}

	if len(args) != 0 {
		return &ast.GenericTypeExpression{
			Identifier:  ident,
			Arguments:   args,
			LChevronPos: start,
			RChevronPos: end,
		}
	}

	return &ast.IdentifierTypeExpression{
		Identifier: ident,
	}

}

func (p *Parser) parseGenericArgumentClauseExpression() ([]ast.TypeExpression, token.TokenPosition, token.TokenPosition) {

	args := []ast.TypeExpression{}
	start := p.expect(token.LSS)

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

	end := p.expect(token.GTR)

	if len(args) == 0 {
		panic("expected arguments")
	}

	return args, start.Pos, end.Pos
}
