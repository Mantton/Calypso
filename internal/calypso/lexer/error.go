package lexer

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/token"
)

type CompilerError struct {
	File    *File
	Range   token.SyntaxRange
	Message string
}

type ErrorList []error

func (l *ErrorList) Add(err error) {
	*l = append(*l, err)
}

func (e *CompilerError) Error() string {
	msg := fmt.Sprintf("\n%s:%d:%d -> %s", e.File.Path, e.Range.Start.Line, e.Range.Start.Offset, e.Message)
	msg += fmt.Sprintf("\n\t%s", e.File.Lines[max(0, e.Range.Start.Line-1)])
	// TODO: Arrow
	return msg
}
