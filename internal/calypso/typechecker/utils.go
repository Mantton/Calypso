package typechecker

import (
	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/token"
)

type Literal byte

const (
	INTEGER Literal = iota
	FLOAT
	STRING
	BOOLEAN
	ARRAY
	MAP
	NULL
	VOID
	ANY
)

func (c *Checker) isType(t SymbolType) bool {
	return t == TypeSymbol || t == AliasSymbol || t == StandardSymbol || t == GenericTypeSymbol
}

func (c *Checker) addError(msg string, pos token.SyntaxRange) {
	c.Errors.Add(lexer.Error{
		Message: msg,
		Range:   pos,
	})
}

func (c *Checker) injectLiterals() {
	if c.mode != STD {
		return
	}

	integerLit := newSymbolInfo("IntegerLiteral", TypeSymbol)
	floatLit := newSymbolInfo("FloatLiteral", TypeSymbol)
	stringLit := newSymbolInfo("StringLiteral", TypeSymbol)
	booleanLit := newSymbolInfo("BooleanLiteral", TypeSymbol)

	voidLit := newSymbolInfo("void", TypeSymbol)
	nullLit := newSymbolInfo("null", TypeSymbol)

	anyLit := newSymbolInfo("any", TypeSymbol)

	c.define(integerLit)
	c.define(floatLit)
	c.define(stringLit)
	c.define(booleanLit)
	c.define(nullLit)
	c.define(voidLit)
	c.define(anyLit)

	arrayLit := newSymbolInfo("ArrayLiteral", TypeSymbol)
	err := arrayLit.addGenericArgument(anyLit)

	if err != nil {
		panic(err)
	}

	c.define(arrayLit)
}

func (c *Checker) resolveLiteral(l Literal) *SymbolInfo {
	var name string
	if c.mode == STD {

		switch l {
		case INTEGER:
			name = "IntegerLiteral"
		case FLOAT:
			name = "FloatLiteral"
		case STRING:
			name = "StringLiteral"
		case BOOLEAN:
			name = "BooleanLiteral"
		case NULL:
			name = "null"
		case VOID:
			name = "void"
		case ANY:
			name = "any"
		default:
			panic("unresolved literal")
		}

		desc, ok := c.find(name)

		if !ok {
			panic("Unable to locate literal type")
		}

		return desc

	}

	panic("unable to resolve literal")
}

func (c *Checker) resolveGenericArgument(n string) (*SymbolInfo, bool) {
	if c.currentSym == nil {
		return c.find(n)
	}

	for _, arg := range c.currentSym.GenericParams {
		if arg.Name == n {
			return arg, true
		}
	}

	return c.find(n)

}
