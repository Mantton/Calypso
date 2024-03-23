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
	op_b  // Operators Begin
	PLUS  // +
	MINUS // -
	STAR  // *
	QUO   // /
	PCT   // %
	AMP   // &

	L_CHEVRON // <
	R_CHEVRON // >
	EQL       // ==
	NEQ       // !=

	LEQ // <=
	GEQ // >=

	PLUS_EQ  // +=
	MINUS_EQ // -=
	STAR_EQ  // *=
	QUO_EQ   // /=
	PCT_EQ   // %=

	AMP_EQ             // &=
	BAR_EQ             // |=
	CARET_EQ           // ^=
	BIT_SHIFT_LEFT_EQ  // <<=
	BIT_SHIFT_RIGHT_EQ // >>=

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

	DOUBLE_AMP      // &&
	DOUBLE_BAR      // ||
	BAR             // |
	CARET           // ^
	BIT_SHIFT_RIGHT // >>
	BIT_SHIFT_LEFT  // <<
	PIPE            // |>
	R_ARROW         // ->
	op_e            // Operators End

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
	BREAK
	/// Modifiers
	ASYNC
	STATIC
	MUTATING
	PUB
	kw_e // Keywords End
)

func (t ScannedToken) String() string {

	return fmt.Sprintf("{'%s' @ %d}", t.Lit, t.Pos)
}

var keywords = map[string]Token{
	"fn":        FUNC,
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
	"break":     BREAK,
	"pub":       PUB,
	"static":    STATIC,
	"mutating":  MUTATING,
	"async":     ASYNC,
}

func LookupIdent(ident string) Token {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENTIFIER
}

func IsDeclaration(t Token) bool {
	switch t {
	case FUNC, CONST, MODULE, STANDARD, EXTENSION, CONFORM, EXTERN:
		return true
	}
	return false
}

func IsStatement(t Token) bool {
	switch t {
	case FUNC, LET, CONST, IF, RETURN, STRUCT, FOR, ENUM, SWITCH, BREAK, WHILE, TYPE:
		return true
	}

	return false
}

func IsModifier(t Token) bool {
	switch t {
	case ASYNC, STATIC, MUTATING, PUB:
		return true
	}

	return false
}

func IsModifiable(t Token) bool {
	switch t {
	case STRUCT, FUNC, CONST, LET, TYPE, STANDARD, ENUM:
		return true
	}
	return false
}

var tokens = map[Token]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	PLUS:  "+",
	MINUS: "-",
	STAR:  "*",
	QUO:   "/",
	PCT:   "%",
	AMP:   "&",

	BIT_SHIFT_LEFT:  "<<",
	BIT_SHIFT_RIGHT: ">>",
	CARET:           "^",
	BAR:             "|",

	DOUBLE_BAR: "||",
	DOUBLE_AMP: "&&",

	EQL:       "==",
	L_CHEVRON: "<",
	R_CHEVRON: ">",
	ASSIGN:    "=",
	NOT:       "!",

	PLUS_EQ:  "+=", // +=
	MINUS_EQ: "-=", // -=
	STAR_EQ:  "*=", // *=
	QUO_EQ:   "/=", // /=
	PCT_EQ:   "%=", // %=

	AMP_EQ:             "&=",  // &=
	BAR_EQ:             "|=",  // |=
	CARET_EQ:           "^=",  // ^=
	BIT_SHIFT_LEFT_EQ:  "<<=", // <<=
	BIT_SHIFT_RIGHT_EQ: ">>=", // >>=

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
	FUNC:      "fn",
	LET:       "let",
	IF:        "if",
	ELSE:      "else",
	RETURN:    "return",
	WHILE:     "while",
	STANDARD:  "standard",
	TYPE:      "type",
	STRUCT:    "struct",
	EXTENSION: "extension",
	CONFORM:   "conform",
	FOR:       "for",
	TO:        "to",
	SWITCH:    "switch",
	DEFAULT:   "default",
	BREAK:     "break",

	PUB:      "public",
	STATIC:   "static",
	MUTATING: "mutating",
	ASYNC:    "async",

	IDENTIFIER: "IDENTIFIER",
}

var ModifierPrecedent = map[Token]int{
	// Generic Mods
	ASYNC: 2,

	// Nested Function Mods
	STATIC:   3,
	MUTATING: 3,

	// Visibility Modifiers
	PUB: 0,
}

func LookUp(t Token) string {
	v, ok := tokens[t]

	if ok {
		return v
	}

	return fmt.Sprintf("UNKNOWN %d", t)
}

func String(t Token) string {
	return LookUp(t)
}
