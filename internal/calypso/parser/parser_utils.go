package parser

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (p *Parser) current() token.Token {
	return p.file.Tokens[p.cursor].Tok
}

func (p *Parser) isAtEnd() bool {
	return p.cursor == len(p.file.Tokens)-1
}

func (p *Parser) currentScannedToken() token.ScannedToken {
	return p.file.Tokens[p.cursor]
}

func (p *Parser) peakAheadScannedToken() (token.ScannedToken, bool) {
	if p.isAtEnd() {
		return token.ScannedToken{}, false
	}
	return p.file.Tokens[p.cursor+1], true
}

func (p *Parser) previousScannedToken() token.ScannedToken {
	return p.file.Tokens[p.cursor-1]
}

// bool indicating the current token is of the specified type
func (p *Parser) currentMatches(t token.Token) bool {
	return p.current() == t
}

func (p *Parser) match(tokens ...token.Token) bool {

	for _, tok := range tokens {
		if p.currentMatches(tok) {
			p.next()
			return true
		}
	}

	return false
}

func (p *Parser) expect(t token.Token) (token.ScannedToken, error) {
	if p.current() != t {
		return token.ScannedToken{}, p.error(fmt.Sprintf("expected `%s`", token.LookUp(t)))
	} else {
		defer p.next()
		return p.currentScannedToken(), nil
	}
}

func (p *Parser) next() {
	if p.isAtEnd() {
		return
	}
	p.cursor++
}

func (p *Parser) advance(check token.NodeChecker) bool {
	moves := 0
	for p.current() != token.EOF {
		if check(p.current()) {
			break
		} else {
			moves += 1
			p.next()
		}
	}

	return moves != 0
}

func (p *Parser) error(message string) *lexer.CompilerError {
	end, ok := p.peakAheadScannedToken()

	if !ok {
		end = p.currentScannedToken()
	}
	return &lexer.CompilerError{
		File: p.file,
		Range: token.SyntaxRange{
			Start: p.currentScannedToken().Pos,
			End:   end.Pos,
		},
		Message: message,
	}

}

func (p *Parser) errorAt(message string, r token.SyntaxRange) *lexer.CompilerError {

	return &lexer.CompilerError{
		File:    p.file,
		Range:   r,
		Message: message,
	}

}
