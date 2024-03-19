/*
Copyright (c) 2009 The Go Authors. All rights reserved.
*/

package lexer

import (
	"fmt"
	"unicode"

	"github.com/mantton/calypso/internal/calypso/token"
)

const (
	eof = -1
)

type Lexer struct {
	file         *File
	source       []rune // an array of each rune in the file
	sourceLength int

	anchor int // the start of the token
	cursor int // the current position in the source

	line       int // the current line
	lineOffset int
}

func New(file *File) *Lexer {
	l := &Lexer{source: file.Chars}
	l.file = file
	l.sourceLength = file.Length
	l.line = 1
	l.lineOffset = 1

	return l
}

func (l *Lexer) ScanAll() {
	tokens := []token.ScannedToken{}

	for !l.isAtEnd() {
		// at the start of next lexeme, drop anchor
		l.anchor = l.cursor

		// parse next token
		tok := l.parseToken()

		if tok.Tok == token.IGNORE {
			continue
		}

		// append to token list
		tokens = append(tokens, tok)
	}

	// Add EOF token
	tokens = append(tokens, token.ScannedToken{Pos: l.genPosition(), Tok: token.EOF, Lit: "EOF"})
	l.file.Tokens = tokens
}

func (l *Lexer) isAtEnd() bool {
	return l.cursor >= l.sourceLength
}

func (l *Lexer) next() rune {
	c := l.source[l.cursor]
	l.cursor++
	l.lineOffset++
	return c
}

func (l *Lexer) parseToken() token.ScannedToken {
	c := l.next() // get next char
	tok := l.build(token.ILLEGAL)

	switch c {

	case ' ', '\r', '\t':
		// Ignore whitespace.
		tok = l.build(token.IGNORE)
	case '\n':
		l.newLine()
		tok = l.build(token.IGNORE)
	case '(':
		tok = l.build(token.LPAREN)
	case ')':
		tok = l.build(token.RPAREN)
	case '[':
		tok = l.build(token.LBRACKET)
	case ']':
		tok = l.build(token.RBRACKET)
	case '{':
		tok = l.build(token.LBRACE)
	case '}':
		tok = l.build(token.RBRACE)
	case ';':
		tok = l.build(token.SEMICOLON)
	case ':':
		tok = l.build(token.COLON)
	case ',':
		tok = l.build(token.COMMA)
	case '.':
		tok = l.build(token.PERIOD)

	// * Operators
	case '-':
		if l.match('>') {
			tok = l.build(token.R_ARROW)
		} else {
			tok = l.build(token.SUB)
		}
	case '+':
		tok = l.build(token.ADD)
	case '*':
		tok = l.build(token.MUL)

	case '!':
		if l.match('=') {
			tok = l.build(token.NEQ)
		} else {
			tok = l.build(token.NOT)

		}

	case '=':
		if l.match('=') {
			tok = l.build(token.EQL)
		} else {
			tok = l.build(token.ASSIGN)
		}

	case '<':
		if l.match('=') {
			tok = l.build(token.LEQ)
		} else if l.match('<') {
			tok = l.build(token.BIT_SHIFT_LEFT)
		} else {
			tok = l.build(token.L_CHEVRON)
		}

	case '>':
		if l.match('=') {
			tok = l.build(token.GEQ)
		} else if l.match('>') {
			tok = l.build(token.BIT_SHIFT_RIGHT)
		} else {
			tok = l.build(token.R_CHEVRON)
		}

	case '/':
		// Comment, match to end of line
		if l.match('/') {
			for l.peek() != '\n' && !l.isAtEnd() {
				l.next()
			}

			tok = l.build(token.IGNORE)
		} else {
			tok = l.build(token.QUO)
		}
	// * Literals
	case '"':
		tok = l.string()
	case '\'':
		tok = l.char()
	case '&':
		if l.match('&') {
			tok = l.build(token.AND)
		} else {
			tok = l.build(token.AMP)
		}
	case '|':
		if l.match('|') {
			tok = l.build(token.OR)
		} else if l.match('>') {
			tok = l.build(token.PIPE)
		} else {
			tok = l.build(token.BAR)
		}
	case '^':
		tok = l.build(token.CARET)
	default:
		if isDigit(c) {
			tok = l.number()
		} else if isLetter(c) {
			tok = l.identifier()
		}
	}

	return tok
}

func (l *Lexer) build(t token.Token) token.ScannedToken {
	return token.ScannedToken{
		Pos: l.genPosition(),
		Tok: t,
		Lit: string(l.source[l.anchor:l.cursor]),
	}
}

func (l *Lexer) match(expected rune) bool {
	if l.isAtEnd() {
		return false
	}

	if l.source[l.cursor] != expected {
		return false
	}

	l.cursor++
	return true
}

func (l *Lexer) peek() rune {
	if l.isAtEnd() {
		return eof
	}

	return l.source[l.cursor]
}

func (l *Lexer) char() token.ScannedToken {
	// Reference: https://cs.opensource.google/go/go/+/refs/tags/go1.22.0:src/go/scanner/scanner.go;l=609
	n := 0

	for {
		ch := l.peek()

		// if new line or invalid character
		if ch == '\n' || ch < 0 {
			panic("char literal not terminated")
		}

		// Move to next token
		l.next()

		// closing quote
		if ch == '\'' {
			break
		}

		// increment char count
		n++

		// Scan/Parse Escape Character
		if ch == '\\' {
			l.scanEscape('\'')
		}

		// read to closing quote

	}

	if n != 1 {
		panic("invalid char literal")
	}

	s := string(l.source[l.anchor:l.cursor])

	return token.ScannedToken{
		Lit: s,
		Tok: token.CHAR,
		Pos: l.genPosition(),
	}
}

// REFERENCE : https://cs.opensource.google/go/go/+/refs/tags/go1.22.0:src/go/scanner/scanner.go;l=556
func (l *Lexer) scanEscape(br rune) {
	var n int
	var base, max uint32
	switch l.peek() {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', br:
		l.next()
	case '0', '1', '2', '3', '4', '5', '6', '7':
		n, base, max = 3, 8, 255
	case 'x':
		l.next()
		n, base, max = 2, 16, 255
	case 'u':
		l.next()
		n, base, max = 4, 16, unicode.MaxRune
	case 'U':
		l.next()
		n, base, max = 8, 16, unicode.MaxRune
	default:
		msg := "unknown escape sequence"
		if l.peek() < 0 {
			msg = "escape sequence not terminated"
		}
		panic(msg)
	}

	var x uint32
	for n > 0 {
		d := uint32(digitVal(l.peek()))
		if d >= base {
			msg := fmt.Sprintf("illegal character %#U in escape sequence", l.peek())
			if l.peek() < 0 {
				msg = "escape sequence not terminated"
			}
			panic(msg)
		}
		x = x*base + d
		l.next()
		n--
	}

	if x > max || 0xD800 <= x && x < 0xE000 {
		panic("escape sequence is invalid Unicode code point")
	}
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= lower(ch) && lower(ch) <= 'f':
		return int(lower(ch) - 'a' + 10)
	}
	return 16 // larger than any legal digit val
}

func lower(ch rune) rune { return ('a' - 'A') | ch } // returns lower-case ch iff ch is ASCII letter

func (l *Lexer) string() token.ScannedToken {
	for l.peek() != '"' && !l.isAtEnd() {
		if l.peek() == '\n' {
			l.newLine()
		}

		l.next()
	}

	if l.isAtEnd() {
		panic("unterminated string")
	}

	l.next()

	str := string(l.source[l.anchor+1 : l.cursor-1])

	return token.ScannedToken{
		Lit: str,
		Tok: token.STRING,
		Pos: l.genPosition(),
	}
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) number() token.ScannedToken {

	tType := token.INTEGER
	for isDigit(l.peek()) {
		l.next()
	}

	// if the current tok is a period, look at float
	if l.peek() == '.' && isDigit(l.peekAhead()) {
		tType = token.FLOAT

		// consume '.'
		l.next()

		for isDigit(l.peek()) {
			l.next()
		}
	}
	str := string(l.source[l.anchor:l.cursor])

	return token.ScannedToken{
		Lit: str,
		Tok: tType,
		Pos: l.genPosition(),
	}
}

func (l *Lexer) peekAhead() rune {

	idx := l.cursor + 1

	if idx >= l.sourceLength {
		return eof
	}

	return l.source[idx]
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isAlphaNumeric(c rune) bool {
	return isLetter(c) || isDigit(c)
}

func (l *Lexer) identifier() token.ScannedToken {
	for isAlphaNumeric(l.peek()) {
		l.next()
	}

	tok := l.build(token.IDENTIFIER)

	tok.Tok = token.LookupIdent(tok.Lit)

	return tok
}

func (l *Lexer) newLine() {
	l.line++
	l.lineOffset = 1
}

func (l *Lexer) genPosition() token.TokenPosition {
	return token.TokenPosition{
		Line:   l.line,
		Offset: l.lineOffset,
		Start:  l.anchor,
		End:    l.cursor + 1,
	}
}
