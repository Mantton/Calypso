package token

import "fmt"

type Token byte

type NodeChecker func(Token) bool

type ScannedToken struct {
	Tok Token
	Pos TokenPosition
	Lit string
}

type TokenPosition struct {
	Line   int
	Offset int
	Index  int
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
	TRUE
	FALSE
	NULL
	VOID
	MODULE
	IF
	ELSE
	RETURN
	WHILE
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
	"true":   TRUE,
	"false":  FALSE,
	"null":   NULL,
	"void":   VOID,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"while":  WHILE,
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

func IsStatement(t Token) bool {
	switch t {
	case FUNC, LET, CONST, IF, RETURN:
		return true
	}

	return false
}

var tokens = map[Token]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	QUO: "/",
	REM: "%",

	EQL:    "==",
	LSS:    "<",
	GTR:    ">",
	ASSIGN: "=",
	NOT:    "!",

	NEQ: "!=",
	LEQ: "<=",
	GEQ: ">=",

	LPAREN: "(",
	LBRACE: "{",
	COMMA:  ",",
	PERIOD: ".",

	RPAREN:    ")",
	RBRACE:    "}",
	SEMICOLON: ";",
	COLON:     ":",

	CONST:  "const",
	FUNC:   "func",
	LET:    "let",
	IF:     "if",
	ELSE:   "else",
	RETURN: "return",
	WHILE:  "while",
}

func LookUp(t Token) string {
	v, ok := tokens[t]

	if ok {
		return v
	}

	return fmt.Sprintf("UNKNOWN %d", t)
}
