package parser

import (
	"errors"
	"strconv"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (p *Parser) parseExpression() (ast.Expression, error) {

	return p.parseBinaryExpression()

}

func (p *Parser) parseBinaryExpression() (ast.Expression, error) {
	return p.parseEqualityExpression()
}

func (p *Parser) parseEqualityExpression() (ast.Expression, error) {
	expr, err := p.parseComparisonExpression()
	if err != nil {
		return nil, err
	}
	for p.match(token.ASSIGN, token.EQL) {
		op := p.previous()
		right, err := p.parseComparisonExpression()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpression{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}

	return expr, nil
}

func (p *Parser) parseComparisonExpression() (ast.Expression, error) {
	expr, err := p.parseTermExpression()
	if err != nil {
		return nil, err
	}
	for p.match(token.GTR, token.GEQ, token.LEQ, token.LSS) {
		op := p.previous()
		right, err := p.parseTermExpression()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpression{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}

	return expr, nil
}

func (p *Parser) parseTermExpression() (ast.Expression, error) {
	expr, err := p.parseFactorExpression()
	if err != nil {
		return nil, err
	}
	for p.match(token.SUB, token.ADD) {
		op := p.previous()
		right, err := p.parseFactorExpression()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpression{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}

	return expr, nil
}

func (p *Parser) parseFactorExpression() (ast.Expression, error) {
	expr, err := p.parseUnaryExpression()
	if err != nil {
		return nil, err
	}
	for p.match(token.QUO, token.MUL) {
		op := p.previous()
		right, err := p.parseUnaryExpression()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpression{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}

	return expr, nil
}

func (p *Parser) parseUnaryExpression() (ast.Expression, error) {

	if p.match(token.NOT, token.SUB) {
		op := p.previous()
		right, err := p.parseUnaryExpression()

		if err != nil {
			return nil, err
		}

		return &ast.UnaryExpression{
			Op:   op,
			Expr: right,
		}, nil
	}
	return p.parsePrimaryExpression()
}

func (p *Parser) parsePrimaryExpression() (ast.Expression, error) {
	var expr ast.Expression
	switch p.current() {
	case token.FALSE:
		expr = &ast.BooleanLiteral{Value: false}
	case token.TRUE:
		expr = &ast.BooleanLiteral{Value: true}
	case token.NULL:
		expr = &ast.NullLiteral{}
	case token.VOID:
		expr = &ast.VoidLiteral{}
	case token.INTEGER:
		v, err := strconv.ParseInt(p.currentScannedToken().Lit, 10, 64)

		if err != nil {
			return nil, err
		}
		expr = &ast.IntegerLiteral{
			Value: int(v),
		}
	case token.FLOAT:
		v, err := strconv.ParseFloat(p.currentScannedToken().Lit, 64)
		if err != nil {
			return nil, err
		}
		expr = &ast.FloatLiteral{
			Value: v,
		}
	case token.STRING:
		expr = &ast.StringLiteral{
			Value: p.currentScannedToken().Lit,
		}
	case token.IDENTIFIER:
		expr = &ast.IdentifierLiteral{
			Value: p.currentScannedToken().Lit,
		}

	case token.LPAREN:
		p.next() // Move to next token
		expr, err := p.parseExpression()

		if err != nil {
			return nil, err
		}

		p.expect(token.RPAREN)
		return &ast.GroupedExpression{
			Expr: expr,
		}, nil
	}

	if expr != nil {
		p.next()
		return expr, nil

	}
	return nil, errors.New("expected expression")
}
