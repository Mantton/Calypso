package ast

type Visitor interface {
	// Literals
	visitIntegerLiteral(*IntegerLiteral)
	visitFloatLiteral(*FloatLiteral)
	visitStringLiteral(*StringLiteral)
	visitBooleanLiteral(*BooleanLiteral)
	visitNullLiteral(*NullLiteral)
	visitVoidLiteral(*VoidLiteral)

	// Complex Literals
	visitCompositeLiteral(*CompositeLiteral)
	visitArrayLiteral(*ArrayLiteral)
	visitMapLiteral(*MapLiteral)

	// Base Expressions
	visitFunctionExpression(*FunctionExpression)
	visitIdentifierExpression(*IdentifierExpression)

	// Other Expressions
	visitGroupedExpression(*GroupedExpression)
	visitCallExpression(*CallExpression)
	visitUnaryExpression(*UnaryExpression)
	visitBinaryExpression(*BinaryExpression)
	visitAssignmentExpression(*AssignmentExpression)
	visitIndexExpression(*IndexExpression)
	visitPropertyExpression(*PropertyExpression)
	visitKeyValueExpression(*KeyValueExpression)
	visitCompositeLiteralBodyClause(*CompositeLiteralBodyClause)

	// Statements
	visitIfStatement(*IfStatement)
	visitExpressionStatement(*ExpressionStatement)
	visitWhileStatement(*WhileStatement)
	visitReturnStatement(*ReturnStatement)
	visitBlockStatement(*BlockStatement)
	visitVariableStatement(*VariableStatement)
	visitFunctionStatement(*FunctionStatement)
	visitAliasStatement(*AliasStatement)
	visitStructStatement(*StructStatement)

	// Declarations
	visitConstantDeclaration(*ConstantDeclaration)
	visitStatementDeclaration(*StatementDeclaration)
	visitFunctionDeclaration(*FunctionDeclaration)
	visitStandardDeclaration(*StandardDeclaration)
	visitTypeDeclaration(*TypeDeclaration)
	visitExtensionDeclaration(*ExtensionDeclaration)
	visitConformanceDeclaration(*ConformanceDeclaration)

	// Types
	visitIdentifierTypeExpression(*IdentifierTypeExpression)
	visitArrayTypeExpression(*ArrayTypeExpression)
	visitMapTypeExpression(*MapTypeExpression)
}

// Literal Conformance
func (n *IntegerLiteral) accept(v Visitor) {
	v.visitIntegerLiteral(n)
}
func (n *FloatLiteral) accept(v Visitor) {
	v.visitFloatLiteral(n)
}
func (n *StringLiteral) accept(v Visitor) {
	v.visitStringLiteral(n)
}
func (n *BooleanLiteral) accept(v Visitor) {
	v.visitBooleanLiteral(n)
}
func (n *NullLiteral) accept(v Visitor) {
	v.visitNullLiteral(n)
}
func (n *VoidLiteral) accept(v Visitor) {
	v.visitVoidLiteral(n)
}

// Complex Literals Conformance
func (n *CompositeLiteral) accept(v Visitor) {
	v.visitCompositeLiteral(n)
}
func (n *ArrayLiteral) accept(v Visitor) {
	v.visitArrayLiteral(n)
}
func (n *MapLiteral) accept(v Visitor) {
	v.visitMapLiteral(n)
}

// Base Expressions
func (n *FunctionExpression) accept(v Visitor) {
	v.visitFunctionExpression(n)
}
func (n *IdentifierExpression) accept(v Visitor) {
	v.visitIdentifierExpression(n)
}

// Complex Expressions
func (n *GroupedExpression) accept(v Visitor) {
	v.visitGroupedExpression(n)
}
func (n *CallExpression) accept(v Visitor) {
	v.visitCallExpression(n)
}
func (n *UnaryExpression) accept(v Visitor) {
	v.visitUnaryExpression(n)
}
func (n *BinaryExpression) accept(v Visitor) {
	v.visitBinaryExpression(n)
}
func (n *AssignmentExpression) accept(v Visitor) {
	v.visitAssignmentExpression(n)
}
func (n *IndexExpression) accept(v Visitor) {
	v.visitIndexExpression(n)
}
func (n *PropertyExpression) accept(v Visitor) {
	v.visitPropertyExpression(n)
}
func (n *KeyValueExpression) accept(v Visitor) {
	v.visitKeyValueExpression(n)
}
func (n *CompositeLiteralBodyClause) accept(v Visitor) {
	v.visitCompositeLiteralBodyClause(n)
}

// Statements
func (n *IfStatement) accept(v Visitor) {
	v.visitIfStatement(n)
}
func (n *ExpressionStatement) accept(v Visitor) {
	v.visitExpressionStatement(n)
}
func (n *WhileStatement) accept(v Visitor) {
	v.visitWhileStatement(n)
}
func (n *ReturnStatement) accept(v Visitor) {
	v.visitReturnStatement(n)
}
func (n *BlockStatement) accept(v Visitor) {
	v.visitBlockStatement(n)
}
func (n *VariableStatement) accept(v Visitor) {
	v.visitVariableStatement(n)
}
func (n *FunctionStatement) accept(v Visitor) {
	v.visitFunctionStatement(n)
}
func (n *AliasStatement) accept(v Visitor) {
	v.visitAliasStatement(n)
}
func (n *StructStatement) accept(v Visitor) {
	v.visitStructStatement(n)
}

// Declarations
func (n *ConstantDeclaration) accept(v Visitor) {
	v.visitConstantDeclaration(n)
}
func (n *StatementDeclaration) accept(v Visitor) {
	v.visitStatementDeclaration(n)
}
func (n *FunctionDeclaration) accept(v Visitor) {
	v.visitFunctionDeclaration(n)
}
func (n *StandardDeclaration) accept(v Visitor) {
	v.visitStandardDeclaration(n)
}
func (n *TypeDeclaration) accept(v Visitor) {
	v.visitTypeDeclaration(n)
}
func (n *ExtensionDeclaration) accept(v Visitor) {
	v.visitExtensionDeclaration(n)
}
func (n *ConformanceDeclaration) accept(v Visitor) {
	v.visitConformanceDeclaration(n)
}

// Types
func (n *IdentifierTypeExpression) accept(v Visitor) {
	v.visitIdentifierTypeExpression(n)
}
func (n *ArrayTypeExpression) accept(v Visitor) {
	v.visitArrayTypeExpression(n)
}
func (n *MapTypeExpression) accept(v Visitor) {
	v.visitMapTypeExpression(n)
}
