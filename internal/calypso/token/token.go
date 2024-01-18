package token

import "fmt"

type Token byte

type NodeChecker func(Token) bool

type ScannedToken struct {
	Tok Token
	Pos int
	Lit string
}

const (
	ILLEGAL Token = iota
	IGNORE
	EOF

	// * LITERALS
	lit_b // Literals Begin
	IDENTIFIER
	INTEGER
	FLOAT
	STRING
	lit_e // Literals End

	ASSIGN // =
	NOT    // !

	//* OPERATORS
	op_b // Operators Begin
	ADD  // +
	SUB  // -
	MUL  // *
	QUO  // /
	REM  // %

	LSS // <
	GTR // >
	EQL // ==
	NEQ // !=

	LEQ // <=
	GEQ // >=

	COMMA     // ,
	PERIOD    // .
	SEMICOLON // ;
	COLON     // :
	LPAREN    // (
	RPAREN    // )
	LBRACE    // {
	RBRACE    // }
	LBRACKET  // [
	RBRACKET  // ]
	op_e      // Operators End

	//* KEYWORDS
	kw_b // Keywords Begin
	FUNC
	CONST
	LET
	MODULE
	kw_e // Keywords End
)

func (t ScannedToken) String() string {

	return fmt.Sprintf("{'%s' @ %d}", t.Lit, t.Pos)
}

var keywords = map[string]Token{
	"func":   FUNC,
	"let":    LET,
	"const":  CONST,
	"module": MODULE,
}

func LookupIdent(ident string) Token {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENTIFIER
}

func IsDeclaration(t Token) bool {
	switch t {
	case FUNC, CONST, MODULE:
		return true
	}
	return false
}
