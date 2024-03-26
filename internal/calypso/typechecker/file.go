package typechecker

import (
	"errors"
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) Check() (*SymbolTable, error) {

	main := types.NewScope(types.GlobalScope, "__cly__main__scope")
	main.Parent = types.GlobalScope
	c.table.Main = main
	c.ctx = NewContext(main, nil, nil)

	// Run Passes
	c.pass()

	if DEBUG {
		c.table.DebugPrintScopes()
	}

	if len(c.Errors) != 0 {
		return nil, errors.New(c.Errors.String())
	}

	return c.table, nil
}

/*
The Pass function runs multiple TC passes in the following order

Passes 0 - 4 - Register Signatures, Constants, & Type Declarations
Standard,
Types | Structs | Enums,
Conformance
Extensions
Constants,
External Functions  Signatures,
Local Functions Signatures,

Passes 5-8
*/
func (c *Checker) pass() {
	passes := []func(*ast.File){c.pass0, c.pass1, c.pass2, c.pass3, c.pass4}

	for _, pass := range passes {
		for _, file := range c.fileSet.Files {
			c.file = file
			pass(file)
		}
	}
}

// Collects Types & Standards
func (c *Checker) pass0(f *ast.File) {

	// Standards
	for _, d := range f.Nodes.Standards {
		symbol := c.define(d.Identifier, d, c.ParentScope())
		underlying := types.NewStandard(symbol.Name())
		symbol.SetType(underlying)

		ctx := NewContext(symbol.GetScope(), nil, nil)
		for _, t := range d.Block.Statements {
			switch t := t.(type) {
			case *ast.TypeStatement:
				alias := c.defineAlias(t, ctx)

				if alias != nil {
					underlying.AddType(alias)
				}

			case *ast.FunctionStatement:
				c.registerFunctionExpression(t.Func, ctx.scope)
			}
		}
	}

	// Types
	for _, d := range f.Nodes.Types {
		c.defineAlias(d, c.ctx)
	}
}

// Pass 1 - Collect Composite Types
func (c *Checker) pass1(f *ast.File) {
	// Enums
	for _, d := range f.Nodes.Enums {
		c.define(d.Identifier, d, c.ParentScope())
	}

	// Structs
	for _, d := range f.Nodes.Structs {
		c.define(d.Identifier, d, c.ParentScope())
	}
}

// Pass 2 - Type-Standard Conformance
func (c *Checker) pass2(f *ast.File) {
	// Conformance Types & Functions
	for _, d := range f.Nodes.Conformances {
		c.registerConformance(d)
	}

	// Extensions
	for _, d := range f.Nodes.Extensions {
		// Get Type
		ident := d.Identifier.Value
		symbol := c.ParentScope().MustResolve(ident)

		if symbol == nil {
			c.addError(fmt.Sprintf("cannot locate %s", ident), d.Identifier.Range())
			continue
		}

		definition := types.AsDefined(symbol.Type())
		for _, fn := range d.Content {
			c.registerFunctionExpression(fn.Func, definition.GetScope())
		}
	}
}

// Pass 3 - All Types should be registered now, register variables
func (c *Checker) pass3(f *ast.File) {
	// define constants as unresolved
	for _, d := range f.Nodes.Constants {
		stmt := d.Stmt
		def := types.NewVar(stmt.Identifier.Value, unresolved)
		def.Mutable = !stmt.IsConstant
		err := c.ParentScope().Define(def)

		if err != nil {
			c.addError(
				fmt.Sprintf(err.Error(), def.Name()),
				stmt.Identifier.Range(),
			)
		}
	}
}

// Pass 4 - Functions With Placeholders
func (c *Checker) pass4(f *ast.File) {

	// External Functions
	for _, d := range f.Nodes.ExternalFunctions {
		for _, fn := range d.Signatures {
			c.registerFunctionExpression(fn.Func, c.ParentScope())
		}
	}
	// Functions
	for _, d := range f.Nodes.Functions {
		c.registerFunctionExpression(d.Func, c.ParentScope())
	}
}

// At this point, all required entries are in scope, their RHS Values & bodies are left to color
func (c *Checker) pass5(f *ast.File) {

}
