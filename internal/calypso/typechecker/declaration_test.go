package typechecker

import (
	"testing"

	"github.com/mantton/calypso/internal/calypso/types"
)

func TestConstantDeclaration(t *testing.T) {
	input := `
		module main;
		const Foo = 20;
		const Bar: string = "hello, world";
		const Baz = true;
	`

	res, err := CheckString(input)

	if err != nil {
		t.Fatal(err)
	}

	fooSym := res.Scope.MustResolve("Foo")

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

	_, err := CheckString(input)

	if err != nil {
		t.Fatal(err)
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

	res, err := CheckString(input)

	if err != nil {
		t.Fatal(err)
	}

	foo := res.Scope.MustResolve("foo")

	if foo == nil {
		t.Error("function \"foo\" is not in scope")
	}

	intT := types.LookUp(types.Int)
	sg := types.NewFunctionSignature()
	sg.Result.SetType(intT)
	_, err = types.Validate(foo.Type(), sg)

	if err != nil {
		t.Error(err)
	}

	main := res.Scope.MustResolve("main")
	if main == nil {
		t.Error("function \"main\" not in scope")
	}

	fn := types.AsFunction(main)

	if fn == nil {
		t.Error("\"main\" is not a function")
	}

	vars := []string{"A", "B", "C", "D"}

	for _, v := range vars {
		A := fn.Scope.MustResolve(v)

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

	_, err := CheckString(input)

	if err != nil {
		t.Fatal(err)
	}
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

	_, err := CheckString(input)

	if err != nil {
		t.Fatal(err)
	}
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

	_, err := CheckString(input)

	if err != nil {
		t.Fatal(err)
	}
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

	_, err := CheckString(input)

	if err != nil {
		t.Fatal(err)
	}
}

func TestExternDeclaration(t *testing.T) {
	input := `
		module main;
		
		extern "libc" {
			fn malloc(size: u64) -> *u8;
		}

		fn main() {
			const ptr = malloc(size: 20);
		}
	`

	_, err := CheckString(input)

	if err != nil {
		t.Fatal(err)
	}
}
