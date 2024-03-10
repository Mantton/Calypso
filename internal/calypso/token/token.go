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
	Start  int
	End    int
}

type SyntaxRange struct {
	Start TokenPosition
	End   TokenPosition
}

func (p TokenPosition) Length() int {
	return p.End - p.Start
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
	CHAR
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
	AMP  // &

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

	R_ARROW // ->
	op_e    // Operators End

	//* KEYWORDS
	kw_b // Keywords Begin
	FUNC
	CONST
	LET
	TRUE
	FALSE
	NIL
	VOID
	MODULE
	IF
	ELSE
	RETURN
	WHILE
	ALIAS
	STANDARD
	TYPE
	STRUCT
	EXTENSION
	CONFORM
	FOR
	TO
	EXTERN
	SWITCH
	ENUM
	CASE
	DEFAULT
	/// Modifiers
	ASYNC
	STATIC
	MUT
	PUB
	PRIVATE
	kw_e // Keywords End
)

func (t ScannedToken) String() string {

	return fmt.Sprintf("{'%s' @ %d}", t.Lit, t.Pos)
}

var keywords = map[string]Token{
	"func":      FUNC,
	"let":       LET,
	"const":     CONST,
	"module":    MODULE,
	"true":      TRUE,
	"false":     FALSE,
	"nil":       NIL,
	"void":      VOID,
	"if":        IF,
	"else":      ELSE,
	"return":    RETURN,
	"while":     WHILE,
	"alias":     ALIAS,
	"standard":  STANDARD,
	"type":      TYPE,
	"struct":    STRUCT,
	"extension": EXTENSION,
	"conform":   CONFORM,
	"for":       FOR,
	"to":        TO,
	"extern":    EXTERN,
	"enum":      ENUM,
	"switch":    SWITCH,
	"case":      CASE,
	"default":   DEFAULT,
}

func LookupIdent(ident string) Token {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENTIFIER
}

func IsDeclaration(t Token) bool {
	switch t {
	case FUNC, CONST, MODULE, STANDARD, TYPE, EXTENSION, CONFORM, EXTERN:
		return true
	}
	return false
}

func IsStatement(t Token) bool {
	switch t {
	case FUNC, LET, CONST, IF, RETURN, ALIAS, STRUCT, FOR, ENUM, SWITCH:
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
	AMP: "&",

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
	R_ARROW:   "->",

	CONST:     "const",
	FUNC:      "func",
	LET:       "let",
	IF:        "if",
	ELSE:      "else",
	RETURN:    "return",
	WHILE:     "while",
	ALIAS:     "alias",
	STANDARD:  "standard",
	TYPE:      "type",
	STRUCT:    "struct",
	EXTENSION: "extension",
	CONFORM:   "conform",
	FOR:       "for",
	TO:        "to",
	SWITCH:    "switch",
	DEFAULT:   "default",

	IDENTIFIER: "IDENTIFIER",
}

func LookUp(t Token) string {
	v, ok := tokens[t]

	if ok {
		return v
	}

	return fmt.Sprintf("UNKNOWN %d", t)
}
