package parser

import (
	"fmt"

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
		panic(fmt.Errorf("expected `%s` got `%s`", token.LookUp(t), token.LookUp(p.current())))
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
