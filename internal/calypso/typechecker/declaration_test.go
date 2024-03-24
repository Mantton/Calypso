package typechecker

import (
	"testing"

	"github.com/mantton/calypso/internal/calypso/parser"
	"github.com/mantton/calypso/internal/calypso/types"
)

func TestValidConstant(t *testing.T) {
	input := `
		module main;
		const Foo = 20;
		const Bar: string = "hello, world";
		const Baz = true;
	`

	file, errs := parser.ParseString(input)

	if len(errs) != 0 {
		t.Fatal(errs)
	}

	c := New(USER, file)
	res := c.Check()

	if len(c.Errors) != 0 {
		t.Error("expected no errors")
	}

	fooSym := res.Main.MustResolve("Foo")

	if fooSym == nil {
		t.Error("Expected Foo Symbol")
	} else if fooSym.Type() != types.LookUp(types.Int) {
		t.Errorf("Expected Integer Type got %s", fooSym.Type())
	}
}
