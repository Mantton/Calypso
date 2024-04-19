package parser

import (
	"testing"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lexer"
)

func scan(input string) *Parser {
	F := lexer.NewFileFromString(input)
	return New(F)

}
func TestTypeAnnotation(t *testing.T) {

	input := `
		const A : int = 20;
		const B : array<int> = 20;
		const C : foo.int = 20;
		const D : foo.array<int> = 20;
		const F : dict<Foo, Bar> = 20;
		const E = foo { A: 20 };
		const F : int[] = [1];
		const G : {string : int} = [];
	`

	p := scan(input)
	_, err := p.TestParse()

	if err != nil {
		t.Error(err)
	}
}

func TestStructInitialization(t *testing.T) {
	input := `
	// const A = Foo { A : B };
	// const B = Foo <string> { A : B };
	// const C = Foo.Bar { A : B };
	// const D = Foo.Bar <string>{ A : B };

	fn main() {
		if (!D) {}
	}
	`
	p := scan(input)
	f, err := p.TestParse()

	if err != nil {
		t.Fatal(err)
	}

	for _, c := range f.Nodes.Constants {

		_, ok := c.Stmt.Value.(*ast.CompositeLiteral)
		if !ok {
			t.Errorf("[%s] Expected composite literal got %T, %s", c.Stmt.Identifier.Value, c.Stmt.Value, c.Stmt.Value)
		}
	}
}
