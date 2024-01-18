package parser

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

type Parser struct {
	tokens []token.ScannedToken
	errors []string

	cursor int
}

func New(tokens []token.ScannedToken) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) Parse() *ast.File {

	moduleName := "main"
	// - Parse Module Declaration
	// module main
	p.expect(token.MODULE)     // Consume module token
	p.expect(token.IDENTIFIER) // TODO: Parse Identifier
	p.expect(token.SEMICOLON)  // Consume semicolon

	// Imports

	// Declarations
	var declarations []ast.Declaration

	for p.current() != token.EOF {
		declarations = append(declarations, p.parseDeclaration())
	}

	return &ast.File{
		ModuleName:   moduleName,
		Declarations: declarations,
		Errors:       p.errors,
	}

}

//

func (p *Parser) current() token.Token {
	return p.tokens[p.cursor].Tok
}

func (p *Parser) currentScannedToken() token.ScannedToken {
	return p.tokens[p.cursor]
}

func (p *Parser) peekAhead() token.Token {

	idx := p.cursor + 1

	if idx >= len(p.tokens) {
		return token.EOF
	}

	return p.tokens[idx].Tok
}

// bool indicating the current token is of the specified type
func (p *Parser) currentMatches(t token.Token) bool {
	return p.current() == t
}

// consumes a token if the peek matches a specified token, returns true if the peek matches setting the current token to it
func (p *Parser) consumeIfPeekMatches(t token.Token) bool {
	if p.peekMatches(t) {
		p.next()
		return true
	} else {
		return false
	}
}

func (p *Parser) expect(t token.Token) {
	if p.current() != t {
		fmt.Println(p.currentScannedToken())
		panic("expected something")
	} else {
		p.next()
	}
}
func (p *Parser) next() {
	p.cursor++
}

func (p *Parser) advance(check token.NodeChecker) {
	for p.current() != token.EOF {
		if check(p.current()) {
			break
		} else {
			p.next()
		}
	}
}

// bool indicating the peek/next token is of the specified type
func (p *Parser) peekMatches(t token.Token) bool {
	return p.peekAhead() == t
}

func (p *Parser) parseDeclaration() ast.Declaration {

	switch p.current() {
	case token.CONST:
		panic("parse constant")
	case token.FUNC:
		panic("parse function")
	}

	panic("expected declaration")
}
