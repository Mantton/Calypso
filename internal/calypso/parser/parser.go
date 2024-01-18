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

	moduleName := "unknown"
	// - Parse Module Declaration
	// module main
	p.expect(token.MODULE)            // Consume module token
	tok := p.expect(token.IDENTIFIER) // Identifier
	moduleName = tok.Lit
	p.expect(token.SEMICOLON) // Consume semicolon

	// Imports

	// Declarations
	var declarations []ast.Declaration

	for p.current() != token.EOF {

		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("DECL ERROR: ", r)
					p.advance(token.IsDeclaration)
				}
			}()
			declarations = append(declarations, p.parseDeclaration())
		}()

	}

	return &ast.File{
		ModuleName:   moduleName,
		Declarations: declarations,
		Errors:       p.errors,
	}

}
