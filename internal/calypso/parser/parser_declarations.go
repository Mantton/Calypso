package parser

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (p *Parser) parseDeclaration() ast.Declaration {
	switch p.current() {
	case token.CONST:
		stmt := p.parseVariableStatement()

		return &ast.ConstantDeclaration{
			Stmt: stmt,
		}

	case token.FUNC:
		return p.parseFunctionDeclaration()
	case token.STANDARD:
		return p.parseStandardDeclaration()
	case token.TYPE:
		return p.parseTypeDeclaration()
	default:
		return p.parseStatementDeclaration()
	}
}

func (p *Parser) parseFunctionDeclaration() *ast.FunctionDeclaration {

	fn := p.parseFunctionExpression(true)

	if fn.Body == nil {
		panic(p.error("expected body in function declaration"))
	}

	return &ast.FunctionDeclaration{
		Func: fn,
	}
}

func (p *Parser) parseStatementDeclaration() *ast.StatementDeclaration {

	switch p.current() {
	case token.ALIAS:
		stmt := p.parseStatement()
		return &ast.StatementDeclaration{
			Stmt: stmt,
		}
	default:
		msg := fmt.Sprintf("expected declaration, `%s` does not start a declaration", p.currentScannedToken().Lit)
		panic(p.error(msg))
	}
}

func (p *Parser) parseStatementList() []ast.Statement {

	var list = []ast.Statement{}
	cancel := false
	for p.current() != token.RBRACE && p.current() != token.EOF && !cancel {

		func() {
			defer func() {
				if r := recover(); r != nil {
					if err, y := r.(lexer.Error); y {
						p.errors.Add(err)
					} else {
						panic(r)
					}
					hasMoved := p.advance(token.IsStatement)

					// avoid infinite loop
					if !hasMoved {
						p.next()
					}
				}
			}()

			statement := p.parseStatement()

			list = append(list, statement)
		}()

	}

	return list

}

func (p *Parser) parseStandardDeclaration() *ast.StandardDeclaration {

	keyw := p.expect(token.STANDARD)
	ident := p.parseIdentifier()

	/*
		TODO:
			Standards should behave similarly to rust traits,
			They should have constants, methods & types
	*/

	block := p.parseBlockStatement()

	return &ast.StandardDeclaration{
		KeyWPos:    keyw.Pos,
		Identifier: ident,
		Block:      block,
	}
}

func (p *Parser) parseTypeDeclaration() *ast.TypeDeclaration {
	// Consume Keyword
	kwPos := p.expect(token.TYPE).Pos

	// Consume TypeExpression

	ident := p.parseIdentifier()

	// Has Generic Parameters
	var params *ast.GenericParametersClause
	if p.currentMatches(token.LSS) {
		params = p.parseGenericParameterClause()
	}

	// Assign
	eqPos := p.expect(token.ASSIGN).Pos

	value := p.parseTypeExpression()

	p.expect(token.SEMICOLON)

	return &ast.TypeDeclaration{
		KeyWPos:       kwPos,
		EqPos:         eqPos,
		GenericParams: params,
		Value:         value,
		Identifier:    ident,
	}
}
