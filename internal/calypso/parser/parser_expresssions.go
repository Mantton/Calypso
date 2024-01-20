package parser

import (
	"fmt"
	"strconv"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (p *Parser) parseExpression() ast.Expression {

	return p.parseBinaryExpression()

}

func (p *Parser) parseBinaryExpression() ast.Expression {
	return p.parseAssignmentExpression()
}

func (p *Parser) parseAssignmentExpression() ast.Expression {
	expr := p.parseEqualityExpression()

	if p.match(token.ASSIGN) {
		value := p.parseAssignmentExpression()

		return &ast.AssignmentExpression{
			Target: expr,
			Value:  value,
		}
	}

	return expr
}

func (p *Parser) parseEqualityExpression() ast.Expression {
	expr := p.parseComparisonExpression()

	for p.match(token.NEQ, token.EQL) {
		op := p.previous()
		right := p.parseComparisonExpression()

		expr = &ast.BinaryExpression{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseComparisonExpression() ast.Expression {
	expr := p.parseTermExpression()

	for p.match(token.GTR, token.GEQ, token.LEQ, token.LSS) {
		op := p.previous()
		right := p.parseTermExpression()

		expr = &ast.BinaryExpression{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseTermExpression() ast.Expression {
	expr := p.parseFactorExpression()

	for p.match(token.SUB, token.ADD) {
		op := p.previous()
		right := p.parseFactorExpression()

		expr = &ast.BinaryExpression{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseFactorExpression() ast.Expression {
	expr := p.parseUnaryExpression()

	for p.match(token.QUO, token.MUL) {
		op := p.previous()
		right := p.parseUnaryExpression()

		expr = &ast.BinaryExpression{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseUnaryExpression() ast.Expression {

	if p.match(token.NOT, token.SUB) {
		op := p.previous()
		right := p.parseUnaryExpression()

		return &ast.UnaryExpression{
			Op:   op,
			Expr: right,
		}
	}
	return p.parseCallExpression()
}

func (p *Parser) parseCallExpression() ast.Expression {
	expr := p.parsePropertyExpression()

	if p.currentMatches(token.LPAREN) {

		list := p.parseExpressionList(token.LPAREN, token.RPAREN)

		return &ast.CallExpression{
			Target:    expr,
			Arguments: list,
		}
	}

	return expr
}

func (p *Parser) parsePropertyExpression() ast.Expression {
	expr := p.parseIndexExpression()

	if p.match(token.PERIOD) {
		prop := p.parseIdentifier()

		return &ast.PropertyExpression{
			Target:   expr,
			Property: prop,
		}
	}

	return expr

}

func (p *Parser) parseIndexExpression() ast.Expression {
	expr := p.parsePrimaryExpression()

	if p.currentMatches(token.LBRACKET) {
		p.expect(token.LBRACKET)
		idx := p.parseExpression()

		p.expect(token.RBRACKET)

		return &ast.IndexExpression{
			Target: expr,
			Index:  idx,
		}
	}

	return expr
}

func (p *Parser) parsePrimaryExpression() ast.Expression {
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
			return nil
		}
		expr = &ast.IntegerLiteral{
			Value: int(v),
		}
	case token.FLOAT:
		v, err := strconv.ParseFloat(p.currentScannedToken().Lit, 64)
		if err != nil {
			return nil
		}
		expr = &ast.FloatLiteral{
			Value: v,
		}
	case token.STRING:
		expr = &ast.StringLiteral{
			Value: p.currentScannedToken().Lit,
		}
	case token.IDENTIFIER:
		expr = &ast.IdentifierExpression{
			Value: p.currentScannedToken().Lit,
		}

	case token.LPAREN:
		p.next() // Move to next token
		expr := p.parseExpression()

		p.expect(token.RPAREN)
		return &ast.GroupedExpression{
			Expr: expr,
		}

	case token.LBRACKET:
		return p.parseArrayLit()
	case token.LBRACE:
		return p.parseMapLiteral()
	}

	if expr != nil {
		p.next()
		return expr

	}

	msg := fmt.Sprintf("expected expression, got `%s`", p.currentScannedToken().Lit)
	panic(p.error(msg))
}

func (p *Parser) parseExpressionList(start, end token.Token) []ast.Expression {
	list := []ast.Expression{}

	// expect start token
	p.expect(start)

	// if immediately followed by end token, return
	if p.match(end) {
		return list
	}

	expr := p.parseExpression()

	list = append(list, expr)

	for p.match(token.COMMA) {
		expr := p.parseExpression()

		// TODO: Report Individual Errors

		list = append(list, expr)
	}

	p.expect(end)

	return list
}

func (p *Parser) parseFunctionExpression() *ast.FunctionExpression {
	p.expect(token.FUNC) // Expect current to be `func`, consume

	// Name
	name := p.expect(token.IDENTIFIER).Lit // Function Name

	// Parameters
	params := p.parseFunctionParameters()

	if len(params) > 99 {
		panic(p.error("too many parameters, maximum of 99"))
	}

	// Body
	body := p.parseFunctionBody()

	return &ast.FunctionExpression{
		Name:   name,
		Body:   body,
		Params: params,
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

func (p *Parser) parseFunctionParameters() []*ast.IdentifierExpression {
	identifiers := []*ast.IdentifierExpression{}

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

func (p *Parser) parseIdentifier() *ast.IdentifierExpression {

	tok := p.expect(token.IDENTIFIER)

	return &ast.IdentifierExpression{
		Value: tok.Lit,
	}
}

func (p *Parser) parseArrayLit() *ast.ArrayLiteral {

	elements := p.parseExpressionList(token.LBRACKET, token.RBRACKET)

	return &ast.ArrayLiteral{
		Elements: elements,
	}
}

func (p *Parser) parseMapLiteral() *ast.MapLiteral {
	lit := &ast.MapLiteral{
		Pairs: make(map[ast.Expression]ast.Expression),
	}
	p.expect(token.LBRACE)

	// closes immediately
	if p.match(token.RBRACE) {
		return lit
	}
	// Loop until match with RBRACE
	for !p.match(token.RBRACE) {

		// Parse Key
		key := p.parseExpression()

		// Parse Colon Divider
		p.expect(token.COLON)

		// Parse Value

		value := p.parseExpression()

		lit.Pairs[key] = value

		if p.currentMatches(token.RBRACE) {
			break
		} else {
			p.expect(token.COMMA)
		}

	}

	// expect closing brace
	p.expect(token.RBRACE)

	return lit
}
