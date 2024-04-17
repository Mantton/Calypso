package parser

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (p *Parser) parseDeclaration() (ast.Declaration, error) {

	// Modifiers
	for token.IsModifier(p.current()) {
		err := p.handleModifier(p.current())

		// Notify error, but do not skip parsing
		if err != nil {
			p.errors.Add(err)
		}
		p.next()
	}

	if !token.IsModifiable(p.current()) && len(p.modifiers) != 0 {
		p.modifiers = nil
		return nil, p.error("declaration is not modifiable")
	}

	switch p.current() {
	case token.CONST:
		stmt, err := p.parseVariableStatement()

		if err != nil {
			fmt.Println("err", err)

			return nil, err
		}

		stmt.IsGlobal = true
		return &ast.ConstantDeclaration{
			Stmt: stmt,
		}, nil

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
	case token.IMPORT:
		return p.parseImportDeclaration()
	default:
		return p.parseStatementDeclaration()
	}
}

func (p *Parser) parseFunctionDeclaration() (*ast.FunctionDeclaration, error) {

	fn, err := p.parseFunctionExpression(true)
	if err != nil {
		return nil, err
	}

	if fn.Body == nil {
		return nil, p.error("expected body in function declaration")
	}

	return &ast.FunctionDeclaration{
		Func: fn,
	}, nil
}

func (p *Parser) parseStatementDeclaration() (*ast.StatementDeclaration, error) {

	switch p.current() {
	case token.STRUCT, token.ENUM, token.TYPE:
		stmt, err := p.parseStatement()

		if err != nil {
			return nil, err
		}

		return &ast.StatementDeclaration{
			Stmt: stmt,
		}, nil

	default:
		msg := fmt.Sprintf("expected declaration, `%s` does not start a declaration", p.currentScannedToken().Lit)
		return nil, p.error(msg)
	}
}

func (p *Parser) parseStatementList() ([]ast.Statement, error) {

	var list = []ast.Statement{}
	hasError := false
	for p.current() != token.RBRACE && p.current() != token.EOF {
		statement, err := p.parseStatement()

		if err != nil {
			// return nil, err
			p.handleError(err, STMT)
			hasError = true
		}
		list = append(list, statement)
	}

	if hasError {
		return nil, fmt.Errorf("error in statement list")
	}

	return list, nil
}

func (p *Parser) parseStandardDeclaration() (*ast.StandardDeclaration, error) {
	// Visibility Modifiers
	vis, err := p.resolveNonFuncMods()
	if err != nil {
		p.errors.Add(err)
	}
	keyw, err := p.expect(token.STANDARD)

	if err != nil {
		return nil, err
	}

	ident, err := p.parseIdentifierWithoutAnnotation()
	if err != nil {
		return nil, err
	}

	block, err := p.parseBlockStatement()

	if err != nil {
		return nil, err
	}

	return &ast.StandardDeclaration{
		KeyWPos:    keyw.Pos,
		Identifier: ident,
		Block:      block,
		Visibility: vis,
	}, nil
}

func (p *Parser) parseExtensionDeclaration() (*ast.ExtensionDeclaration, error) {

	kw, err := p.expect(token.EXTENSION)

	if err != nil {
		return nil, err
	}

	ident, err := p.parseIdentifierWithoutAnnotation()
	if err != nil {
		return nil, err
	}

	lBrace, err := p.expect(token.LBRACE)

	if err != nil {
		return nil, err
	}

	// Parse Functions in Extension

	content := []*ast.FunctionStatement{}

	for p.current() != token.RBRACE {
		// Modifiers
		for token.IsModifier(p.current()) {
			err := p.handleModifier(p.current())

			// Notify error, but do not skip parsing
			if err != nil {
				p.errors.Add(err)
			}
			p.next()
		}

		fn, err := p.parseFunctionExpression(true)
		if err != nil {
			return nil, err
		}
		f := &ast.FunctionStatement{
			Func: fn,
		}
		content = append(content, f)
	}

	rBrace, err := p.expect(token.RBRACE)

	if err != nil {
		return nil, err
	}

	return &ast.ExtensionDeclaration{
		KeyWPos:    kw.Pos,
		Identifier: ident,
		LBracePos:  lBrace.Pos,
		Content:    content,
		RBracePos:  rBrace.Pos,
	}, nil

}

func (p *Parser) parseConformanceDeclaration() (*ast.ConformanceDeclaration, error) {

	kw, err := p.expect(token.CONFORM)

	if err != nil {
		return nil, err
	}

	target, err := p.parseIdentifierWithoutAnnotation()
	if err != nil {
		return nil, err
	}

	_, err = p.expect(token.TO)
	if err != nil {
		return nil, err
	}

	standard, err := p.parseIdentifierWithoutAnnotation()
	if err != nil {
		return nil, err
	}

	lBrace, err := p.expect(token.LBRACE)

	if err != nil {
		return nil, err
	}

	// Parse Functions in Extension

	content := []*ast.FunctionStatement{}
	types := []*ast.TypeStatement{}

	for p.current() != token.RBRACE {

		if p.currentMatches(token.TYPE) {
			t, err := p.parseTypeStatement()

			if err != nil {
				return nil, err
			}

			types = append(types, t)
		} else {
			// Modifiers
			for token.IsModifier(p.current()) {
				err := p.handleModifier(p.current())

				// Notify error, but do not skip parsing
				if err != nil {
					p.errors.Add(err)
				}
				p.next()
			}
			fn, err := p.parseFunctionExpression(true)
			if err != nil {
				return nil, err
			}
			f := &ast.FunctionStatement{
				Func: fn,
			}
			content = append(content, f)
		}

	}

	rBrace, err := p.expect(token.RBRACE)

	if err != nil {
		return nil, err
	}

	return &ast.ConformanceDeclaration{
		KeyWPos:    kw.Pos,
		Standard:   standard,
		Target:     target,
		LBracePos:  lBrace.Pos,
		Signatures: content,
		Types:      types,
		RBracePos:  rBrace.Pos,
	}, nil

}

func (p *Parser) parseExternDeclaration() (*ast.ExternDeclaration, error) {

	kw, err := p.expect(token.EXTERN)

	if err != nil {
		return nil, err
	}

	ex, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	lit, ok := ex.(*ast.StringLiteral)

	if !ok {
		return nil, p.error("expected string literal in extern path")
	}

	lBrace, err := p.expect(token.LBRACE)

	if err != nil {
		return nil, err
	}

	// Parse Functions in Extension

	content := []*ast.FunctionStatement{}

	for p.current() != token.RBRACE {
		fn, err := p.parseFunctionExpression(false)
		if err != nil {
			return nil, err
		}
		f := &ast.FunctionStatement{
			Func: fn,
		}
		content = append(content, f)
	}

	rBrace, err := p.expect(token.RBRACE)

	if err != nil {
		return nil, err
	}

	return &ast.ExternDeclaration{
		KeyWPos:    kw.Pos,
		LBracePos:  lBrace.Pos,
		RBracePos:  rBrace.Pos,
		Signatures: content,
		Target:     lit,
	}, nil

}

func (p *Parser) parseImportDeclaration() (*ast.ImportDeclaration, error) {

	decl := &ast.ImportDeclaration{}

	// KW
	kw, err := p.expect(token.IMPORT)

	if err != nil {
		return nil, err
	}

	decl.KeyWPos = kw.Pos

	// Path
	pathTok, err := p.expect(token.STRING)

	if err != nil {
		return nil, err
	}

	path := &ast.StringLiteral{
		Value: pathTok.Lit,
		Pos:   pathTok.Pos,
	}

	decl.Path = path

	if !p.currentMatches(token.AS) {
		_, err := p.expect(token.SEMICOLON)
		if err != nil {
			return nil, err
		}

		return decl, nil
	}

	// as Keyword
	asKW, err := p.expect(token.AS)
	if err != nil {
		return nil, err
	}

	decl.AsKeywPos = asKW.Pos

	// Alias
	alias, err := p.parseIdentifierWithoutAnnotation()

	if err != nil {
		return nil, err
	}

	decl.Alias = alias

	// Semicolon
	_, err = p.expect(token.SEMICOLON)
	if err != nil {
		return nil, err
	}

	return decl, nil
}
