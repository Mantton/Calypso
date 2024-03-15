package parser

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (p *Parser) parseStatement() ast.Statement {
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
	case token.ALIAS:
		return p.parseAliasStatement()
	case token.FUNC:
		return &ast.FunctionStatement{
			Func: p.parseFunctionExpression(false),
		}
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

	panic(p.error(fmt.Sprintf("expected statement, got %s", p.currentScannedToken().Lit)))
}

func (p *Parser) parseVariableStatement() *ast.VariableStatement {
	/**
	let x = `expr`;
	const y = `expr`;
	const z :int = `expr`;
	*/
	isConst := p.current() == token.CONST
	start := p.currentScannedToken().Pos
	p.next() // Move to next token
	ident := p.parseIdentifierWithOptionalAnnotation()

	// Parse Type Expression If Found

	p.expect(token.ASSIGN)

	expr := p.parseExpression()

	p.expect(token.SEMICOLON)

	return &ast.VariableStatement{
		KeyWPos:    start,
		Identifier: ident,
		Value:      expr,
		IsConstant: isConst,
	}

}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	/**
	   {
	  	let x = 10;
	  	print("hello");
	  }
	*/

	// Opening
	start := p.expect(token.LBRACE)
	statements := p.parseStatementList()
	// Closing
	end := p.expect(token.RBRACE)

	return &ast.BlockStatement{
		LBrackPos:  start.Pos,
		Statements: statements,
		RBrackPos:  end.Pos,
	}

}

func (p *Parser) parseIfStatement() ast.Statement {
	/**
	  if (true) {
		return false;
	  } else {
		return true;
	  }
	*/

	start := p.expect(token.IF)

	// Condition
	p.expect(token.LPAREN)

	stmt := &ast.IfStatement{
		KeyWPos: start.Pos,
	}
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

	return stmt

}

func (p *Parser) parseReturnStatement() ast.Statement {
	start := p.expect(token.RETURN)

	var expr ast.Expression
	if p.currentMatches(token.SEMICOLON) {
		expr = &ast.VoidLiteral{
			Pos: p.currentScannedToken().Pos,
		}
	} else {
		expr = p.parseExpression()
	}

	p.expect(token.SEMICOLON)

	return &ast.ReturnStatement{
		Value:   expr,
		KeyWPos: start.Pos,
	}

}

func (p *Parser) parseWhileStatement() ast.Statement {
	start := p.expect(token.WHILE)
	// Condition
	p.expect(token.LPAREN)

	stmt := &ast.WhileStatement{
		KeyWPos: start.Pos,
	}
	condition := p.parseExpression()

	stmt.Condition = condition

	p.expect(token.RPAREN)

	// Action Block
	block := p.parseBlockStatement()
	stmt.Action = block

	return stmt
}

func (p *Parser) parseExpressionStatement() ast.Statement {

	expr := p.parseExpression()

	switch expr := expr.(type) {
	case *ast.AssignmentExpression, *ast.FunctionCallExpression:
		p.expect(token.SEMICOLON)
		return &ast.ExpressionStatement{
			Expr: expr,
		}
	default:
		panic(p.error("expected statement, not expression"))
	}

}

func (p *Parser) parseAliasStatement() *ast.AliasStatement {

	// Consume Keyword
	kwPos := p.expect(token.ALIAS).Pos

	// Consume TypeExpression

	ident := p.parseIdentifierWithOptionalAnnotation()

	// Has Generic Parameters
	var params *ast.GenericParametersClause
	if p.currentMatches(token.L_CHEVRON) {
		params = p.parseGenericParameterClause()
	}

	// Assign
	eqPos := p.expect(token.ASSIGN).Pos

	target := p.parseTypeExpression()
	p.expect(token.SEMICOLON)

	return &ast.AliasStatement{
		KeyWPos:       kwPos,
		EqPos:         eqPos,
		Identifier:    ident,
		Target:        target,
		GenericParams: params,
	}

}

func (p *Parser) parseStructStatement() *ast.StructStatement {

	keyw := p.expect(token.STRUCT)

	ident := p.parseIdentifierWithoutAnnotation()

	var genericParams *ast.GenericParametersClause

	if p.currentMatches(token.L_CHEVRON) {
		genericParams = p.parseGenericParameterClause()
	}

	lBrace := p.expect(token.LBRACE)

	properties := []*ast.IdentifierExpression{}

	for p.current() != token.RBRACE {
		properties = append(properties, p.parseIdentifierWithRequiredAnnotation())
		p.expect(token.SEMICOLON)
	}

	rBrace := p.expect(token.RBRACE)

	return &ast.StructStatement{
		KeyWPos:       keyw.Pos,
		Identifier:    ident,
		GenericParams: genericParams,
		LBracePos:     lBrace.Pos,
		RBracePos:     rBrace.Pos,
		Fields:        properties,
	}
}

func (p *Parser) parseEnumStatement() *ast.EnumStatement {
	kwPos := p.expect(token.ENUM).Pos

	ident := p.parseIdentifierWithoutAnnotation()

	var genericParams *ast.GenericParametersClause

	if p.currentMatches(token.L_CHEVRON) {
		genericParams = p.parseGenericParameterClause()
	}

	lbracePos := p.expect(token.LBRACE).Pos

	stmt := &ast.EnumStatement{
		KeyWPos:       kwPos,
		Identifier:    ident,
		GenericParams: genericParams,
		LBracePos:     lbracePos,
	}

	for p.current() != token.RBRACE {

		ident := p.parseIdentifierWithoutAnnotation()
		var discriminator *ast.EnumDiscriminantExpression
		var fields *ast.FieldListExpression

		if p.match(token.ASSIGN) {
			// Discriminator

			discriminator = &ast.EnumDiscriminantExpression{
				Value: p.parseExpression(),
			}
		} else if p.currentMatches(token.LPAREN) {
			// Tuple
			f := p.parseFieldList()
			fields = &ast.FieldListExpression{
				Fields: f,
			}
		}
		p.expect(token.COMMA)

		variant := &ast.EnumVariantExpression{
			Identifier:   ident,
			Discriminant: discriminator,
			Fields:       fields,
		}

		stmt.Variants = append(stmt.Variants, variant)
	}

	rBrace := p.expect(token.RBRACE)

	stmt.RBracePos = rBrace.Pos
	return stmt
}

func (p *Parser) parseFieldList() []ast.TypeExpression {
	params := []ast.TypeExpression{}

	p.expect(token.LPAREN)
	if p.match(token.RPAREN) {
		return params
	}

	a := p.parseTypeExpression()
	params = append(params, a)
	for p.match(token.COMMA) {
		expr := p.parseTypeExpression()
		params = append(params, expr)
	}

	p.expect(token.RPAREN)

	return params
}

func (p *Parser) parseSwitchStatement() *ast.SwitchStatement {
	prevState := p.inSwitch
	defer func() {
		p.inSwitch = prevState
	}()

	p.inSwitch = true
	defer func() {
		p.inSwitch = false
	}()

	// 1 - Keyword
	kwPos := p.expect(token.SWITCH).Pos

	// 2 - Condition
	cond := p.parseExpression()

	// 3 - L Brace

	lBracePos := p.expect(token.LBRACE).Pos

	//  4 - Cases
	cases := p.parseSwitchCases()

	// 5 - R Brace
	rBracePos := p.expect(token.RBRACE).Pos

	return &ast.SwitchStatement{
		KeyWPos:   kwPos,
		LBracePos: lBracePos,
		RBracePos: rBracePos,
		Condition: cond,
		Cases:     cases,
	}

}

func (p *Parser) parseSwitchCases() []*ast.SwitchCaseExpression {

	cases := []*ast.SwitchCaseExpression{}

	for p.currentMatches(token.CASE) || p.currentMatches(token.DEFAULT) {
		cases = append(cases, p.parseSwitchCase())
	}

	return cases
}

func (p *Parser) parseSwitchCase() *ast.SwitchCaseExpression {

	if p.currentMatches(token.CASE) {

		// 1 - Case Keyword
		kwPos := p.expect(token.CASE).Pos

		// 2 - Condition
		cond := p.parseExpression()

		// 3 - Body
		var body *ast.BlockStatement
		var colonPos token.TokenPosition
		if p.currentMatches(token.LBRACE) {
			body = p.parseBlockStatement()
			colonPos = body.LBrackPos
		} else {
			colonPos = p.expect(token.COLON).Pos
			stmts := []ast.Statement{}
			for !p.currentMatches(token.CASE) &&
				!p.currentMatches(token.RBRACE) &&
				!p.currentMatches(token.DEFAULT) {
				s := p.parseStatement()
				stmts = append(stmts, s)
			}

			body = &ast.BlockStatement{
				LBrackPos:  colonPos,
				Statements: stmts,
				RBrackPos:  stmts[len(stmts)-1].Range().End,
			}
		}

		return &ast.SwitchCaseExpression{
			KeyWPos:   kwPos,
			Condition: cond,
			ColonPos:  colonPos,
			Action:    body,
		}

	}

	// Default Case

	// 1 - Keyword
	kwPos := p.expect(token.DEFAULT).Pos

	// 2 - Body
	var body *ast.BlockStatement
	var colonPos token.TokenPosition
	if p.currentMatches(token.LBRACE) {
		body = p.parseBlockStatement()
		colonPos = body.LBrackPos
	} else {
		colonPos = p.expect(token.COLON).Pos
		stmts := []ast.Statement{}
		for !p.currentMatches(token.CASE) &&
			!p.currentMatches(token.RBRACE) &&
			!p.currentMatches(token.DEFAULT) {
			s := p.parseStatement()
			stmts = append(stmts, s)
		}

		body = &ast.BlockStatement{
			LBrackPos:  colonPos,
			Statements: stmts,
			RBrackPos:  stmts[len(stmts)-1].Range().End,
		}
	}

	return &ast.SwitchCaseExpression{
		KeyWPos:   kwPos,
		ColonPos:  colonPos,
		Action:    body,
		IsDefault: true,
		Condition: nil,
	}

}

func (p *Parser) parseBreakStatement() *ast.BreakStatement {

	if !p.inSwitch {
		panic(p.error("cannot break outside switch statement"))
	}

	kwPos := p.expect(token.BREAK).Pos
	p.expect(token.SEMICOLON)
	return &ast.BreakStatement{
		KeyWPos: kwPos,
	}
}

func (p *Parser) parseTypeStatement() *ast.TypeStatement {
	// Consume Keyword
	kwPos := p.expect(token.TYPE).Pos

	// Consume TypeExpression

	ident := p.parseIdentifierWithoutAnnotation()

	// Has Generic Parameters
	var params *ast.GenericParametersClause
	if p.currentMatches(token.L_CHEVRON) {
		params = p.parseGenericParameterClause()
	}

	// Assign
	var eqPos token.TokenPosition
	var value ast.TypeExpression
	if p.currentMatches(token.ASSIGN) {
		eqPos = p.expect(token.ASSIGN).Pos
		value = p.parseTypeExpression()
	}

	p.expect(token.SEMICOLON)

	return &ast.TypeStatement{
		KeyWPos:       kwPos,
		EqPos:         eqPos,
		GenericParams: params,
		Value:         value,
		Identifier:    ident,
	}
}
