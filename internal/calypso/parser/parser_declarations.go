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
		stmt.IsGlobal = true
		return &ast.ConstantDeclaration{
			Stmt: stmt,
		}

	case token.FUNC:
		return p.parseFunctionDeclaration()
	case token.STANDARD:
		return p.parseStandardDeclaration()
	case token.EXTENSION:
		return p.parseExtensionDeclaration()
	case token.CONFORM:
		return p.parseConformanceDeclaration()
	case token.EXTERN:
		return p.parseExternDeclaration()
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
	case token.ALIAS, token.STRUCT, token.ENUM, token.TYPE:
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
	ident := p.parseIdentifierWithoutAnnotation()
	block := p.parseBlockStatement()

	return &ast.StandardDeclaration{
		KeyWPos:    keyw.Pos,
		Identifier: ident,
		Block:      block,
	}
}

func (p *Parser) parseExtensionDeclaration() *ast.ExtensionDeclaration {

	kw := p.expect(token.EXTENSION)

	ident := p.parseIdentifierWithoutAnnotation()

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

func (p *Parser) parseConformanceDeclaration() *ast.ConformanceDeclaration {

	kw := p.expect(token.CONFORM)

	target := p.parseIdentifierWithoutAnnotation()

	p.expect(token.TO)

	standard := p.parseIdentifierWithoutAnnotation()

	lBrace := p.expect(token.LBRACE)

	// Parse Functions in Extension

	content := []*ast.FunctionStatement{}
	types := []*ast.TypeStatement{}

	for p.current() != token.RBRACE {

		if p.currentMatches(token.TYPE) {
			t := p.parseTypeStatement()
			types = append(types, t)
		} else {
			f := &ast.FunctionStatement{
				Func: p.parseFunctionExpression(true),
			}
			content = append(content, f)
		}

	}

	rBrace := p.expect(token.RBRACE)

	return &ast.ConformanceDeclaration{
		KeyWPos:    kw.Pos,
		Standard:   standard,
		Target:     target,
		LBracePos:  lBrace.Pos,
		Signatures: content,
		Types:      types,
		RBracePos:  rBrace.Pos,
	}

}

func (p *Parser) parseExternDeclaration() *ast.ExternDeclaration {

	kw := p.expect(token.EXTERN)

	lit, ok := p.parseExpression().(*ast.StringLiteral)

	if !ok {
		panic(p.error("expected string literal in extern path"))
	}

	lBrace := p.expect(token.LBRACE)

	// Parse Functions in Extension

	content := []*ast.FunctionStatement{}

	for p.current() != token.RBRACE {

		f := &ast.FunctionStatement{
			Func: p.parseFunctionExpression(false),
		}
		content = append(content, f)
	}

	rBrace := p.expect(token.RBRACE)

	return &ast.ExternDeclaration{
		KeyWPos:    kw.Pos,
		LBracePos:  lBrace.Pos,
		RBracePos:  rBrace.Pos,
		Signatures: content,
		Target:     lit,
	}

}
