package ast

type Visitor interface {
	PreVisitBlock(bl *Block) bool
	PostVisitBlock(bl *Block)

	PreVisitVarDecl(vd *VarDecl) bool
	PostVisitVarDecl(vd *VarDecl)

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

	PreVisitCallStmt(cs *CallStmt) bool
	PostVisitCallStmt(cs *CallStmt)

	PreVisitModuleStmt(ms *ModuleStmt) bool
	PostVisitModuleStmt(ms *ModuleStmt)

	PreVisitReturnStmt(fs *ReturnStmt) bool
	PostVisitReturnStmt(fs *ReturnStmt)

	PreVisitFunctionStmt(fs *FunctionStmt) bool
	PostVisitFunctionStmt(fs *FunctionStmt)

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

	PreVisitVariableExpr(ve *VariableExpr) bool
	PostVisitVariableExpr(ve *VariableExpr)

	PreVisitCallExpr(ce *CallExpr) bool
	PostVisitCallExpr(ce *CallExpr)
}
