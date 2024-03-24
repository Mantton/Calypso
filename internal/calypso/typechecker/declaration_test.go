package typechecker

import (
	"testing"

	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/parser"
	"github.com/mantton/calypso/internal/calypso/types"
)

func Build(input string, t *testing.T, fOnE bool) (*SymbolTable, lexer.ErrorList) {
	file, errs := parser.ParseString(input)

	if len(errs) != 0 {
		t.Fatal(errs)
	}

	c := New(USER, file)
	res := c.Check()

	if fOnE && len(c.Errors) != 0 {
		t.Fatalf("expected no errors, got \n\t%s", c.Errors)
	}

	return res, c.Errors
}

func TestConstantDeclaration(t *testing.T) {
	input := `
		module main;
		const Foo = 20;
		const Bar: string = "hello, world";
		const Baz = true;
	`

	res, _ := Build(input, t, true)

	fooSym := res.Main.MustResolve("Foo")

	if fooSym == nil {
		t.Error("Expected Foo Symbol")
	} else if fooSym.Type() != types.LookUp(types.Int) {
		t.Errorf("Expected Integer Type got %s", fooSym.Type())
	}
}

func TestInvalidConstantDeclaration(t *testing.T) {
	input := `
		module main;

		fn call() {}

		const Foo = call();
	`

	_, errs := Build(input, t, false)

	if len(errs) == 0 {
		t.Error("expected error: global constants must be known at compile time")
	}
}

func TestFunctionDeclaration(t *testing.T) {

	input := `
		module main;


		fn foo() -> int {
			return 10;
		}

		fn baz<T>(_ a : T) -> T {
			return a;
		}

		fn foobar<T>(_ a: T, _ b: T) -> T {
			return a;
		}

		fn fibonacci(_ n: int) -> int {
			if (n <= 1) {
				return n;
			} else {
				return fibonacci(n - 1) + fibonacci(n - 2);
			}
		}

		fn main() {
			const A = foo();
			const B = baz(A);
			const C = foobar(A, B);
			const D = fibonacci(35);
		}
	`

	table, _ := Build(input, t, true)

	foo := table.Main.MustResolve("foo")

	if foo == nil {
		t.Error("function \"foo\" is not in scope")
	}

	intT := types.LookUp(types.Int)
	sg := types.NewFunctionSignature()
	sg.Result.SetType(intT)
	_, err := types.Validate(foo.Type(), sg)

	if err != nil {
		t.Error(err)
	}

	main := table.Main.MustResolve("main")
	if main == nil {
		t.Error("function \"main\" not in scope")
	}

	fn := types.AsFunction(main)

	if fn == nil {
		t.Error("\"main\" is not a function")
	}

	vars := []string{"A", "B", "C", "D"}

	for _, v := range vars {
		A := fn.Sg().Scope.MustResolve(v)

		_, err = types.Validate(A.Type(), intT)

		if err != nil {
			t.Error(err)
		}
	}

}
