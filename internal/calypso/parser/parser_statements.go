package parser

import (
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

	}

	panic(p.error("expected statement"))
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
	ident := p.parseIdentifier(false)

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
	case *ast.AssignmentExpression, *ast.CallExpression:
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

	ident := p.parseIdentifier(false)

	// Has Generic Parameters
	var params *ast.GenericParametersClause
	if p.currentMatches(token.LSS) {
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

	ident := p.parseIdentifier(false)

	lBrace := p.expect(token.LBRACE)

	properties := []*ast.IdentifierExpression{}

	for p.current() != token.RBRACE {
		properties = append(properties, p.parseIdentifier(true))
		p.expect(token.SEMICOLON)
	}

	rBrace := p.expect(token.RBRACE)

	p.expect(token.SEMICOLON)

	return &ast.StructStatement{
		KeyWPos:    keyw.Pos,
		Identifier: ident,
		LBracePos:  lBrace.Pos,
		RBracePos:  rBrace.Pos,
		Properties: properties,
	}
}
