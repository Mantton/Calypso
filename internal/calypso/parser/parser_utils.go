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

func (p *Parser) match(tokens ...token.Token) bool {

	for _, tok := range tokens {
		if p.currentMatches(tok) {
			p.next()
			return true
		}
	}

	return false
}

func (p *Parser) isAtEnd() bool {
	return p.peekAhead() == token.EOF
}

func (p *Parser) expect(t token.Token) {
	if p.current() != t {
		fmt.Println(p.currentScannedToken(), t)
		panic("expected diff token")
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
