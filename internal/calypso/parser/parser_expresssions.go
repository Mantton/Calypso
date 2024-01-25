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
		pos := p.previousScannedToken().Pos
		value := p.parseAssignmentExpression()

		return &ast.AssignmentExpression{
			Target: expr,
			Value:  value,
			OpPos:  pos,
		}
	}

	return expr
}

func (p *Parser) parseEqualityExpression() ast.Expression {
	expr := p.parseComparisonExpression()

	for p.match(token.NEQ, token.EQL) {
		op := p.previousScannedToken()
		right := p.parseComparisonExpression()

		expr = &ast.BinaryExpression{
			Left:  expr,
			Op:    op.Tok,
			OpPos: op.Pos,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseComparisonExpression() ast.Expression {
	expr := p.parseTermExpression()

	for p.match(token.GTR, token.GEQ, token.LEQ, token.LSS) {
		op := p.previousScannedToken()
		right := p.parseTermExpression()

		expr = &ast.BinaryExpression{
			Left:  expr,
			Op:    op.Tok,
			OpPos: op.Pos,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseTermExpression() ast.Expression {
	expr := p.parseFactorExpression()

	for p.match(token.SUB, token.ADD) {
		op := p.previousScannedToken()
		right := p.parseFactorExpression()

		expr = &ast.BinaryExpression{
			Left:  expr,
			Op:    op.Tok,
			OpPos: op.Pos,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseFactorExpression() ast.Expression {
	expr := p.parseUnaryExpression()

	for p.match(token.QUO, token.MUL) {
		op := p.previousScannedToken()
		right := p.parseUnaryExpression()

		expr = &ast.BinaryExpression{
			Left:  expr,
			Op:    op.Tok,
			OpPos: op.Pos,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseUnaryExpression() ast.Expression {

	if p.match(token.NOT, token.SUB) {
		op := p.previousScannedToken()
		right := p.parseUnaryExpression()

		return &ast.UnaryExpression{
			Op:         op.Tok,
			OpPosition: op.Pos,
			Expr:       right,
		}
	}
	return p.parseCallExpression()
}

func (p *Parser) parseCallExpression() ast.Expression {
	expr := p.parsePropertyExpression()

	if p.currentMatches(token.LPAREN) {

		list, start, end := p.parseExpressionList(token.LPAREN, token.RPAREN)

		return &ast.CallExpression{
			Target:    expr,
			Arguments: list,
			LParenPos: start,
			RParenPos: end,
		}
	}

	return expr
}

func (p *Parser) parsePropertyExpression() ast.Expression {
	expr := p.parseIndexExpression()

	if p.match(token.PERIOD) {
		opPos := p.previousScannedToken().Pos
		prop := p.parseIdentifier(false)

		return &ast.PropertyExpression{
			Target:   expr,
			Property: prop,
			DotPos:   opPos,
		}
	}

	return expr

}

func (p *Parser) parseIndexExpression() ast.Expression {
	expr := p.parsePrimaryExpression()

	if p.match(token.LBRACKET) {
		lbrackPos := p.previousScannedToken().Pos
		idx := p.parseExpression()

		rbrackPos := p.expect(token.RBRACKET).Pos

		return &ast.IndexExpression{
			Target:      expr,
			Index:       idx,
			LBracketPos: lbrackPos,
			RBracketPos: rbrackPos,
		}
	}

	return expr
}

func (p *Parser) parsePrimaryExpression() ast.Expression {
	var expr ast.Expression
	switch p.current() {
	case token.FALSE:
		expr = &ast.BooleanLiteral{
			Value: false,
			Pos:   p.currentScannedToken().Pos,
		}
	case token.TRUE:
		expr = &ast.BooleanLiteral{
			Value: true,
			Pos:   p.currentScannedToken().Pos,
		}
	case token.NULL:
		expr = &ast.NullLiteral{
			Pos: p.currentScannedToken().Pos,
		}
	case token.VOID:
		expr = &ast.VoidLiteral{
			Pos: p.currentScannedToken().Pos,
		}
	case token.INTEGER:
		v, err := strconv.ParseInt(p.currentScannedToken().Lit, 10, 64)

		if err != nil {
			return nil
		}
		expr = &ast.IntegerLiteral{
			Value: int(v),
			Pos:   p.currentScannedToken().Pos,
		}
	case token.FLOAT:
		v, err := strconv.ParseFloat(p.currentScannedToken().Lit, 64)
		if err != nil {
			return nil
		}
		expr = &ast.FloatLiteral{
			Value: v,
			Pos:   p.currentScannedToken().Pos,
		}
	case token.STRING:
		expr = &ast.StringLiteral{
			Value: p.currentScannedToken().Lit,
			Pos:   p.currentScannedToken().Pos,
		}
	case token.IDENTIFIER:
		expr = &ast.IdentifierExpression{
			Value: p.currentScannedToken().Lit,
			Pos:   p.currentScannedToken().Pos,
		}

	case token.LPAREN:
		start := p.expect(token.LPAREN).Pos
		expr := p.parseExpression()

		end := p.expect(token.RPAREN).Pos
		return &ast.GroupedExpression{
			LParenPos: start,
			Expr:      expr,
			RParenPos: end,
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

func (p *Parser) parseExpressionList(start, end token.Token) ([]ast.Expression, token.TokenPosition, token.TokenPosition) {
	list := []ast.Expression{}

	// expect start token
	s := p.expect(start)

	// if immediately followed by end token, return
	if p.currentMatches(end) {
		e := p.expect(end)
		return list, s.Pos, e.Pos
	}

	expr := p.parseExpression()

	list = append(list, expr)

	for p.match(token.COMMA) {
		expr := p.parseExpression()

		// TODO: Report Individual Errors

		list = append(list, expr)
	}

	e := p.expect(end)

	return list, s.Pos, e.Pos
}

func (p *Parser) parseFunctionExpression(requiresBody bool) *ast.FunctionExpression {
	start := p.expect(token.FUNC) // Expect current to be `func`, consume

	// Name
	ident := p.parseIdentifier(false)

	var genParams *ast.GenericParametersClause

	if p.currentMatches(token.LSS) {
		genParams = p.parseGenericParameterClause()
	}

	// Parameters
	params := p.parseFunctionParameters()

	if len(params) > 99 {
		panic(p.error("too many parameters, maximum of 99"))
	}
	// Return Type
	retType := p.parseFunctionReturnType()

	// Body

	var body *ast.BlockStatement

	if p.currentMatches(token.LBRACE) {
		body = p.parseFunctionBody()
	} else if requiresBody {
		panic(p.error("expected function body"))
	} else {
		p.expect(token.SEMICOLON)
	}

	return &ast.FunctionExpression{
		KeyWPos:       start.Pos,
		Identifier:    ident,
		Body:          body,
		Params:        params,
		ReturnType:    retType,
		GenericParams: genParams,
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

func (p *Parser) parseFunctionReturnType() ast.TypeExpression {
	if p.match(token.R_ARROW) {
		annotatedType := p.parseTypeExpression()
		return annotatedType
	} else {
		return nil
	}
}

func (p *Parser) parseFunctionParameters() []*ast.IdentifierExpression {
	identifiers := []*ast.IdentifierExpression{}

	p.expect(token.LPAREN)

	// if immediately followed by end token, return
	if p.match(token.RPAREN) {
		return identifiers
	}

	expr := p.parseIdentifier(true)

	identifiers = append(identifiers, expr)

	for p.match(token.COMMA) {
		expr := p.parseIdentifier(true)
		identifiers = append(identifiers, expr)
	}

	p.expect(token.RPAREN)

	return identifiers
}

func (p *Parser) parseIdentifier(expectAnnotation bool) *ast.IdentifierExpression {

	tok := p.expect(token.IDENTIFIER)
	var annotation ast.TypeExpression

	if p.match(token.COLON) {
		annotation = p.parseTypeExpression()
	}

	if expectAnnotation && annotation == nil {
		panic(p.error("expected type annotation"))
	}

	return &ast.IdentifierExpression{
		Value:         tok.Lit,
		Pos:           tok.Pos,
		AnnotatedType: annotation,
	}
}

func (p *Parser) parseArrayLit() *ast.ArrayLiteral {

	elements, start, end := p.parseExpressionList(token.LBRACKET, token.RBRACKET)

	return &ast.ArrayLiteral{
		Elements:    elements,
		LBracketPos: start,
		RBracketPos: end,
	}
}

func (p *Parser) parseMapLiteral() *ast.MapLiteral {
	lit := &ast.MapLiteral{
		Pairs: make(map[ast.Expression]ast.Expression),
	}
	start := p.expect(token.LBRACE)
	lit.LBracePos = start.Pos
	// closes immediately
	if p.currentMatches(token.RBRACE) {
		end := p.expect(token.RBRACE)
		lit.RBracePos = end.Pos
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
	end := p.expect(token.RBRACE)
	lit.RBracePos = end.Pos
	return lit
}
