package ast

type Visitor interface {
	PreVisitBlock(bl *Block) bool
	PostVisitBlock(bl *Block)

	PreVisitVarDecl(vd *VarDecl) bool
	PostVisitVarDecl(vd *VarDecl)

	PreVisitConstantStmt(cs *ConstantStmt) bool
	PostVisitConstantStmt(cs *ConstantStmt)

	PreVisitDeclareStmt(ds *DeclareStmt) bool
	PostVisitDeclareStmt(ds *DeclareStmt)

	PreVisitDisplayStmt(ds *DisplayStmt) bool
	PostVisitDisplayStmt(ds *DisplayStmt)

	PreVisitInputStmt(is *InputStmt) bool
	PostVisitInputStmt(is *InputStmt)

	PreVisitSetStmt(ss *SetStmt) bool
	PostVisitSetStmt(ss *SetStmt)

	PreVisitIfStmt(is *IfStmt) bool
	PostVisitIfStmt(is *IfStmt)
	PreVisitCondBlock(cb *CondBlock) bool
	PostVisitCondBlock(cb *CondBlock)

	PreVisitSelectStmt(ss *SelectStmt) bool
	PostVisitSelectStmt(ss *SelectStmt)
	PreVisitCaseBlock(cb *CaseBlock) bool
	PostVisitCaseBlock(cb *CaseBlock)

	PreVisitDoStmt(ds *DoStmt) bool
	PostVisitDoStmt(ds *DoStmt)

	PreVisitWhileStmt(ws *WhileStmt) bool
	PostVisitWhileStmt(ws *WhileStmt)

	PreVisitForStmt(fs *ForStmt) bool
	PostVisitForStmt(fs *ForStmt)

	PreVisitIntegerLiteral(il *IntegerLiteral) bool
	PostVisitIntegerLiteral(il *IntegerLiteral)

	PreVisitRealLiteral(rl *RealLiteral) bool
	PostVisitRealLiteral(rl *RealLiteral)

	PreVisitStringLiteral(sl *StringLiteral) bool
	PostVisitStringLiteral(sl *StringLiteral)

	PreVisitCharacterLiteral(cl *CharacterLiteral) bool
	PostVisitCharacterLiteral(cl *CharacterLiteral)

	PreVisitBooleanLiteral(bl *BooleanLiteral) bool
	PostVisitBooleanLiteral(bl *BooleanLiteral)

	PreVisitUnaryOperation(uo *UnaryOperation) bool
	PostVisitUnaryOperation(uo *UnaryOperation)

	PreVisitBinaryOperation(bo *BinaryOperation) bool
	PostVisitBinaryOperation(bo *BinaryOperation)

	PreVisitVariableExpression(ve *VariableExpression) bool
	PostVisitVariableExpression(ve *VariableExpression)
}
