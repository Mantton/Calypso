package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/collections"
	"github.com/mantton/calypso/internal/calypso/lexer"
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
	t.registerBaseLiterals()
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
	default:
		msg := fmt.Sprintf("declaration check not implemented, %T", decl)
		panic(msg)

	}
}

func (t *TypeChecker) registerBaseLiterals() {
	// Only Register Literal Types if compiling std lib
	if t.mode != STD {
		return
	}

	t.define("IntegerLiteral", GenerateBaseType("IntegerLiteral"))
	t.define("FloatLiteral", GenerateBaseType("FloatLiteral"))
	t.define("StringLiteral", GenerateBaseType("StringLiteral"))
	t.define("BooleanLiteral", GenerateBaseType("BooleanLiteral"))
	t.define("AnyLiteral", GenerateBaseType("AnyLiteral"))
	t.define("NullLiteral", GenerateBaseType("NullLiteral"))
	t.define("VoidLiteral", GenerateBaseType("VoidLiteral"))
	t.define("ArrayLiteral", GenerateGenericType("ArrayLiteral", GenerateBaseType("AnyLiteral")))
	t.define("MapLiteral", GenerateGenericType("MapLiteral", GenerateBaseType("AnyLiteral"), GenerateBaseType("AnyLiteral")))
}

func (t *TypeChecker) error(message string, node ast.Node) lexer.Error {
	return lexer.Error{
		Range:   node.Range(),
		Message: message,
	}
}
