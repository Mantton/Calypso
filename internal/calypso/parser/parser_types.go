package parser

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (p *Parser) parseTypeExpression() ast.TypeExpression {

	var typ ast.TypeExpression
	switch p.current() {
	case token.MUL:
		typ = p.parsePointerTypeExpression()
	case token.IDENTIFIER:
		typ = p.parseIdentifierTypeExpression()
	case token.LBRACE:
		typ = p.parseMapTypeExpression()
	default:
		panic(p.error("expected type expression"))
	}

	if p.match(token.LBRACKET) {
		lBrackPos := p.previousScannedToken().Pos
		rBrackPos := p.expect(token.RBRACKET).Pos

		return &ast.ArrayTypeExpression{
			LBracketPos: lBrackPos,
			RBracketPos: rBrackPos,
			Element:     typ,
		}
	}

	return typ
}

func (p *Parser) parseMapTypeExpression() ast.TypeExpression {
	start := p.expect(token.LBRACE)
	expr := p.parseTypeExpression()
	p.expect(token.COLON)
	value := p.parseTypeExpression()
	end := p.expect(token.RBRACE)

	return &ast.MapTypeExpression{
		Key:         expr,
		Value:       value,
		LBracketPos: start.Pos,
		RBracketPos: end.Pos,
	}
}

func (p *Parser) parseIdentifierTypeExpression() ast.TypeExpression {

	// 1 - Identifier
	ident := p.parseIdentifierWithoutAnnotation()

	// 2 - Type Parameters
	var args *ast.GenericArgumentsClause
	if p.currentMatches(token.LSS) {
		args = p.parseGenericArgumentsClause()
	}

	return &ast.IdentifierTypeExpression{
		Identifier: ident,
		Arguments:  args,
	}

}

/*
This parses a generic argument clause

# Example

`const foo: GenericType<int, string>`
*/
func (p *Parser) parseGenericArgumentsClause() *ast.GenericArgumentsClause {

	args := []ast.TypeExpression{}
	start := p.expect(token.LSS)

	if p.match(token.GTR) {
		panic(p.error("expected at least 1 generic argument"))
	}

	// First Argument
	expr := p.parseTypeExpression()
	args = append(args, expr)

	// Check For Others
	for p.match(token.COMMA) {

		if p.match(token.GTR) {
			panic("expected type expression")
		}

		expr := p.parseTypeExpression()

		args = append(args, expr)
	}

	end := p.expect(token.GTR)

	if len(args) == 0 {
		panic("expected arguments")
	}

	return &ast.GenericArgumentsClause{
		LChevronPos: start.Pos,
		RChevronPos: end.Pos,
		Arguments:   args,
	}
}

/*
This parses a generic parameter clause

# Example

`alias set<T : Foo> = set<T>`
*/
func (p *Parser) parseGenericParameterClause() *ast.GenericParametersClause {
	params := []*ast.GenericParameterExpression{}
	start := p.expect(token.LSS)

	if p.match(token.GTR) {
		panic(p.error("expected at least 1 generic parameter"))
	}

	// First Argument
	param := p.parseGenericParameterExpression()
	params = append(params, param)

	// Check For Others
	for p.match(token.COMMA) {

		if p.match(token.GTR) {
			panic(p.error("expected generic parameter"))
		}

		param := p.parseGenericParameterExpression()

		params = append(params, param)
	}

	end := p.expect(token.GTR)

	if len(params) == 0 {
		panic(p.error("expected at least 1 generic parameter"))
	}

	return &ast.GenericParametersClause{
		Parameters:  params,
		LChevronPos: start.Pos,
		RChevronPos: end.Pos,
	}
}

/*
Parses A Generic Parameter

# Example

`alias set<T : Foo & Bar & Baz> = set<T>`

It will parse the `T : Foo & Bar & Baz` Parameter
*/
func (p *Parser) parseGenericParameterExpression() *ast.GenericParameterExpression {
	ident := p.parseIdentifierWithoutAnnotation()
	standards := []*ast.IdentifierExpression{}

	// parse standards
	if p.match(token.COLON) {
		// First standard
		standard := p.parseIdentifierWithoutAnnotation()
		standards = append(standards, standard)

		// Others
		for p.match(token.AMP) {
			standard := p.parseIdentifierWithoutAnnotation()
			standards = append(standards, standard)
		}
	}
	return &ast.GenericParameterExpression{
		Identifier: ident,
		Standards:  standards,
	}
}

func (p *Parser) parsePointerTypeExpression() *ast.PointerTypeExpression {
	pos := p.expect(token.MUL)

	expr := p.parseTypeExpression()

	return &ast.PointerTypeExpression{
		StarPos:   pos.Pos,
		PointerTo: expr,
	}
}
