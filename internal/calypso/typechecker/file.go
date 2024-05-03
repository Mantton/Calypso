package typechecker

import (
	"errors"
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) Check() (*types.Module, error) {

	main := types.NewScope(types.GlobalScope, "MAIN_SCOPE")
	main.Parent = types.GlobalScope
	c.module.Scope = main
	c.ctx = NewContext(main, nil, nil)

	// Run Passes
	c.pass()

	if len(c.Errors) != 0 {
		return nil, errors.New(c.Errors.String())
	}

	return c.module, nil
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
	passes := []func(*ast.File){
		// Set 0
		c.passN,
		// Set 1 - White
		c.pass0, c.pass1, c.pass2, c.pass3, c.pass4,
		// Set 2
		c.pass5, c.pass6, c.pass7, c.pass8,
	}

	for _, pass := range passes {
		for _, file := range c.module.AST.Set.Files {
			c.file = file
			pass(file)
		}
	}
}

func (c *Checker) passN(f *ast.File) {

	for _, d := range f.Nodes.Imports {
		key := d.PopulatedImportKey

		mod := c.mp.Modules[key]

		// trying to import same module being checked
		if mod == c.module {
			msg := fmt.Sprintf("importing current module, %s", mod)
			c.addError(msg, d.Range())
			continue
		} else if !mod.IsVisible(c.module) && mod.ParentModule != c.module { // trying to import a private module from outside it's source module
			msg := fmt.Sprintf("cannot import private module, %s", mod)
			c.addError(msg, d.Range())
			continue
		}

		// Define in scope
		name := mod.Name()
		if d.Alias != nil {
			name = d.Alias.Value
		}

		err := c.module.Scope.CustomDefine(mod, name)
		if err != nil {
			c.addError(err.Error(), d.Range())
		}
	}
}

// Collects Types & Standards
func (c *Checker) pass0(f *ast.File) {

	// Types
	for _, d := range f.Nodes.Types {
		alias := c.defineAlias(d, c.ctx)
		if alias == nil {
			continue
		}
		if d.GenericParams != nil {
			for _, p := range d.GenericParams.Parameters {
				d := types.NewTypeParam(p.Identifier.Value, nil)
				alias.AddTypeParameter(d)
			}
		}
	}

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
}

// Pass 1 - Collect Composite Types
func (c *Checker) pass1(f *ast.File) {
	// Enums
	for _, d := range f.Nodes.Enums {
		def := c.define(d.Identifier, d, c.ParentScope())
		c.registerTypeParameters(d.GenericParams, def)

	}

	// Structs
	for _, d := range f.Nodes.Structs {
		def := c.define(d.Identifier, d, c.ParentScope())
		c.registerTypeParameters(d.GenericParams, def)
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
			def := c.registerFunctionExpression(fn.Func, c.ParentScope())
			def.Target = &types.FunctionTarget{Target: d.Target.Value}
		}
	}
	// Functions
	for _, d := range f.Nodes.Functions {
		c.registerFunctionExpression(d.Func, c.ParentScope())
	}
}

// At this point, all required entries are in scope, their RHS Values & bodies are left to color
// eval types, enums, structs & standards
func (c *Checker) pass5(f *ast.File) {

	// Types
	for _, d := range f.Nodes.Types {
		c.checkTypeStatement(d, c.ctx)
	}

	// Enums
	for _, d := range f.Nodes.Enums {
		c.checkEnumStatement(d, c.ctx)
	}

	// Structs
	for _, d := range f.Nodes.Structs {
		c.checkStructStatement(d, c.ctx)
	}

	// Variables
	for _, d := range f.Nodes.Constants {
		c.checkVariableStatement(d.Stmt, c.ctx, true)
	}

	// Standards
	for _, d := range f.Nodes.Standards {
		c.checkStandardDeclaration(d)
	}
}

// Functions Signatures
func (c *Checker) pass6(f *ast.File) {
	// Conformance Types & Functions
	for _, d := range f.Nodes.Conformances {
		c.checkConformanceDeclaration(d)
	}

	// Extensions
	for _, d := range f.Nodes.Extensions {
		c.checkExtensionDeclaration(d)
	}

	// External Functions
	for _, d := range f.Nodes.ExternalFunctions {
		for _, fn := range d.Signatures {
			sg := c.registerFunctionSignatures(fn.Func)

			if types.IsGeneric(sg) {
				c.addError(fmt.Sprintf("external function %s cannot be generic", fn.Func.Identifier.Value), fn.Range())
			}
		}
	}
	// Functions
	for _, d := range f.Nodes.Functions {
		c.registerFunctionSignatures(d.Func)
	}
}

// constraint checking
func (c *Checker) pass7(f *ast.File) {

}

// bodies
func (c *Checker) pass8(f *ast.File) {
	for _, d := range f.Nodes.Conformances {
		for _, e := range d.Signatures {
			c.evaluateFunctionExpression(e.Func)
		}
	}

	// Extensions
	for _, d := range f.Nodes.Extensions {
		for _, e := range d.Content {
			c.evaluateFunctionExpression(e.Func)
		}
	}
	// NOTE: External Functions are complete

	// Functions
	for _, d := range f.Nodes.Functions {
		c.evaluateFunctionExpression(d.Func)
	}
}
