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
	var functions []*ast.FunctionDeclaration
	var constants []*ast.ConstantDeclaration

	for p.current() != token.EOF {

		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("DECL ERROR: ", r)
					hasMoved := p.advance(token.IsDeclaration)

					// avoid infinite loop
					if !hasMoved {
						p.next()
					}
				}
			}()

			decl := p.parseDeclaration()

			switch decl := decl.(type) {
			case *ast.ConstantDeclaration:
				constants = append(constants, decl)
			case *ast.FunctionDeclaration:
				functions = append(functions, decl)
			default:
				panic("unknown declaration")
			}
		}()

	}

	return &ast.File{
		ModuleName: moduleName,
		Errors:     p.errors,
		Functions:  functions,
		Constants:  constants,
	}

}
