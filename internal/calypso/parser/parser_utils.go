package parser

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (p *Parser) current() token.Token {
	return p.tokens[p.cursor].Tok
}

func (p *Parser) previous() token.Token {
	return p.tokens[p.cursor-1].Tok

}

func (p *Parser) currentScannedToken() token.ScannedToken {
	return p.tokens[p.cursor]
}

func (p *Parser) peakAheadScannedToken() token.ScannedToken {
	return p.tokens[p.cursor+1]
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

func (p *Parser) expect(t token.Token) token.ScannedToken {
	if p.current() != t {
		panic(p.error(fmt.Sprintf("expected `%s`", token.LookUp(t)))) // never executed
	} else {
		defer p.next()
		return p.currentScannedToken()
	}
}
func (p *Parser) next() {
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

func (p *Parser) error(message string) lexer.Error {
	return lexer.Error{
		Start:   p.currentScannedToken().Pos,
		End:     p.peakAheadScannedToken().Pos,
		Message: message,
	}

}
