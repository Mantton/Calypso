package t

import (
	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/token"
)

func (c *Checker) addError(msg string, pos token.SyntaxRange) {
	c.Errors.Add(lexer.Error{
		Message: msg,
		Range:   pos,
	})
}
