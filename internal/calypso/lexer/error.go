package lexer

import "github.com/mantton/calypso/internal/calypso/token"

type Error struct {
	Range   token.SyntaxRange
	Message string
}

type ErrorList []*Error

func (l *ErrorList) Add(err Error) {
	*l = append(*l, &err)
}
