package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
)

func (c *Checker) checkStatement(stmt ast.Statement) {
	c.currentNode = stmt
	switch stmt := stmt.(type) {
	case *ast.VariableStatement:
		c.checkVariableStatement(stmt)
	case *ast.BlockStatement:
		c.checkBlockStatement(stmt)
	case *ast.AliasStatement:
		c.checkAliasStatement(stmt)
	default:
		msg := fmt.Sprintf("statement check not implemented, %T", stmt)
		panic(msg)
	}
}

func (c *Checker) checkVariableStatement(stmt *ast.VariableStatement) {

	var annotation *SymbolInfo

	// Check Annotation
	if t := stmt.Identifier.AnnotatedType; t != nil {
		annotation = c.evaluateTypeExpression(t)
	}

	initializer := c.evaluateExpression(stmt.Value)

	if annotation == nil {
		s := newSymbolInfo(stmt.Identifier.Value, VariableSymbol)
		s.TypeDesc = initializer
		ok := c.define(s)

		if !ok {
			c.addError(
				fmt.Sprintf("`%s` is already defined", s.Name),
				stmt.Identifier.Range(),
			)
		}
		return
	}

	// Annotation Present, Ensure Annotated Type Matches the provided Type
	ok := c.validate(annotation, initializer)

	if !ok {
		// c.addError(
		// 	fmt.Sprintf("cannot assign `%s` to `%s`", initializer.Name, annotation.Name),
		// 	stmt.Identifier.Range(),
		// )

		s := newSymbolInfo(stmt.Identifier.Value, VariableSymbol)
		s.TypeDesc = unresolved
		ok := c.define(s)

		if !ok {
			c.addError(
				fmt.Sprintf("`%s` is already defined", s.Name),
				stmt.Identifier.Range(),
			)
			return
		}
	}

	// c.declareUnresolved(stmt.Identifier.Value, VariableSymbol)

}

func (c *Checker) checkBlockStatement(blk *ast.BlockStatement) {

	if len(blk.Statements) == 0 {
		return
	}

	for _, stmt := range blk.Statements {
		c.checkStatement(stmt)
	}
}

func (c *Checker) checkAliasStatement(stmt *ast.AliasStatement) {

	s := newSymbolInfo(stmt.Identifier.Value, AliasSymbol)
	hasErrors := false

	c.currentSym = s

	ok := c.define(s)

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is already defined", stmt.Identifier.Value),
			stmt.Identifier.Range(),
		)
		hasErrors = true
	}

	// TODO: Check Generics
	// alias Foo<T: Hashing> = Array<T>
	// expectedArguments := len(expected.GenericArguments)

	genericParams := stmt.GenericParams

	if genericParams != nil {
		for _, param := range genericParams.Parameters {
			p := newSymbolInfo(param.Identifier.Value, GenericTypeSymbol)

			// Resolve Standards
			for _, standard := range param.Standards {
				standardSym, ok := c.find(standard.Value)
				// not defined
				if !ok {
					c.addError(
						fmt.Sprintf("`%s` is not found", standard.Value),
						standard.Range(),
					)
					hasErrors = true
				}

				// Ensure Sym is Standard
				if standardSym.Type != StandardSymbol {
					c.addError(
						fmt.Sprintf("`%s` is not a conformable standard", standard.Value),
						standard.Range(),
					)
					hasErrors = true
				}

				// Add Standard to Parameter
				err := p.addConstraint(standardSym)

				if err != nil {
					c.addError(
						err.Error(),
						param.Identifier.Range(),
					)
					hasErrors = true

				}

			}
			// Add Parameters to Symbol
			err := s.addGenericParameter(p)

			if err != nil {
				c.addError(
					err.Error(),
					param.Identifier.Range(),
				)
				hasErrors = true
			}
		}
	}

	expected := c.evaluateTypeExpression(stmt.Target)

	// Ensure LHS & RHS have the same number of arguments
	if len(expected.GenericArguments) != len(s.GenericParams) {
		c.addError(
			fmt.Sprintf("`%s` expected %d arguments, provided %d", expected.Name, len(expected.GenericArguments), len(s.GenericParams)),
			stmt.Range(),
		)
		hasErrors = true
	}

	if !hasErrors {
		s.AliasOf = expected
		s.convertGenericParamsToArguments()
		fmt.Println(s.Name, "is alias of", expected.Name)
	}

	c.currentSym = nil
}
