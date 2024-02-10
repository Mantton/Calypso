package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/symbols"
)

func (c *Checker) checkStatement(stmt ast.Statement) {
	fmt.Printf(
		"\nChecking Statement: %T @ Line %d\n",
		stmt,
		stmt.Range().Start.Line,
	)
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
	case *ast.StructStatement:
		c.checkStructStatement(stmt)
	default:
		msg := fmt.Sprintf("statement check not implemented, %T", stmt)
		panic(msg)
	}
}

func (c *Checker) checkVariableStatement(stmt *ast.VariableStatement) {

	s := symbols.NewSymbol(stmt.Identifier.Value, symbols.VariableSymbol)
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

	var annotation *symbols.SymbolInfo

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
	err := c.validate(annotation, initializer, annotation.Specializations)

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

	s := symbols.NewSymbol(stmt.Identifier.Value, symbols.AliasSymbol)
	prev := c.currentSym
	c.currentSym = s
	defer func() {
		c.currentSym = prev
	}()

	ok := c.define(s)

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is already defined", stmt.Identifier.Value),
			stmt.Identifier.Range(),
		)
		return
	}

	// alias Foo<T: Hashing> = Array<T>
	genericParams := stmt.GenericParams

	if genericParams != nil {
		c.evaluateGenericParameters(s, genericParams)
	}

	expected := c.evaluateTypeExpression(stmt.Target)
	s.AliasOf = expected
}

func (c *Checker) evaluateGenericParameters(s *symbols.SymbolInfo, clause *ast.GenericParametersClause) bool {
	hasErrors := false
	for _, param := range clause.Parameters {
		p := symbols.NewSymbol(param.Identifier.Value, symbols.GenericTypeSymbol)

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
			if standardSym.Type != symbols.StandardSymbol {
				c.addError(
					fmt.Sprintf("`%s` is not a conformable standard", standard.Value),
					standard.Range(),
				)
				hasErrors = true
			}

			// Add Standard to Parameter
			err := p.AddConstraint(standardSym)

			if err != nil {
				c.addError(
					err.Error(),
					param.Identifier.Range(),
				)
				hasErrors = true

			}

		}
		// Add Parameters to Symbol
		err := s.AddGenericParameter(p)

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
		err := c.validate(expected, provided, nil)

		if err != nil {
			c.addError(err.Error(), stmt.Range())
		}

		return
	}

	// No Expected Return Type, Try Inference
	inferred := c.currentSym.FuncDesc.InferredReturnType

	if inferred != nil {
		err := c.validate(inferred, provided, nil)
		// Types Do not match, set to any
		if err != nil {
			c.currentSym.FuncDesc.InferredReturnType = c.resolveLiteral(symbols.ANY)
			return
		}
		// Types Match, Do nothing
	} else {
		c.currentSym.FuncDesc.InferredReturnType = provided
	}

}

func (c *Checker) checkStructStatement(stmt *ast.StructStatement) {

	sym := symbols.NewSymbol(stmt.Identifier.Value, symbols.StructSymbol)
	// Enter new Symbol Scope
	prev := c.currentSym
	c.currentSym = sym
	defer func() {
		c.currentSym = prev
	}()

	// Define
	ok := c.define(sym)

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is already defined", sym.Name),
			stmt.Identifier.Range(),
		)
		return
	}

	// Register Generic Parameters
	if stmt.GenericParams != nil {
		c.evaluateGenericParameters(sym, stmt.GenericParams)
	}

	// Register Properties
	for _, prop := range stmt.Properties {
		pSym := symbols.NewSymbol(prop.Value, symbols.VariableSymbol)
		t := c.evaluateTypeExpression(prop.AnnotatedType)
		pSym.TypeDesc = t
		sym.AddProperty(pSym)
	}
}
