package parser

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (p *Parser) parseExpression() (ast.Expression, error) {

	return p.parseBinaryExpression()

}

func (p *Parser) parseBinaryExpression() (ast.Expression, error) {
	return p.parseAssignmentExpression()
}

func (p *Parser) parseAssignmentExpression() (ast.Expression, error) {
	expr, err := p.parseEqualityExpression()

	if err != nil {
		return nil, err
	}

	if p.match(token.ASSIGN) {
		value, err := p.parseAssignmentExpression()
		if err != nil {
			return nil, err
		}
		return &ast.AssignmentExpression{
			Ident: expr,
			Value: value,
		}, nil
	}

	return expr, nil
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
	return p.parseCallExpression()
}

func (p *Parser) parseCallExpression() (ast.Expression, error) {
	expr, err := p.parsePrimaryExpression()

	if err != nil {
		return nil, err
	}

	if p.currentMatches(token.LPAREN) {

		list, err := p.parseExpressionList(token.LPAREN, token.RPAREN)

		if err != nil {
			return nil, err
		}

		return &ast.CallExpression{
			Target:    expr,
			Arguments: list,
		}, nil
	}

	return expr, nil
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

func (p *Parser) parseExpressionList(start, end token.Token) ([]ast.Expression, error) {
	list := []ast.Expression{}

	// expect start token
	p.expect(start)

	// if immediately followed by end token, return
	if p.match(end) {
		return list, nil
	}

	expr, err := p.parseExpression()

	if err != nil {
		return nil, err
	}

	list = append(list, expr)

	for p.match(token.COMMA) {
		expr, err := p.parseExpression()

		// TODO: Report Error

		if err != nil {
			fmt.Println(err)
			p.next()
		}

		list = append(list, expr)
	}

	p.expect(end)

	return list, nil
}

func (p *Parser) parseFunctionLiteral() (*ast.FunctionLiteral, error) {
	p.expect(token.FUNC) // Expect current to be `func`, consume

	// Name
	name := p.expect(token.IDENTIFIER).Lit // Function Name

	// Parameters
	params := p.parseFunctionParameters()

	if len(params) > 99 {
		panic("too many parameters")
	}

	// Body
	body := p.parseFunctionBody()

	return &ast.FunctionLiteral{
		Name:   name,
		Body:   body,
		Params: params,
	}, nil
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

func (p *Parser) parseFunctionParameters() []*ast.IdentifierLiteral {
	identifiers := []*ast.IdentifierLiteral{}

	p.expect(token.LPAREN)

	// if immediately followed by end token, return
	if p.match(token.RPAREN) {
		return identifiers
	}

	expr := p.parseIdentifier()
	identifiers = append(identifiers, expr)

	for p.match(token.COMMA) {
		expr := p.parseIdentifier()

		identifiers = append(identifiers, expr)
	}

	p.expect(token.RPAREN)

	return identifiers
}

func (p *Parser) parseIdentifier() *ast.IdentifierLiteral {

	tok := p.expect(token.IDENTIFIER)

	return &ast.IdentifierLiteral{
		Value: tok.Lit,
	}
}
