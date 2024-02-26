package ast

type visitor interface {
	// Literals
	VisitIntegerLiteral(*IntegerLiteral)
	VisitFloatLiteral(*FloatLiteral)
	VisitStringLiteral(*StringLiteral)
	VisitCharLiteral(*CharLiteral)
	VisitBooleanLiteral(*BooleanLiteral)
	VisitNullLiteral(*NullLiteral)
	VisitVoidLiteral(*VoidLiteral)

	// Complex Literals
	VisitCompositeLiteral(*CompositeLiteral)
	VisitArrayLiteral(*ArrayLiteral)
	VisitMapLiteral(*MapLiteral)

	// Base Expressions
	VisitFunctionExpression(*FunctionExpression)
	VisitIdentifierExpression(*IdentifierExpression)

	// Other Expressions
	VisitGroupedExpression(*GroupedExpression)
	VisitCallExpression(*CallExpression)
	VisitUnaryExpression(*UnaryExpression)
	VisitBinaryExpression(*BinaryExpression)
	VisitAssignmentExpression(*AssignmentExpression)
	VisitIndexExpression(*IndexExpression)
	VisitPropertyExpression(*PropertyExpression)
	VisitKeyValueExpression(*KeyValueExpression)
	VisitCompositeLiteralBodyClause(*CompositeLiteralBodyClause)

	// Statements
	VisitIfStatement(*IfStatement)
	VisitExpressionStatement(*ExpressionStatement)
	VisitWhileStatement(*WhileStatement)
	VisitReturnStatement(*ReturnStatement)
	VisitBlockStatement(*BlockStatement)
	VisitVariableStatement(*VariableStatement)
	VisitFunctionStatement(*FunctionStatement)
	VisitAliasStatement(*AliasStatement)
	VisitStructStatement(*StructStatement)

	// Declarations
	VisitConstantDeclaration(*ConstantDeclaration)
	VisitStatementDeclaration(*StatementDeclaration)
	VisitFunctionDeclaration(*FunctionDeclaration)
	VisitStandardDeclaration(*StandardDeclaration)
	VisitTypeDeclaration(*TypeDeclaration)
	VisitExtensionDeclaration(*ExtensionDeclaration)
	VisitConformanceDeclaration(*ConformanceDeclaration)

	// Types
	VisitIdentifierTypeExpression(*IdentifierTypeExpression)
	VisitArrayTypeExpression(*ArrayTypeExpression)
	VisitMapTypeExpression(*MapTypeExpression)
}

// Literal Conformance
func (n *IntegerLiteral) Accept(v visitor) {
	v.VisitIntegerLiteral(n)
}
func (n *FloatLiteral) Accept(v visitor) {
	v.VisitFloatLiteral(n)
}
func (n *StringLiteral) Accept(v visitor) {
	v.VisitStringLiteral(n)
}
func (n *CharLiteral) Accept(v visitor) {
	v.VisitCharLiteral(n)
}
func (n *BooleanLiteral) Accept(v visitor) {
	v.VisitBooleanLiteral(n)
}
func (n *NullLiteral) Accept(v visitor) {
	v.VisitNullLiteral(n)
}
func (n *VoidLiteral) Accept(v visitor) {
	v.VisitVoidLiteral(n)
}

// Complex Literals Conformance
func (n *CompositeLiteral) Accept(v visitor) {
	v.VisitCompositeLiteral(n)
}
func (n *ArrayLiteral) Accept(v visitor) {
	v.VisitArrayLiteral(n)
}
func (n *MapLiteral) Accept(v visitor) {
	v.VisitMapLiteral(n)
}

// Base Expressions
func (n *FunctionExpression) Accept(v visitor) {
	v.VisitFunctionExpression(n)
}
func (n *IdentifierExpression) Accept(v visitor) {
	v.VisitIdentifierExpression(n)
}

// Complex Expressions
func (n *GroupedExpression) Accept(v visitor) {
	v.VisitGroupedExpression(n)
}
func (n *CallExpression) Accept(v visitor) {
	v.VisitCallExpression(n)
}
func (n *UnaryExpression) Accept(v visitor) {
	v.VisitUnaryExpression(n)
}
func (n *BinaryExpression) Accept(v visitor) {
	v.VisitBinaryExpression(n)
}
func (n *AssignmentExpression) Accept(v visitor) {
	v.VisitAssignmentExpression(n)
}
func (n *IndexExpression) Accept(v visitor) {
	v.VisitIndexExpression(n)
}
func (n *PropertyExpression) Accept(v visitor) {
	v.VisitPropertyExpression(n)
}
func (n *KeyValueExpression) Accept(v visitor) {
	v.VisitKeyValueExpression(n)
}
func (n *CompositeLiteralBodyClause) Accept(v visitor) {
	v.VisitCompositeLiteralBodyClause(n)
}

// Statements
func (n *IfStatement) Accept(v visitor) {
	v.VisitIfStatement(n)
}
func (n *ExpressionStatement) Accept(v visitor) {
	v.VisitExpressionStatement(n)
}
func (n *WhileStatement) Accept(v visitor) {
	v.VisitWhileStatement(n)
}
func (n *ReturnStatement) Accept(v visitor) {
	v.VisitReturnStatement(n)
}
func (n *BlockStatement) Accept(v visitor) {
	v.VisitBlockStatement(n)
}
func (n *VariableStatement) Accept(v visitor) {
	v.VisitVariableStatement(n)
}
func (n *FunctionStatement) Accept(v visitor) {
	v.VisitFunctionStatement(n)
}
func (n *AliasStatement) Accept(v visitor) {
	v.VisitAliasStatement(n)
}
func (n *StructStatement) Accept(v visitor) {
	v.VisitStructStatement(n)
}

// Declarations
func (n *ConstantDeclaration) Accept(v visitor) {
	v.VisitConstantDeclaration(n)
}
func (n *StatementDeclaration) Accept(v visitor) {
	v.VisitStatementDeclaration(n)
}
func (n *FunctionDeclaration) Accept(v visitor) {
	v.VisitFunctionDeclaration(n)
}
func (n *StandardDeclaration) Accept(v visitor) {
	v.VisitStandardDeclaration(n)
}
func (n *TypeDeclaration) Accept(v visitor) {
	v.VisitTypeDeclaration(n)
}
func (n *ExtensionDeclaration) Accept(v visitor) {
	v.VisitExtensionDeclaration(n)
}
func (n *ConformanceDeclaration) Accept(v visitor) {
	v.VisitConformanceDeclaration(n)
}

// Types
func (n *IdentifierTypeExpression) Accept(v visitor) {
	v.VisitIdentifierTypeExpression(n)
}
func (n *ArrayTypeExpression) Accept(v visitor) {
	v.VisitArrayTypeExpression(n)
}
func (n *MapTypeExpression) Accept(v visitor) {
	v.VisitMapTypeExpression(n)
}
