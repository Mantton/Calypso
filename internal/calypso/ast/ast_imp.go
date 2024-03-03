package ast

import "github.com/mantton/calypso/internal/calypso/token"

// * Base Literals

func (e *IntegerLiteral) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Pos,
		End:   e.Pos,
	}
}

func (e *FloatLiteral) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Pos,
		End:   e.Pos,
	}
}

func (e *StringLiteral) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Pos,
		End:   e.Pos,
	}
}

func (e *CharLiteral) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Pos,
		End:   e.Pos,
	}
}

func (e *BooleanLiteral) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Pos,
		End:   e.Pos,
	}
}
func (e *NilLiteral) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Pos,
		End:   e.Pos,
	}
}
func (e *VoidLiteral) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Pos,
		End:   e.Pos,
	}
}

func (e *CompositeLiteral) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Identifier.Pos,
		End:   e.RBracePos,
	}
}

func (e *IntegerLiteral) expressionNode()   {}
func (e *FloatLiteral) expressionNode()     {}
func (e *StringLiteral) expressionNode()    {}
func (e *CharLiteral) expressionNode()      {}
func (e *BooleanLiteral) expressionNode()   {}
func (e *NilLiteral) expressionNode()       {}
func (e *VoidLiteral) expressionNode()      {}
func (e *CompositeLiteral) expressionNode() {}

// * Generic Literals
func (e *ArrayLiteral) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.LBracketPos,
		End:   e.RBracketPos,
	}
}

func (e *MapLiteral) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.LBracePos,
		End:   e.RBracePos,
	}
}
func (e *ArrayLiteral) expressionNode() {}
func (e *MapLiteral) expressionNode()   {}

// * Literal Expressions
func (e *IdentifierExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Pos,
		End:   e.Pos,
	}
}

func (e *FunctionExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   e.RParenPos,
	}
}
func (e *IdentifierExpression) expressionNode() {}
func (e *FunctionExpression) expressionNode()   {}

// Expressions

func (e *GroupedExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.LParenPos,
		End:   e.RParenPos,
	}
}

func (e *CallExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Target.Range().Start,
		End:   e.RParenPos,
	}
}

func (e *UnaryExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.OpPosition,
		End:   e.Expr.Range().End,
	}
}

func (e *BinaryExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Left.Range().Start,
		End:   e.Right.Range().End,
	}
}

func (e *AssignmentExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Target.Range().Start,
		End:   e.Target.Range().End,
	}
}

func (e *IndexExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Target.Range().Start,
		End:   e.RBracketPos,
	}
}

func (e *PropertyExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Target.Range().Start,
		End:   e.Property.Range().End,
	}
}

func (e *KeyValueExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Key.Range().Start,
		End:   e.Value.Range().End,
	}
}

func (e *CompositeLiteralBodyClause) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Key.Range().Start,
		End:   e.Value.Range().End,
	}
}

func (e *GroupedExpression) expressionNode()          {}
func (e *CallExpression) expressionNode()             {}
func (e *UnaryExpression) expressionNode()            {}
func (e *BinaryExpression) expressionNode()           {}
func (e *AssignmentExpression) expressionNode()       {}
func (e *IndexExpression) expressionNode()            {}
func (e *PropertyExpression) expressionNode()         {}
func (e *KeyValueExpression) expressionNode()         {}
func (e *CompositeLiteralBodyClause) expressionNode() {}

// * Statements
func (e *IfStatement) Range() token.SyntaxRange {
	end := e.Action.Range().End

	if e.Alternative != nil {
		end = e.Alternative.Range().End
	}

	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   end,
	}
}

func (e *BlockStatement) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.LBrackPos,
		End:   e.RBrackPos,
	}
}

func (e *ExpressionStatement) Range() token.SyntaxRange {
	return e.Expr.Range()
}

func (e *WhileStatement) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   e.Action.Range().End,
	}
}

func (e *ReturnStatement) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   e.Value.Range().End,
	}
}

func (e *VariableStatement) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   e.Value.Range().End,
	}
}

func (e *FunctionStatement) Range() token.SyntaxRange {
	return e.Func.Range()
}

func (e *AliasStatement) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   e.Target.Range().End,
	}
}

func (e *StructStatement) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   e.RBracePos,
	}
}

func (s *IfStatement) statementNode()         {}
func (s *ExpressionStatement) statementNode() {}
func (s *WhileStatement) statementNode()      {}
func (s *ReturnStatement) statementNode()     {}
func (s *BlockStatement) statementNode()      {}
func (s *VariableStatement) statementNode()   {}
func (s *FunctionStatement) statementNode()   {}
func (s *AliasStatement) statementNode()      {}
func (s *StructStatement) statementNode()     {}

// * Declarations
func (e *ConstantDeclaration) Range() token.SyntaxRange {
	return e.Stmt.Range()
}

func (e *FunctionDeclaration) Range() token.SyntaxRange {
	return e.Func.Range()
}

func (e *StatementDeclaration) Range() token.SyntaxRange {
	return e.Stmt.Range()
}

func (e *StandardDeclaration) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   e.Block.RBrackPos,
	}
}
func (e *TypeDeclaration) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   e.Value.Range().End,
	}
}

func (e *ExtensionDeclaration) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   e.RBracePos,
	}
}

func (e *ConformanceDeclaration) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   e.RBracePos,
	}
}

func (e *ExternDeclaration) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   e.RBracePos,
	}
}

func (d *ConstantDeclaration) declarationNode()    {}
func (d *StatementDeclaration) declarationNode()   {}
func (d *FunctionDeclaration) declarationNode()    {}
func (d *StandardDeclaration) declarationNode()    {}
func (d *TypeDeclaration) declarationNode()        {}
func (d *ExtensionDeclaration) declarationNode()   {}
func (d *ConformanceDeclaration) declarationNode() {}
func (d *ExternDeclaration) declarationNode()      {}

// * Types
func (e *IdentifierTypeExpression) Range() token.SyntaxRange {
	return e.Identifier.Range()
}

func (e *ArrayTypeExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.LBracketPos,
		End:   e.RBracketPos,
	}
}

func (e *MapTypeExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.LBracketPos,
		End:   e.RBracketPos,
	}
}

func (e *GenericArgumentsClause) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.LChevronPos,
		End:   e.RChevronPos,
	}
}

func (e *PointerTypeExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.StarPos,
		End:   e.PointerTo.Range().End,
	}
}

func (e *IdentifierTypeExpression) typeNode() {}
func (e *ArrayTypeExpression) typeNode()      {}
func (e *MapTypeExpression) typeNode()        {}
func (e *PointerTypeExpression) typeNode()    {}
