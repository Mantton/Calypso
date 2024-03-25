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

func TestFunctionOverloading(t *testing.T) {
	input := `
	module main;
	// standard
	fn make(_: string){}
	fn move(_ file: string){}
	fn populate(id: string){}


	// overloaded standard
	fn ping(station: int) -> int {
		return 10;
	}

	fn ping(station: int, frequency: double) -> int {
		return 10;
	}

	// overloaded with generic option
	fn update(name: string){}
	fn update<T>(name: T) {}


	fn main() {

		// standard
		make("Hotdogs");
		move("file.txt");
		populate(id: "Valora");

		// standard overloaded
		const A = ping(station: 100);
		const B = ping(station: 100, frequency: 20);

		// overloaded generic
		update(name: "Jack");
	}
`

	Build(input, t, true)
}

func TestMethodOverloading(t *testing.T) {
	input := `
		module main;

		struct User {
			name : string;
		}


		extension User {
			fn get_first() -> string {
				return self.name;
			}

			fn get_first(name: string) -> string {
				return name;
			}

			fn get_last() -> string {
				return "foo";
			}
		}


		fn main() {
			const user = User { name : "nope" };

			const A = user.get_first();
			const B = user.get_first(name: "hello");
			const C = user.get_last();
		}
	`

	Build(input, t, true)
}
func TestStandardAndConformanceDeclaration(t *testing.T) {
	input := `

		module main;

		standard Foo {
			fn Bar() -> int;
		}

		conform int to Foo {
			fn Bar() -> int {
				return 10;
			}
		}

		struct Baz<T: Foo> {
			Value : T;
		}


		fn main() {
			const A: Baz<int> = Baz { Value: 1 };
			const C = A.Value.Bar();
		}
	`

	Build(input, t, true)

}

func TestExtensionDeclaration(t *testing.T) {
	input := `
		module main;

		struct Foo {
			Value: string;
		}

		extension string {
			fn first() -> char {
				return 'A';
			}
		}


		fn main() {
			const A = Foo { Value: "hello, world" };
			const B = A.Value.first();
		}
	`

	Build(input, t, true)
}

func TestExternDeclaration(t *testing.T) {
	input := `
		module main;
		
		extern "libc" {
			fn malloc(size: u64) -> *u8;
		}

		fn main() {
			const ptr = malloc(20);
		}
	`

	Build(input, t, true)
}
