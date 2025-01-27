package ast

type Node interface {
	Visit(v Visitor)
}

type Visitor interface {
	PreVisitBlock(bl *Block) bool
	PostVisitBlock(bl *Block)

	PreVisitVarDecl(vd *VarDecl) bool
	PostVisitVarDecl(vd *VarDecl)

	PreVisitConstantStmt(stmt *ConstantStmt) bool
	PostVisitConstantStmt(stmt *ConstantStmt)

	PreVisitDeclareStmt(stmt *DeclareStmt) bool
	PostVisitDeclareStmt(stmt *DeclareStmt)

	PreVisitDisplayStmt(d *DisplayStmt) bool
	PostVisitDisplayStmt(d *DisplayStmt)

	PreVisitInputStmt(i *InputStmt) bool
	PostVisitInputStmt(i *InputStmt)

	PreVisitSetStmt(s *SetStmt) bool
	PostVisitSetStmt(s *SetStmt)

	PreVisitIntegerLiteral(il *IntegerLiteral) bool
	PostVisitIntegerLiteral(l *IntegerLiteral)

	PreVisitRealLiteral(rl *RealLiteral) bool
	PostVisitRealLiteral(l *RealLiteral)

	PreVisitStringLiteral(sl *StringLiteral) bool
	PostVisitStringLiteral(sl *StringLiteral)

	PreVisitCharacterLiteral(cl *CharacterLiteral) bool
	PostVisitCharacterLiteral(cl *CharacterLiteral)

	PreVisitBinaryOperation(bo *BinaryOperation) bool
	PostVisitBinaryOperation(bo *BinaryOperation)

	PreVisitVariableExpression(ve *VariableExpression) bool
	PostVisitVariableExpression(ve *VariableExpression)
}
