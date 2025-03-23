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

	PreVisitForEachStmt(fs *ForEachStmt) bool
	PostVisitForEachStmt(fs *ForEachStmt)

	PreVisitCallStmt(cs *CallStmt) bool
	PostVisitCallStmt(cs *CallStmt)

	PreVisitModuleStmt(ms *ModuleStmt) bool
	PostVisitModuleStmt(ms *ModuleStmt)

	PreVisitReturnStmt(fs *ReturnStmt) bool
	PostVisitReturnStmt(fs *ReturnStmt)

	PreVisitFunctionStmt(fs *FunctionStmt) bool
	PostVisitFunctionStmt(fs *FunctionStmt)

	PreVisitLiteral(l *Literal) bool
	PostVisitLiteral(l *Literal)

	PreVisitParenExpr(pe *ParenExpr) bool
	PostVisitParenExpr(pe *ParenExpr)

	PreVisitUnaryOperation(uo *UnaryOperation) bool
	PostVisitUnaryOperation(uo *UnaryOperation)

	PreVisitBinaryOperation(bo *BinaryOperation) bool
	PostVisitBinaryOperation(bo *BinaryOperation)

	PreVisitVariableExpr(ve *VariableExpr) bool
	PostVisitVariableExpr(ve *VariableExpr)

	PreVisitCallExpr(ce *CallExpr) bool
	PostVisitCallExpr(ce *CallExpr)

	PreVisitArrayRef(ar *ArrayRef) bool
	PostVisitArrayRef(ar *ArrayRef)

	PreVisitArrayInitializer(ai *ArrayInitializer) bool
	PostArrayInitializer(ai *ArrayInitializer)
}
