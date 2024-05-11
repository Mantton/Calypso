package ast

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/token"
)

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
		Start: e.Target.Range().Start,
		End:   e.Body.RBracePos,
	}
}

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

func (e *CallArgument) Range() token.SyntaxRange {
	if e.Label == nil {
		return e.Value.Range()
	}

	return token.SyntaxRange{
		Start: e.Label.Pos,
		End:   e.Value.Range().End,
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
		End:   e.Value.Range().End,
	}
}

func (e *ShorthandAssignmentExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Target.Range().Start,
		End:   e.Right.Range().End,
	}
}

func (e *IndexExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Target.Range().Start,
		End:   e.RBracketPos,
	}
}

func (e *FieldAccessExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Target.Range().Start,
		End:   e.Field.Range().End,
	}
}

func (e *KeyValueExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Key.Range().Start,
		End:   e.Value.Range().End,
	}
}

func (e *CompositeLiteralField) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Key.Range().Start,
		End:   e.Value.Range().End,
	}
}
func (e *GenericParametersClause) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.LChevronPos,
		End:   e.RChevronPos,
	}
}

func (e *SpecializationExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Expression.Range().Start,
		End:   e.Clause.Range().End,
	}
}

func (e *SwitchCaseExpression) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   e.Action.RBrackPos,
	}
}
func (e *FunctionParameter) Range() token.SyntaxRange {
	if e.Label == nil {
		return token.SyntaxRange{
			Start: e.Name.Pos,
			End:   e.Type.Range().End,
		}
	}

	return token.SyntaxRange{
		Start: e.Label.Pos,
		End:   e.Type.Range().End,
	}
}

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

func (e *StructStatement) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   e.RBracePos,
	}
}

func (e *EnumStatement) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   e.RBracePos,
	}
}
func (e *SwitchStatement) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   e.RBracePos,
	}
}

func (e *BreakStatement) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   e.KeyWPos,
	}
}

func (s *IfStatement) statementNode()                    {}
func (s *ExpressionStatement) statementNode()            {}
func (s *WhileStatement) statementNode()                 {}
func (s *ReturnStatement) statementNode()                {}
func (s *BlockStatement) statementNode()                 {}
func (s *VariableStatement) statementNode()              {}
func (s *FunctionStatement) statementNode()              {}
func (s *StructStatement) statementNode()                {}
func (s *EnumStatement) statementNode()                  {}
func (s *SwitchStatement) statementNode()                {}
func (s *BreakStatement) statementNode()                 {}
func (d *TypeStatement) statementNode()                  {}
func (d *DereferenceAssignmentStatement) statementNode() {}

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
func (e *TypeStatement) Range() token.SyntaxRange {

	if e.Value != nil {
		return token.SyntaxRange{
			Start: e.KeyWPos,
			End:   e.Value.Range().End,
		}
	} else {
		return token.SyntaxRange{
			Start: e.KeyWPos,
			End:   e.Identifier.Range().End,
		}
	}

}

func (e *DereferenceAssignmentStatement) Range() token.SyntaxRange {
	return token.SyntaxRange{
		Start: e.Target.Range().Start,
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

func (e *ImportDeclaration) Range() token.SyntaxRange {

	if e.Alias != nil {
		return token.SyntaxRange{
			Start: e.KeyWPos,
			End:   e.Alias.Pos,
		}
	}
	return token.SyntaxRange{
		Start: e.KeyWPos,
		End:   e.Path.Pos,
	}
}

func (d *ConstantDeclaration) declarationNode()    {}
func (d *StatementDeclaration) declarationNode()   {}
func (d *FunctionDeclaration) declarationNode()    {}
func (d *StandardDeclaration) declarationNode()    {}
func (d *ExtensionDeclaration) declarationNode()   {}
func (d *ConformanceDeclaration) declarationNode() {}
func (d *ExternDeclaration) declarationNode()      {}
func (d *ImportDeclaration) declarationNode()      {}

// * Types

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

func IsTypeNode(n Node) bool {
	switch n.(type) {
	case *IdentifierExpression,
		*SpecializationExpression,
		*ArrayTypeExpression,
		*MapTypeExpression,
		*PointerTypeExpression,
		*FieldAccessExpression:
		return true
	}

	return false
}

func (e *CallArgument) GetLabel() string {
	if e.Label != nil {
		return e.Label.Value
	} else {
		return ""
	}
}

func (n *IntegerLiteral) String() string {
	return fmt.Sprintf("%d", n.Value)
}
func (n *FloatLiteral) String() string {
	return fmt.Sprintf("%f", n.Value)
}
func (n *StringLiteral) String() string {
	return fmt.Sprintf("\"%s\"", n.Value)
}
func (n *CharLiteral) String() string {
	return fmt.Sprintf("'%c'", n.Value)
}
func (n *BooleanLiteral) String() string {
	if n.Value {
		return "true"
	}

	return "false"
}
func (n *NilLiteral) String() string {
	return "nil"
}
func (n *VoidLiteral) String() string {
	return "void"
}
func (n *CompositeLiteral) String() string {
	return fmt.Sprintf("_composite::%s::{ %s }", n.Target, n.Body)
}
func (n *ArrayLiteral) String() string {
	return ""
}
func (n *MapLiteral) String() string {
	return ""
}
func (n *IdentifierExpression) String() string {
	return "_ident" + "::" + n.Value
}
func (n *FunctionExpression) String() string {
	return ""
}
func (n *GroupedExpression) String() string {
	return ""
}
func (n *CallExpression) String() string {
	return ""
}
func (n *CallArgument) String() string {
	return ""
}
func (n *UnaryExpression) String() string {
	return ""
}
func (n *BinaryExpression) String() string {
	return ""
}
func (n *AssignmentExpression) String() string {
	return ""
}
func (n *ShorthandAssignmentExpression) String() string {
	return ""
}
func (n *IndexExpression) String() string {
	return ""
}
func (n *FieldAccessExpression) String() string {
	return fmt.Sprintf("%s:accessing:%s", n.Target, n.Field)
}
func (n *KeyValueExpression) String() string {
	return ""
}
func (n *CompositeLiteralField) String() string {
	return ""
}
func (n *GenericParametersClause) String() string {
	o := "_TParams::"
	for _, p := range n.Parameters {
		o += p.String()
	}
	return o
}
func (n *SpecializationExpression) String() string {
	return fmt.Sprintf("_concrete::%s::<%s>", n.Expression, n.Clause)
}
func (n *SwitchCaseExpression) String() string {
	return ""
}
func (n *FunctionParameter) String() string {
	return ""
}
func (n *IfStatement) String() string {
	return ""
}
func (n *BlockStatement) String() string {
	return ""
}
func (n *ExpressionStatement) String() string {
	return ""
}
func (n *WhileStatement) String() string {
	return ""
}
func (n *ReturnStatement) String() string {
	return ""
}
func (n *VariableStatement) String() string {
	return ""
}
func (n *FunctionStatement) String() string {
	return ""
}
func (n *StructStatement) String() string {
	return ""
}
func (n *EnumStatement) String() string {
	return ""
}
func (n *SwitchStatement) String() string {
	return ""
}
func (n *BreakStatement) String() string {
	return ""
}
func (n *ConstantDeclaration) String() string {
	return ""
}
func (n *FunctionDeclaration) String() string {
	return ""
}
func (n *StatementDeclaration) String() string {
	return ""
}
func (n *StandardDeclaration) String() string {
	return ""
}
func (n *TypeStatement) String() string {
	return ""
}

func (n *DereferenceAssignmentStatement) String() string {
	return ""
}

func (n *ExtensionDeclaration) String() string {
	return ""
}
func (n *ConformanceDeclaration) String() string {
	return ""
}
func (n *ExternDeclaration) String() string {
	return ""
}
func (n *ImportDeclaration) String() string {
	return ""
}
func (n *ArrayTypeExpression) String() string {
	return ""
}
func (n *MapTypeExpression) String() string {
	return ""
}
func (n *GenericArgumentsClause) String() string {
	o := "_TArgs::"
	for _, p := range n.Arguments {
		o += p.String() + ","
	}
	return o
}
func (n *PointerTypeExpression) String() string {
	return ""
}

func (n *CompositeLiteralBody) String() string {
	return ""
}
func (n *GenericParameterExpression) String() string {
	return n.Identifier.String()
}
