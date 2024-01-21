package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/collections"
)

type Mode byte

const (
	//  Standard Library, Certain Restrictions are lifted
	STD Mode = iota

	// User Scripts, This is the standard language
	USER
)

type TypeChecker struct {
	Errors []string
	scopes collections.Stack[*Scope]
	mode   Mode
}

func New(mode Mode) *TypeChecker {
	return &TypeChecker{
		Errors: []string{},
		scopes: collections.Stack[*Scope]{},
		mode:   mode,
	}
}

// * Scopes
func (t *TypeChecker) enterScope() {
	t.scopes.Push(NewScope())

	if t.scopes.Length() > 1000 {
		panic("exceeded max scope depth") // TODO : Error
	}
}

func (t *TypeChecker) leaveScope() {
	t.scopes.Pop()
}

// * Checks
func (t *TypeChecker) CheckFile(file *ast.File) {
	// TODO: Register STD Types?

	t.enterScope() // global enter
	if len(file.Constants) != 0 {

		for _, decl := range file.Constants {
			t.checkDeclaration(decl)
		}

	}

	if len(file.Functions) != 0 {
		for _, decl := range file.Functions {
			t.checkDeclaration(decl)
		}
	}

	t.leaveScope() // global leave
}

func (t *TypeChecker) checkDeclaration(decl ast.Declaration) {
	switch decl := decl.(type) {
	case *ast.ConstantDeclaration:
		stmt := decl.Stmt
		t.checkStatement(stmt)
	case *ast.FunctionDeclaration:
		// expr := decl.Func
		panic("check decl")
	default:
		msg := fmt.Sprintf("declaration check not implemented, %T", decl)
		panic(msg)

	}
}
