package typechecker

import (
	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/symbols"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (c *Checker) isType(t symbols.SymbolType) bool {
	return t == symbols.TypeSymbol || t == symbols.AliasSymbol || t == symbols.StandardSymbol || t == symbols.GenericTypeSymbol
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

	integerLit := symbols.NewSymbol("IntegerLiteral", symbols.TypeSymbol)
	floatLit := symbols.NewSymbol("FloatLiteral", symbols.TypeSymbol)
	stringLit := symbols.NewSymbol("StringLiteral", symbols.TypeSymbol)
	booleanLit := symbols.NewSymbol("BooleanLiteral", symbols.TypeSymbol)

	voidLit := symbols.NewSymbol("void", symbols.TypeSymbol)
	nullLit := symbols.NewSymbol("null", symbols.TypeSymbol)

	anyLit := symbols.NewSymbol("any", symbols.TypeSymbol)

	c.define(integerLit)
	c.define(floatLit)
	c.define(stringLit)
	c.define(booleanLit)
	c.define(nullLit)
	c.define(voidLit)
	c.define(anyLit)

	arrayLit := symbols.NewSymbol("ArrayLiteral", symbols.TypeSymbol)
	genericVal := symbols.NewSymbol("T", symbols.GenericTypeSymbol)
	err := arrayLit.AddGenericParameter(genericVal)

	if err != nil {
		panic(err)
	}

	c.define(arrayLit)
}

func (c *Checker) resolveLiteral(l symbols.Literal) *symbols.SymbolInfo {
	var name string
	if c.mode == STD {

		switch l {
		case symbols.INTEGER:
			name = "IntegerLiteral"
		case symbols.FLOAT:
			name = "FloatLiteral"
		case symbols.STRING:
			name = "StringLiteral"
		case symbols.BOOLEAN:
			name = "BooleanLiteral"
		case symbols.NULL:
			name = "null"
		case symbols.VOID:
			name = "void"
		case symbols.ANY:
			name = "any"
		case symbols.ARRAY:
			name = "ArrayLiteral"
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

func (c *Checker) resolveGenericArgument(n string) (*symbols.SymbolInfo, bool) {
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
