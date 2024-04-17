package compile

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/fs"
	"github.com/mantton/calypso/internal/calypso/parser"
)

type compiler struct {
	pkg *fs.Package
}

func newCompiler(p *fs.Package) *compiler {
	return &compiler{
		pkg: p,
	}
}

func CompilePackage(pkg *fs.Package) error {

	c := newCompiler(pkg)

	// Scan & Parse Package
	astPkg, err := c.parsePackage()

	if err != nil {
		return err
	}

	fmt.Println(astPkg)

	return nil
}

func (c *compiler) parsePackage() (*ast.Package, error) {
	mod := c.pkg.Source

	src, err := parser.ParseModule(mod)

	if err != nil {
		return nil, err
	}
	return &ast.Package{
		Source: src,
	}, nil
}
