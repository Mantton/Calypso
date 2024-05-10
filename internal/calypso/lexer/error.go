package lexer

import (
	"errors"
	"fmt"

	"github.com/mantton/calypso/internal/calypso/token"
)

type CompilerError struct {
	File    *File
	Range   token.SyntaxRange
	Message string
}

func NewError(msg string, pos token.SyntaxRange, file *File) *CompilerError {
	return &CompilerError{
		Message: msg,
		Range:   pos,
		File:    file,
	}
}

type ErrorList []error

func (l *ErrorList) Add(err error) {
	*l = append(*l, err)
}

func (l *ErrorList) String() string {
	s := ""
	for _, e := range *l {
		s += e.Error()
	}

	return s
}

func (e *CompilerError) Error() string {
	msg := fmt.Sprintf("\n%s:%d:%d -> %s", e.File.Path, e.Range.Start.Line, e.Range.Start.Offset, e.Message)
	msg += fmt.Sprintf("\n\t%s", e.File.Lines[max(0, e.Range.Start.Line-1)])
	// TODO: Arrow
	return msg
}

func CombinedErrors(errs []error) error {
	s := ""
	for _, e := range errs {
		s += e.Error()
	}

	return errors.New(s)
}
