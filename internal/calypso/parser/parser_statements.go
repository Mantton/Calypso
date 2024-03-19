package parser

import (
	"fmt"

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
	case token.FUNC:
		fn, err := p.parseFunctionExpression(false)
		if err != nil {
			return nil, err
		}
		return &ast.FunctionStatement{
			Func: fn,
		}, nil
	case token.STRUCT:
		return p.parseStructStatement()
	case token.ENUM:
		return p.parseEnumStatement()
	case token.SWITCH:
		return p.parseSwitchStatement()
	case token.BREAK:
		return p.parseBreakStatement()
	case token.TYPE:
		return p.parseTypeStatement()

	}

	return nil, p.error(fmt.Sprintf("expected statement, got %s", p.currentScannedToken().Lit))
}

func (p *Parser) parseVariableStatement() (*ast.VariableStatement, error) {
	/**
	let x = `expr`;
	const y = `expr`;
	const z :int = `expr`;
	*/
	isConst := p.current() == token.CONST
	start := p.currentScannedToken().Pos
	p.next() // Move to next token
	ident, err := p.parseIdentifierWithOptionalAnnotation()
	if err != nil {
		return nil, err
	}

	// Parse Type Expression If Found

	_, err = p.expect(token.ASSIGN)

	if err != nil {
		return nil, err
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	_, err = p.expect(token.SEMICOLON)

	if err != nil {
		return nil, err
	}

	return &ast.VariableStatement{
		KeyWPos:    start,
		Identifier: ident,
		Value:      expr,
		IsConstant: isConst,
	}, nil

}

func (p *Parser) parseBlockStatement() (*ast.BlockStatement, error) {
	/**
	   {
	  	let x = 10;
	  	print("hello");
	  }
	*/

	// Opening
	start, err := p.expect(token.LBRACE)

	if err != nil {
		return nil, err
	}
	statements, err := p.parseStatementList()

	if err != nil {
		return nil, err
	}
	// Closing
	end, err := p.expect(token.RBRACE)

	if err != nil {
		return nil, err
	}

	return &ast.BlockStatement{
		LBrackPos:  start.Pos,
		Statements: statements,
		RBrackPos:  end.Pos,
	}, nil

}

func (p *Parser) parseIfStatement() (ast.Statement, error) {
	/**
	  if true {
		return false;
	  } else {
		return true;
	  }
	*/

	start, err := p.expect(token.IF)

	if err != nil {
		return nil, err
	}
	stmt := &ast.IfStatement{
		KeyWPos: start.Pos,
	}
	// Condition

	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	stmt.Condition = condition

	// Action Block
	block, err := p.parseBlockStatement()

	if err != nil {
		return nil, err
	}

	stmt.Action = block

	// Conditional Block

	if p.currentMatches(token.ELSE) {
		p.next()
		alt, err := p.parseBlockStatement()

		if err != nil {
			return nil, err
		}

		stmt.Alternative = alt
	}

	return stmt, nil

}

func (p *Parser) parseReturnStatement() (ast.Statement, error) {
	start, err := p.expect(token.RETURN)

	if err != nil {
		return nil, err
	}

	var expr ast.Expression
	if p.currentMatches(token.SEMICOLON) {
		expr = &ast.VoidLiteral{
			Pos: p.currentScannedToken().Pos,
		}
	} else {
		expr, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.expect(token.SEMICOLON)

	if err != nil {
		return nil, err
	}

	return &ast.ReturnStatement{
		Value:   expr,
		KeyWPos: start.Pos,
	}, nil

}

func (p *Parser) parseWhileStatement() (ast.Statement, error) {
	start, err := p.expect(token.WHILE)

	if err != nil {
		return nil, err
	}
	// Condition
	_, err = p.expect(token.LPAREN)

	if err != nil {
		return nil, err
	}

	stmt := &ast.WhileStatement{
		KeyWPos: start.Pos,
	}
	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	stmt.Condition = condition

	_, err = p.expect(token.RPAREN)

	if err != nil {
		return nil, err
	}

	// Action Block
	block, err := p.parseBlockStatement()

	if err != nil {
		return nil, err
	}

	stmt.Action = block

	return stmt, nil
}

func (p *Parser) parseExpressionStatement() (ast.Statement, error) {

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	switch expr := expr.(type) {
	case *ast.AssignmentExpression, *ast.FunctionCallExpression:
		_, err := p.expect(token.SEMICOLON)

		if err != nil {
			return nil, err
		}

		return &ast.ExpressionStatement{
			Expr: expr,
		}, nil
	default:
		return nil, p.error("expected statement, not expression")
	}

}

func (p *Parser) parseStructStatement() (*ast.StructStatement, error) {

	keyw, err := p.expect(token.STRUCT)

	if err != nil {
		return nil, err
	}

	ident, err := p.parseIdentifierWithoutAnnotation()
	if err != nil {
		return nil, err
	}

	var genericParams *ast.GenericParametersClause

	if p.currentMatches(token.L_CHEVRON) {
		genericParams, err = p.parseGenericParameterClause()
		if err != nil {
			return nil, err
		}
	}

	lBrace, err := p.expect(token.LBRACE)

	if err != nil {
		return nil, err
	}

	properties := []*ast.IdentifierExpression{}

	for p.current() != token.RBRACE {
		v, err := p.parseIdentifierWithRequiredAnnotation()
		if err != nil {
			return nil, err
		}
		properties = append(properties, v)
		_, err = p.expect(token.SEMICOLON)
		if err != nil {
			return nil, err
		}
	}

	rBrace, err := p.expect(token.RBRACE)

	if err != nil {
		return nil, err
	}

	return &ast.StructStatement{
		KeyWPos:       keyw.Pos,
		Identifier:    ident,
		GenericParams: genericParams,
		LBracePos:     lBrace.Pos,
		RBracePos:     rBrace.Pos,
		Fields:        properties,
	}, nil
}

func (p *Parser) parseEnumStatement() (*ast.EnumStatement, error) {
	kw, err := p.expect(token.ENUM)

	if err != nil {
		return nil, err
	}

	ident, err := p.parseIdentifierWithoutAnnotation()
	if err != nil {
		return nil, err
	}

	var genericParams *ast.GenericParametersClause

	if p.currentMatches(token.L_CHEVRON) {
		genericParams, err = p.parseGenericParameterClause()
		if err != nil {
			return nil, err
		}
	}

	lbrace, err := p.expect(token.LBRACE)

	if err != nil {
		return nil, err
	}

	stmt := &ast.EnumStatement{
		KeyWPos:       kw.Pos,
		Identifier:    ident,
		GenericParams: genericParams,
		LBracePos:     lbrace.Pos,
	}

	for p.current() != token.RBRACE {

		ident, err := p.parseIdentifierWithoutAnnotation()
		if err != nil {
			return nil, err
		}
		var discriminator *ast.EnumDiscriminantExpression
		var fields *ast.FieldListExpression

		if p.match(token.ASSIGN) {
			// Discriminator
			v, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			discriminator = &ast.EnumDiscriminantExpression{
				Value: v,
			}
		} else if p.currentMatches(token.LPAREN) {
			// Tuple
			f, err := p.parseFieldList()

			if err != nil {
				return nil, err
			}

			fields = &ast.FieldListExpression{
				Fields: f,
			}
		}
		_, err = p.expect(token.COMMA)

		if err != nil {
			return nil, err
		}

		variant := &ast.EnumVariantExpression{
			Identifier:   ident,
			Discriminant: discriminator,
			Fields:       fields,
		}

		stmt.Variants = append(stmt.Variants, variant)
	}

	rBrace, err := p.expect(token.RBRACE)

	if err != nil {
		return nil, err
	}

	stmt.RBracePos = rBrace.Pos
	return stmt, nil
}

func (p *Parser) parseFieldList() ([]ast.TypeExpression, error) {
	params := []ast.TypeExpression{}

	_, err := p.expect(token.LPAREN)

	if err != nil {
		return nil, err
	}

	if p.match(token.RPAREN) {
		return params, nil
	}

	a, err := p.parseTypeExpression()
	if err != nil {
		return nil, err
	}
	params = append(params, a)

	for p.match(token.COMMA) {
		expr, err := p.parseTypeExpression()
		if err != nil {
			return nil, err
		}
		params = append(params, expr)
	}

	_, err = p.expect(token.RPAREN)

	if err != nil {
		return nil, err
	}

	return params, nil
}

func (p *Parser) parseSwitchStatement() (*ast.SwitchStatement, error) {
	prevState := p.inSwitch
	defer func() {
		p.inSwitch = prevState
	}()

	p.inSwitch = true
	defer func() {
		p.inSwitch = false
	}()

	// 1 - Keyword
	kw, err := p.expect(token.SWITCH)

	if err != nil {
		return nil, err
	}

	// 2 - Condition
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// 3 - L Brace

	lbrace, err := p.expect(token.LBRACE)

	if err != nil {
		return nil, err
	}

	//  4 - Cases
	cases, err := p.parseSwitchCases()

	if err != nil {
		return nil, err
	}

	// 5 - R Brace
	rbrace, err := p.expect(token.RBRACE)

	if err != nil {
		return nil, err
	}

	return &ast.SwitchStatement{
		KeyWPos:   kw.Pos,
		LBracePos: lbrace.Pos,
		RBracePos: rbrace.Pos,
		Condition: cond,
		Cases:     cases,
	}, nil

}

func (p *Parser) parseSwitchCases() ([]*ast.SwitchCaseExpression, error) {

	cases := []*ast.SwitchCaseExpression{}

	for p.currentMatches(token.CASE) || p.currentMatches(token.DEFAULT) {
		c, err := p.parseSwitchCase()

		if err != nil {
			return nil, err
		}
		cases = append(cases, c)
	}

	return cases, nil
}

func (p *Parser) parseSwitchCase() (*ast.SwitchCaseExpression, error) {

	if p.currentMatches(token.CASE) {

		// 1 - Case Keyword
		kw, err := p.expect(token.CASE)

		if err != nil {
			return nil, err
		}

		// 2 - Condition
		cond, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		// 3 - Body
		var body *ast.BlockStatement
		var colonPos token.TokenPosition
		if p.currentMatches(token.LBRACE) {
			body, err = p.parseBlockStatement()

			if err != nil {
				return nil, err
			}
			colonPos = body.LBrackPos
		} else {
			col, err := p.expect(token.COLON)

			if err != nil {
				return nil, err
			}

			colonPos = col.Pos
			stmts := []ast.Statement{}
			for !p.currentMatches(token.CASE) &&
				!p.currentMatches(token.RBRACE) &&
				!p.currentMatches(token.DEFAULT) {
				s, err := p.parseStatement()

				if err != nil {
					return nil, err
				}

				stmts = append(stmts, s)
			}

			body = &ast.BlockStatement{
				LBrackPos:  colonPos,
				Statements: stmts,
				RBrackPos:  stmts[len(stmts)-1].Range().End,
			}
		}

		return &ast.SwitchCaseExpression{
			KeyWPos:   kw.Pos,
			Condition: cond,
			ColonPos:  colonPos,
			Action:    body,
		}, nil

	}

	// Default Case

	// 1 - Keyword
	kw, err := p.expect(token.DEFAULT)

	if err != nil {
		return nil, err
	}

	// 2 - Body
	var body *ast.BlockStatement
	var colonPos token.TokenPosition
	if p.currentMatches(token.LBRACE) {
		body, err = p.parseBlockStatement()

		if err != nil {
			return nil, err
		}
		colonPos = body.LBrackPos
	} else {
		col, err := p.expect(token.COLON)

		if err != nil {
			return nil, err
		}

		colonPos = col.Pos
		stmts := []ast.Statement{}
		for !p.currentMatches(token.CASE) &&
			!p.currentMatches(token.RBRACE) &&
			!p.currentMatches(token.DEFAULT) {
			s, err := p.parseStatement()

			if err != nil {
				return nil, err
			}
			stmts = append(stmts, s)
		}

		body = &ast.BlockStatement{
			LBrackPos:  colonPos,
			Statements: stmts,
			RBrackPos:  stmts[len(stmts)-1].Range().End,
		}
	}

	return &ast.SwitchCaseExpression{
		KeyWPos:   kw.Pos,
		ColonPos:  colonPos,
		Action:    body,
		IsDefault: true,
		Condition: nil,
	}, nil

}

func (p *Parser) parseBreakStatement() (*ast.BreakStatement, error) {

	if !p.inSwitch {
		return nil, p.error("cannot break outside switch statement")
	}

	kw, err := p.expect(token.BREAK)

	if err != nil {
		return nil, err
	}

	_, err = p.expect(token.SEMICOLON)

	if err != nil {
		return nil, err
	}

	return &ast.BreakStatement{
		KeyWPos: kw.Pos,
	}, nil
}

func (p *Parser) parseTypeStatement() (*ast.TypeStatement, error) {
	// Consume Keyword
	kw, err := p.expect(token.TYPE)

	if err != nil {
		return nil, err
	}

	// Consume TypeExpression

	ident, err := p.parseIdentifierWithoutAnnotation()
	if err != nil {
		return nil, err
	}

	// Has Generic Parameters
	var params *ast.GenericParametersClause
	if p.currentMatches(token.L_CHEVRON) {
		params, err = p.parseGenericParameterClause()
		if err != nil {
			return nil, err
		}
	}

	// Assign
	var eqPos token.TokenPosition
	var value ast.TypeExpression
	if p.currentMatches(token.ASSIGN) {
		eq, err := p.expect(token.ASSIGN)

		if err != nil {
			return nil, err
		}

		eqPos = eq.Pos
		value, err = p.parseTypeExpression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.expect(token.SEMICOLON)
	if err != nil {
		return nil, err
	}

	return &ast.TypeStatement{
		KeyWPos:       kw.Pos,
		EqPos:         eqPos,
		GenericParams: params,
		Value:         value,
		Identifier:    ident,
	}, nil
}
