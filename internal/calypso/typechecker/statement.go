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
	case *ast.ReturnStatement:
		c.checkReturnStatement(stmt)
	case *ast.ExpressionStatement:
		c.checkExpression(stmt.Expr)
	default:
		msg := fmt.Sprintf("statement check not implemented, %T", stmt)
		panic(msg)
	}
}

func (c *Checker) checkVariableStatement(stmt *ast.VariableStatement) {

	s := newSymbolInfo(stmt.Identifier.Value, VariableSymbol)
	s.TypeDesc = unresolved
	s.Mutable = !stmt.IsConstant
	ok := c.define(s)

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is already defined", s.Name),
			stmt.Identifier.Range(),
		)
		return
	}

	var annotation *SymbolInfo

	// Check Annotation
	if t := stmt.Identifier.AnnotatedType; t != nil {
		annotation = c.evaluateTypeExpression(t)
	}

	initializer := c.evaluateExpression(stmt.Value)

	if annotation == nil {
		s.TypeDesc = initializer
		return
	}

	// Annotation Present, Ensure Annotated Type Matches the provided Type
	err := c.validate(annotation, initializer)

	if err != nil {
		c.addError(
			err.Error(),
			stmt.Value.Range(),
		)

	}

	s.TypeDesc = annotation
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

	prev := c.currentSym
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
		c.evaluateGenericParameters(s, genericParams)
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

	c.currentSym = prev
}

func (c *Checker) evaluateGenericParameters(s *SymbolInfo, clause *ast.GenericParametersClause) bool {
	hasErrors := false
	for _, param := range clause.Parameters {
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

	return !hasErrors
}

func (c *Checker) checkReturnStatement(stmt *ast.ReturnStatement) {

	if c.currentSym == nil {
		c.addError(
			"top level returns are not permitted",
			stmt.Range(),
		)

		return
	}

	if c.currentSym.FuncDesc == nil {
		panic("Current Symbol Scope is not a function")
	}

	provided := c.evaluateExpression(stmt.Value)
	expected := c.currentSym.FuncDesc.AnnotatedReturnType

	// Has Expected Return Type
	if expected != nil {
		err := c.validate(expected, provided)

		if err != nil {
			c.addError(err.Error(), stmt.Range())
		}

		return
	}

	// No Expected Return Type, Try Inference
	inferred := c.currentSym.FuncDesc.InferredReturnType

	if inferred != nil {
		err := c.validate(inferred, provided)
		// Types Do not match, set to any
		if err != nil {
			c.currentSym.FuncDesc.InferredReturnType = c.resolveLiteral(ANY)
			return
		}
		// Types Match, Do nothing
	} else {
		c.currentSym.FuncDesc.InferredReturnType = provided
	}

}
