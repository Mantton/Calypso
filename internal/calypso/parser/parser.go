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

		defer func() {
			if r := recover(); r != nil {
				fmt.Println("DECL ERROR: ", r)
				p.advance(token.IsDeclaration)
			}
		}()
		declarations = append(declarations, p.parseDeclaration())
	}

	return &ast.File{
		ModuleName:   moduleName,
		Declarations: declarations,
		Errors:       p.errors,
	}

}
