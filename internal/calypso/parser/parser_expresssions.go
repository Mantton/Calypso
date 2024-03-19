package parser

import (
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
		pos := p.previousScannedToken().Pos
		value, err := p.parseAssignmentExpression()

		if err != nil {
			return nil, err
		}

		return &ast.AssignmentExpression{
			Target: expr,
			Value:  value,
			OpPos:  pos,
		}, nil
	}

	return expr, nil
}

func (p *Parser) parseEqualityExpression() (ast.Expression, error) {
	expr, err := p.parseComparisonExpression()

	if err != nil {
		return nil, err
	}

	for p.match(token.NEQ, token.EQL) {
		op := p.previousScannedToken()
		right, err := p.parseComparisonExpression()

		if err != nil {
			return nil, err
		}

		expr = &ast.BinaryExpression{
			Left:  expr,
			Op:    op.Tok,
			OpPos: op.Pos,
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

	for p.match(token.R_CHEVRON, token.GEQ, token.LEQ, token.L_CHEVRON) {
		op := p.previousScannedToken()
		right, err := p.parseTermExpression()

		if err != nil {
			return nil, err
		}

		expr = &ast.BinaryExpression{
			Left:  expr,
			Op:    op.Tok,
			OpPos: op.Pos,
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
		op := p.previousScannedToken()
		right, err := p.parseFactorExpression()

		if err != nil {
			return nil, err
		}

		expr = &ast.BinaryExpression{
			Left:  expr,
			Op:    op.Tok,
			OpPos: op.Pos,
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
		op := p.previousScannedToken()
		right, err := p.parseUnaryExpression()

		if err != nil {
			return nil, err
		}

		expr = &ast.BinaryExpression{
			Left:  expr,
			Op:    op.Tok,
			OpPos: op.Pos,
			Right: right,
		}
	}

	return expr, nil
}

func (p *Parser) parseUnaryExpression() (ast.Expression, error) {

	if p.match(token.NOT, token.SUB, token.MUL, token.AMP) {
		op := p.previousScannedToken()
		right, err := p.parseUnaryExpression()

		if err != nil {
			return nil, err
		}

		return &ast.UnaryExpression{
			Op:         op.Tok,
			OpPosition: op.Pos,
			Expr:       right,
		}, nil
	}
	return p.parseFunctionCallExpression()
}

func (p *Parser) parseFunctionCallExpression() (ast.Expression, error) {
	expr, err := p.parsePropertyExpression()

	if err != nil {
		return nil, err
	}

	if p.currentMatches(token.LPAREN) {

		list, err := p.parseExpressionList(token.LPAREN, token.RPAREN)
		if err != nil {
			return nil, err
		}

		return &ast.FunctionCallExpression{
			Target:    expr,
			Arguments: list.Expressions,
			LParenPos: list.LPos,
			RParenPos: list.RPos,
		}, nil
	}

	return expr, nil
}

func (p *Parser) parsePropertyExpression() (ast.Expression, error) {
	expr, err := p.parseIndexExpression()

	if err != nil {
		return nil, err
	}

	for p.match(token.PERIOD) {
		dotPos := p.previousScannedToken().Pos

		property, err := p.parseIndexExpression()

		if err != nil {
			return nil, err
		}

		expr = &ast.FieldAccessExpression{
			Target: expr,
			Field:  property,
			DotPos: dotPos,
		}
	}

	return expr, nil
}

func (p *Parser) parseIndexExpression() (ast.Expression, error) {
	expr, err := p.parseSpecializationExpression()

	if err != nil {
		return nil, err
	}

	for p.match(token.LBRACKET) {
		lbrackPos := p.previousScannedToken().Pos
		idx, err := p.parseExpression()

		if err != nil {
			return nil, err
		}

		rbrack, err := p.expect(token.RBRACKET)

		if err != nil {
			return nil, err
		}

		expr = &ast.IndexExpression{
			Target:      expr,
			Index:       idx,
			LBracketPos: lbrackPos,
			RBracketPos: rbrack.Pos,
		}
	}

	return expr, nil
}

func (p *Parser) parseSpecializationExpression() (ast.Expression, error) {
	expr, err := p.parsePrimaryExpression()
	if err != nil {
		return nil, err
	}

	if ident, ok := expr.(*ast.IdentifierExpression); ok && p.currentMatches(token.L_CHEVRON) {
		c, err := p.parseGenericArgumentsClause()

		if err != nil {
			fmt.Println("???", p.currentScannedToken().Lit)
			// Not a Foo<Bar> specialization but perhaps a Foo < Bar Comparison
			return nil, err
		}

		// check if is composite initializer
		if p.currentMatches(token.LBRACE) {
			body, err := p.parseCompositeLiteralBody()
			if err != nil {
				return nil, err
			}
			expr = &ast.CompositeLiteral{
				Identifier:    ident,
				TypeArguments: c,
				Body:          body,
			}

		} else {
			expr = &ast.GenericSpecializationExpression{
				Identifier: ident,
				Clause:     c,
			}
		}

	}

	return expr, nil
}

func (p *Parser) parsePrimaryExpression() (ast.Expression, error) {
	var expr ast.Expression
	switch p.current() {
	case token.FALSE:
		expr = &ast.BooleanLiteral{
			Value: false,
			Pos:   p.currentScannedToken().Pos,
		}
		p.next()
	case token.TRUE:
		expr = &ast.BooleanLiteral{
			Value: true,
			Pos:   p.currentScannedToken().Pos,
		}
		p.next()

	case token.NIL:
		expr = &ast.NilLiteral{
			Pos: p.currentScannedToken().Pos,
		}

		p.next()
	case token.VOID:
		expr = &ast.VoidLiteral{
			Pos: p.currentScannedToken().Pos,
		}
		p.next()

	case token.INTEGER:
		v, err := strconv.ParseInt(p.currentScannedToken().Lit, 0, 64)

		if err != nil {
			return nil, err
		}
		expr = &ast.IntegerLiteral{
			Value: v,
			Pos:   p.currentScannedToken().Pos,
		}
		p.next()

	case token.FLOAT:
		v, err := strconv.ParseFloat(p.currentScannedToken().Lit, 64)

		if err != nil {
			return nil, err
		}

		expr = &ast.FloatLiteral{
			Value: v,
			Pos:   p.currentScannedToken().Pos,
		}
		p.next()

	case token.STRING:
		expr = &ast.StringLiteral{
			Value: p.currentScannedToken().Lit,
			Pos:   p.currentScannedToken().Pos,
		}
		p.next()
	case token.CHAR:
		l := p.currentScannedToken().Lit
		n := len(l)
		code, _, _, err := strconv.UnquoteChar(l[1:n-1], '\'')

		if err != nil {
			return nil, err
		}

		expr = &ast.CharLiteral{
			Value: int64(code),
			Pos:   p.currentScannedToken().Pos,
		}

		p.next()

	case token.IDENTIFIER:

		// is identifier, but token is `{` or `<`
		if tok, ok := p.peakAheadScannedToken(); ok && (tok.Tok == token.LBRACE) {
			ident, err := p.parseIdentifierWithoutAnnotation()
			if err != nil {
				return nil, err
			}
			body, err := p.parseCompositeLiteralBody()
			if err != nil {
				return nil, err
			}
			expr = &ast.CompositeLiteral{
				Identifier: ident,
				Body:       body,
			}

		} else {
			return p.parseIdentifierWithoutAnnotation()
		}

	case token.LPAREN:
		start, err := p.expect(token.LPAREN)

		if err != nil {
			return nil, err
		}
		expr, err := p.parseExpression()

		if err != nil {
			return nil, err
		}

		end, err := p.expect(token.RPAREN)

		if err != nil {
			return nil, err
		}

		return &ast.GroupedExpression{
			LParenPos: start.Pos,
			Expr:      expr,
			RParenPos: end.Pos,
		}, nil

	case token.LBRACKET:
		return p.parseArrayLit()
	case token.LBRACE:
		return p.parseMapLiteral()
	}

	if expr != nil {
		return expr, nil
	}

	msg := fmt.Sprintf("expected expression, got `%s`", p.currentScannedToken().Lit)
	return nil, p.error(msg)
}

func (p *Parser) parseExpressionList(start, end token.Token) (*ast.ExpressionList, error) {
	list := []ast.Expression{}

	// expect start token
	s, err := p.expect(start)

	if err != nil {
		return nil, err
	}
	// []ast.Expression, token.TokenPosition, token.TokenPosition
	// if immediately followed by end token, return
	if p.currentMatches(end) {
		e, err := p.expect(end)
		if err != nil {
			return nil, err
		}
		return &ast.ExpressionList{
			Expressions: list,
			LPos:        s.Pos,
			RPos:        e.Pos,
		}, nil
	}

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	list = append(list, expr)

	for p.match(token.COMMA) {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		list = append(list, expr)
	}

	e, err := p.expect(end)
	if err != nil {
		return nil, err
	}

	return &ast.ExpressionList{
		Expressions: list,
		LPos:        s.Pos,
		RPos:        e.Pos,
	}, nil
}

func (p *Parser) parseFunctionExpression(requiresBody bool) (*ast.FunctionExpression, error) {
	start, err := p.expect(token.FUNC) // Expect current to be `func`, consume

	if err != nil {
		return nil, err
	}

	// Name
	ident, err := p.parseIdentifierWithoutAnnotation()

	if err != nil {
		return nil, err
	}

	var genParams *ast.GenericParametersClause

	if p.currentMatches(token.L_CHEVRON) {
		genParams, err = p.parseGenericParameterClause()
		if err != nil {
			return nil, err
		}
	}

	// Parameters
	params, err := p.parseFunctionParameters()

	if err != nil {
		return nil, err
	}

	if len(params) > 99 {
		return nil, p.error("too many parameters, maximum of 99")
	}
	// Return Type
	retType, err := p.parseFunctionReturnType()

	if err != nil {
		return nil, err
	}

	// Body

	var body *ast.BlockStatement

	if p.currentMatches(token.LBRACE) {
		body, err = p.parseFunctionBody()

		if err != nil {
			return nil, err
		}

	} else if requiresBody {
		return nil, p.error("expected function body")
	} else {
		_, err := p.expect(token.SEMICOLON)

		if err != nil {
			return nil, err
		}
	}

	return &ast.FunctionExpression{
		KeyWPos:       start.Pos,
		Identifier:    ident,
		Body:          body,
		Params:        params,
		ReturnType:    retType,
		GenericParams: genParams,
	}, nil
}

func (p *Parser) parseFunctionBody() (*ast.BlockStatement, error) {
	// Opening
	_, err := p.expect(token.LBRACE)

	if err != nil {
		return nil, err
	}
	statements, err := p.parseStatementList()

	if err != nil {
		return nil, err
	}

	// Closing
	_, err = p.expect(token.RBRACE)

	if err != nil {
		return nil, err
	}

	return &ast.BlockStatement{
		Statements: statements,
	}, nil

}

func (p *Parser) parseFunctionReturnType() (ast.TypeExpression, error) {
	if p.match(token.R_ARROW) {
		annotatedType, err := p.parseTypeExpression()
		if err != nil {
			return nil, err
		}
		return annotatedType, nil
	} else {
		return nil, nil
	}
}

func (p *Parser) parseFunctionParameters() ([]*ast.IdentifierExpression, error) {
	identifiers := []*ast.IdentifierExpression{}

	p.expect(token.LPAREN)

	// if immediately followed by end token, return
	if p.match(token.RPAREN) {
		return identifiers, nil
	}

	expr, err := p.parseIdentifierWithRequiredAnnotation()

	if err != nil {
		return nil, err
	}

	identifiers = append(identifiers, expr)

	for p.match(token.COMMA) {
		expr, err := p.parseIdentifierWithRequiredAnnotation()

		if err != nil {
			return nil, err
		}

		identifiers = append(identifiers, expr)
	}

	_, err = p.expect(token.RPAREN)

	if err != nil {
		return nil, err
	}

	return identifiers, nil
}

func (p *Parser) parseIdentifierWithOptionalAnnotation() (*ast.IdentifierExpression, error) {
	tok, err := p.expect(token.IDENTIFIER)

	if err != nil {
		return nil, err
	}

	var annotation ast.TypeExpression

	if p.match(token.COLON) {
		annotation, err = p.parseTypeExpression()
		if err != nil {
			return nil, err
		}
	}

	return &ast.IdentifierExpression{
		Value:         tok.Lit,
		Pos:           tok.Pos,
		AnnotatedType: annotation,
	}, nil

}
func (p *Parser) parseIdentifierWithRequiredAnnotation() (*ast.IdentifierExpression, error) {
	tok, err := p.expect(token.IDENTIFIER)

	if err != nil {
		return nil, err
	}

	var annotation ast.TypeExpression

	_, err = p.expect(token.COLON)

	if err != nil {
		return nil, err
	}

	annotation, err = p.parseTypeExpression()
	if err != nil {
		return nil, err
	}

	return &ast.IdentifierExpression{
		Value:         tok.Lit,
		Pos:           tok.Pos,
		AnnotatedType: annotation,
	}, nil

}
func (p *Parser) parseIdentifierWithoutAnnotation() (*ast.IdentifierExpression, error) {
	tok, err := p.expect(token.IDENTIFIER)

	if err != nil {
		return nil, err
	}

	return &ast.IdentifierExpression{
		Value: tok.Lit,
		Pos:   tok.Pos,
	}, nil
}

func (p *Parser) parseArrayLit() (*ast.ArrayLiteral, error) {

	list, err := p.parseExpressionList(token.LBRACKET, token.RBRACKET)
	if err != nil {
		return nil, err
	}

	return &ast.ArrayLiteral{
		Elements:    list.Expressions,
		LBracketPos: list.LPos,
		RBracketPos: list.RPos,
	}, nil
}

func (p *Parser) parseMapLiteral() (*ast.MapLiteral, error) {
	lit := &ast.MapLiteral{}
	start, err := p.expect(token.LBRACE)

	if err != nil {
		return nil, err
	}

	lit.LBracePos = start.Pos
	// closes immediately
	if p.currentMatches(token.RBRACE) {
		end, err := p.expect(token.RBRACE)

		if err != nil {
			return nil, err
		}
		lit.RBracePos = end.Pos
		return lit, nil
	}
	// Loop until match with RBRACE
	for !p.match(token.RBRACE) {

		// Parse Key
		key, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		// Parse Colon Divider
		colon, err := p.expect(token.COLON)
		if err != nil {
			return nil, err
		}

		// Parse Value

		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		expr := &ast.KeyValueExpression{
			Key:      key,
			Value:    value,
			ColonPos: colon.Pos,
		}

		lit.Pairs = append(lit.Pairs, expr)

		if p.currentMatches(token.RBRACE) {
			break
		} else {
			_, err := p.expect(token.COMMA)

			if err != nil {
				return nil, err
			}
		}

	}

	// expect closing brace
	end, err := p.expect(token.RBRACE)

	if err != nil {
		return nil, err
	}

	lit.RBracePos = end.Pos
	return lit, nil
}

func (p *Parser) parseCompositeLiteralBody() (*ast.CompositeLiteralBody, error) {

	lBrace, err := p.expect(token.LBRACE)

	if err != nil {
		return nil, err
	}

	pairs := []*ast.CompositeLiteralField{}
	// Loop until match with RBRACE
	for !p.match(token.RBRACE) {

		// Parse Key
		key, err := p.parseIdentifierWithoutAnnotation()

		if err != nil {
			return nil, err
		}

		// Parse Colon Divider
		colon, err := p.expect(token.COLON)

		if err != nil {
			return nil, err
		}

		// Parse Value

		value, err := p.parseExpression()

		if err != nil {
			return nil, err
		}

		expr := &ast.CompositeLiteralField{
			Key:      key,
			Value:    value,
			ColonPos: colon.Pos,
		}

		pairs = append(pairs, expr)

		if p.currentMatches(token.RBRACE) {
			break
		} else {
			_, err = p.expect(token.COMMA)

			if err != nil {
				return nil, err
			}
		}

	}

	// parse kv expression
	rBrace, err := p.expect(token.RBRACE)

	if err != nil {
		return nil, err
	}

	return &ast.CompositeLiteralBody{
		LBracePos: lBrace.Pos,
		Fields:    pairs,
		RBracePos: rBrace.Pos,
	}, nil
}
