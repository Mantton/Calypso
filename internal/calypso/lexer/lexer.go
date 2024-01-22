package lexer

import (
	"github.com/mantton/calypso/internal/calypso/token"
)

const (
	eof = -1
)

type Lexer struct {
	source       []rune // an array of each rune in the file
	sourceLength int

	anchor int // the start of the token
	cursor int // the current position in the source

	line       int // the current line
	lineOffset int
}

func New(input string) *Lexer {
	l := &Lexer{source: []rune(input)}
	l.sourceLength = len(l.source)
	l.line = 1
	l.lineOffset = 1

	return l
}

func (l *Lexer) AllTokens() []token.ScannedToken {
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
	return tokens
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
		} else {
			tok = l.build(token.LSS)
		}

	case '>':
		if l.match('=') {
			tok = l.build(token.GEQ)
		} else {
			tok = l.build(token.GTR)
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
