package parser

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (p *Parser) parseTypeExpression() (ast.TypeExpression, error) {

	var typ ast.TypeExpression
	var err error
	switch p.current() {
	case token.STAR:
		typ, err = p.parsePointerTypeExpression()
		if err != nil {
			return nil, err
		}
	case token.IDENTIFIER:
		typ, err = p.parseIdentifierWithoutAnnotation()
		if err != nil {
			return nil, err
		}
	case token.LBRACE:
		typ, err = p.parseMapTypeExpression()
		if err != nil {
			return nil, err
		}
	default:
		return nil, p.error("expected type expression")
	}

	if p.match(token.LBRACKET) {
		lBrackPos := p.previousScannedToken().Pos
		rBrack, err := p.expect(token.RBRACKET)
		if err != nil {
			return nil, err
		}

		typ = &ast.ArrayTypeExpression{
			LBracketPos: lBrackPos,
			RBracketPos: rBrack.Pos,
			Element:     typ,
		}
	}

	// specialization
	var args *ast.GenericArgumentsClause
	if p.currentMatches(token.L_CHEVRON) {
		args, err = p.parseGenericArgumentsClause()
		if err != nil {
			return nil, err
		}
	}

	if args != nil {
		return &ast.SpecializationExpression{
			Expression: typ,
			Clause:     args,
		}, nil
	}

	// Is field access
	if p.match(token.PERIOD) {
		f, err := p.parseTypeExpression()
		if err != nil {
			return nil, err
		}

		typ = &ast.FieldAccessExpression{
			Target: typ,
			Field:  f,
		}
	}

	return typ, nil
}

func (p *Parser) parseMapTypeExpression() (ast.TypeExpression, error) {
	start, err := p.expect(token.LBRACE)
	if err != nil {
		return nil, err
	}

	expr, err := p.parseTypeExpression()
	if err != nil {
		return nil, err
	}

	_, err = p.expect(token.COLON)
	if err != nil {
		return nil, err
	}

	value, err := p.parseTypeExpression()
	if err != nil {
		return nil, err
	}

	end, err := p.expect(token.RBRACE)
	if err != nil {
		return nil, err
	}

	return &ast.MapTypeExpression{
		Key:         expr,
		Value:       value,
		LBracketPos: start.Pos,
		RBracketPos: end.Pos,
	}, nil
}

/*
This parses a generic argument clause

# Example

`const foo: GenericType<int, string>`
*/
func (p *Parser) parseGenericArgumentsClause() (*ast.GenericArgumentsClause, error) {

	args := []ast.TypeExpression{}
	start, err := p.expect(token.L_CHEVRON)
	if err != nil {
		return nil, err
	}

	if p.match(token.R_CHEVRON) {
		return nil, p.error("expected at least 1 generic argument")
	}

	// First Argument
	expr, err := p.parseTypeExpression()
	if err != nil {
		return nil, err
	}

	args = append(args, expr)

	// Check For Others
	for p.match(token.COMMA) {

		if p.match(token.R_CHEVRON) {
			return nil, p.error("expected type expression")
		}

		expr, err := p.parseTypeExpression()
		if err != nil {
			return nil, err
		}

		args = append(args, expr)
	}

	end, err := p.expect(token.R_CHEVRON)
	if err != nil {
		return nil, err
	}

	if len(args) == 0 {
		return nil, p.error("expected arguments")
	}

	return &ast.GenericArgumentsClause{
		LChevronPos: start.Pos,
		RChevronPos: end.Pos,
		Arguments:   args,
	}, nil
}

/*
This parses a generic parameter clause

# Example

`alias set<T : Foo> = set<T>`
*/
func (p *Parser) parseGenericParameterClause() (*ast.GenericParametersClause, error) {
	params := []*ast.GenericParameterExpression{}
	start, err := p.expect(token.L_CHEVRON)
	if err != nil {
		return nil, err
	}

	if p.match(token.R_CHEVRON) {
		return nil, p.error("expected at least 1 generic parameter")
	}

	// First Argument
	param, err := p.parseGenericParameterExpression()
	if err != nil {
		return nil, err
	}
	params = append(params, param)

	// Check For Others
	for p.match(token.COMMA) {

		if p.match(token.R_CHEVRON) {
			return nil, p.error("expected generic parameter")
		}

		param, err := p.parseGenericParameterExpression()
		if err != nil {
			return nil, err
		}

		params = append(params, param)
	}

	end, err := p.expect(token.R_CHEVRON)
	if err != nil {
		return nil, err
	}

	if len(params) == 0 {
		return nil, p.error("expected at least 1 generic parameter")
	}

	return &ast.GenericParametersClause{
		Parameters:  params,
		LChevronPos: start.Pos,
		RChevronPos: end.Pos,
	}, nil
}

/*
Parses A Generic Parameter

# Example

`alias set<T : Foo & Bar & Baz> = set<T>`

It will parse the `T : Foo & Bar & Baz` Parameter
*/
func (p *Parser) parseGenericParameterExpression() (*ast.GenericParameterExpression, error) {
	ident, err := p.parseIdentifierWithoutAnnotation()
	if err != nil {
		return nil, err
	}
	standards := []*ast.IdentifierExpression{}

	// parse standards
	if p.match(token.COLON) {
		// First standard
		standard, err := p.parseIdentifierWithoutAnnotation()
		if err != nil {
			return nil, err
		}
		standards = append(standards, standard)

		// Others
		for p.match(token.AMP) {
			standard, err := p.parseIdentifierWithoutAnnotation()
			if err != nil {
				return nil, err
			}
			standards = append(standards, standard)
		}
	}
	return &ast.GenericParameterExpression{
		Identifier: ident,
		Standards:  standards,
	}, nil
}

func (p *Parser) parsePointerTypeExpression() (*ast.PointerTypeExpression, error) {
	pos, err := p.expect(token.STAR)
	if err != nil {
		return nil, err
	}

	expr, err := p.parseTypeExpression()
	if err != nil {
		return nil, err
	}

	return &ast.PointerTypeExpression{
		StarPos:   pos.Pos,
		PointerTo: expr,
	}, nil
}
