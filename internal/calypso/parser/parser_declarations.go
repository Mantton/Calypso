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
	case token.EXTENSION:
		return p.parseExtensionDeclaration()
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
	case token.ALIAS, token.STRUCT:
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
	ident := p.parseIdentifier(false)

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

	ident := p.parseIdentifier(false)

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

func (p *Parser) parseExtensionDeclaration() *ast.ExtensionDeclaration {

	kw := p.expect(token.EXTENSION)

	ident := p.parseIdentifier(false)

	lBrace := p.expect(token.LBRACE)

	// Parse Functions in Extension

	content := []*ast.FunctionStatement{}

	for p.current() != token.RBRACE {

		f := &ast.FunctionStatement{
			Func: p.parseFunctionExpression(true),
		}
		content = append(content, f)
	}

	rBrace := p.expect(token.RBRACE)

	return &ast.ExtensionDeclaration{
		KeyWPos:    kw.Pos,
		Identifier: ident,
		LBracePos:  lBrace.Pos,
		Content:    content,
		RBracePos:  rBrace.Pos,
	}

}
