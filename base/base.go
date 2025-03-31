package base

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
)

type Visitor struct {
	Errors []ast.Error
	scope  *ast.Scope
	scopes []*ast.Scope
}

func (v *Visitor) Scope() *ast.Scope {
	return v.scope
}

func (v *Visitor) PushScope(scope *ast.Scope) {
	v.scopes = append(v.scopes, scope)
	v.scope = scope
}

func (v *Visitor) PopScope() {
	v.scopes = v.scopes[:len(v.scopes)-1]
	if len(v.scopes) == 0 {
		v.scope = nil
	} else {
		v.scope = v.scopes[len(v.scopes)-1]
	}
}

var _ ast.ScopeVisitor = &Visitor{}

func (v *Visitor) PreVisitBlock(bl *ast.Block) bool {
	return true
}

func (v *Visitor) PostVisitBlock(bl *ast.Block) {
}

func (v *Visitor) PreVisitVarDecl(vd *ast.VarDecl) bool {
	return true
}

func (v *Visitor) PostVisitVarDecl(vd *ast.VarDecl) {
}

func (v *Visitor) PreVisitDeclareStmt(ds *ast.DeclareStmt) bool {
	return true
}

func (v *Visitor) PostVisitDeclareStmt(ds *ast.DeclareStmt) {
}

func (v *Visitor) PreVisitDisplayStmt(ds *ast.DisplayStmt) bool {
	return true
}

func (v *Visitor) PostVisitDisplayStmt(ds *ast.DisplayStmt) {}

func (v *Visitor) PreVisitInputStmt(is *ast.InputStmt) bool {
	return true
}

func (v *Visitor) PostVisitInputStmt(is *ast.InputStmt) {}

func (v *Visitor) PreVisitSetStmt(ss *ast.SetStmt) bool {
	return true
}

func (v *Visitor) PostVisitSetStmt(ss *ast.SetStmt) {}

func (v *Visitor) PreVisitIfStmt(is *ast.IfStmt) bool {
	return true
}

func (v *Visitor) PreVisitOpenStmt(os *ast.OpenStmt) bool   { return true }
func (v *Visitor) PostVisitOpenStmt(os *ast.OpenStmt)       {}
func (v *Visitor) PreVisitCloseStmt(cs *ast.CloseStmt) bool { return true }
func (v *Visitor) PostVisitCloseStmt(cs *ast.CloseStmt)     {}
func (v *Visitor) PreVisitReadStmt(rs *ast.ReadStmt) bool   { return true }
func (v *Visitor) PostVisitReadStmt(rs *ast.ReadStmt)       {}
func (v *Visitor) PreVisitWriteStmt(ws *ast.WriteStmt) bool { return true }
func (v *Visitor) PostVisitWriteStmt(ws *ast.WriteStmt)     {}

func (v *Visitor) PostVisitIfStmt(is *ast.IfStmt) {}

func (v *Visitor) PreVisitCondBlock(cb *ast.CondBlock) bool {
	return true
}

func (v *Visitor) PostVisitCondBlock(cb *ast.CondBlock) {}

func (v *Visitor) PreVisitSelectStmt(ss *ast.SelectStmt) bool {
	return true
}

func (v *Visitor) PostVisitSelectStmt(ss *ast.SelectStmt) {}

func (v *Visitor) PreVisitCaseBlock(cb *ast.CaseBlock) bool {
	return true
}

func (v *Visitor) PostVisitCaseBlock(cb *ast.CaseBlock) {}

func (v *Visitor) PreVisitDoStmt(ds *ast.DoStmt) bool {
	return true
}

func (v *Visitor) PostVisitDoStmt(ds *ast.DoStmt) {}

func (v *Visitor) PreVisitWhileStmt(ws *ast.WhileStmt) bool {
	return true
}

func (v *Visitor) PostVisitWhileStmt(ws *ast.WhileStmt) {}

func (v *Visitor) PreVisitForStmt(fs *ast.ForStmt) bool {
	return true
}

func (v *Visitor) PostVisitForStmt(fs *ast.ForStmt) {}

func (v *Visitor) PreVisitForEachStmt(fs *ast.ForEachStmt) bool {
	return true
}

func (v *Visitor) PostVisitForEachStmt(fs *ast.ForEachStmt) {}

func (v *Visitor) PreVisitCallStmt(cs *ast.CallStmt) bool {
	return true
}

func (v *Visitor) PostVisitCallStmt(cs *ast.CallStmt) {}

func (v *Visitor) PreVisitModuleStmt(ms *ast.ModuleStmt) bool {
	return true
}

func (v *Visitor) PostVisitModuleStmt(ms *ast.ModuleStmt) {}

func (v *Visitor) PreVisitReturnStmt(rs *ast.ReturnStmt) bool {
	return true
}

func (v *Visitor) PostVisitReturnStmt(rs *ast.ReturnStmt) {}

func (v *Visitor) PreVisitFunctionStmt(fs *ast.FunctionStmt) bool {
	return true
}

func (v *Visitor) PostVisitFunctionStmt(fs *ast.FunctionStmt) {}

func (v *Visitor) PreVisitClassStmt(cs *ast.ClassStmt) bool {
	return true
}

func (v *Visitor) PostVisitClassStmt(cs *ast.ClassStmt) {
}

func (v *Visitor) PreVisitLiteral(i *ast.Literal) bool {
	return true
}

func (v *Visitor) PostVisitLiteral(i *ast.Literal) {}

func (v *Visitor) PreVisitParenExpr(pe *ast.ParenExpr) bool {
	return true
}

func (v *Visitor) PostVisitParenExpr(pe *ast.ParenExpr) {}

func (v *Visitor) PreVisitUnaryOperation(uo *ast.UnaryOperation) bool {
	return true
}

func (v *Visitor) PostVisitUnaryOperation(uo *ast.UnaryOperation) {}

func (v *Visitor) PreVisitBinaryOperation(bo *ast.BinaryOperation) bool {
	return true
}

func (v *Visitor) PostVisitBinaryOperation(bo *ast.BinaryOperation) {}

func (v *Visitor) PreVisitVariableExpr(ve *ast.VariableExpr) bool {
	return true
}

func (v *Visitor) PostVisitVariableExpr(ve *ast.VariableExpr) {}

func (v *Visitor) PreVisitCallExpr(ce *ast.CallExpr) bool {
	return true
}

func (v *Visitor) PostVisitCallExpr(ce *ast.CallExpr) {}

func (v *Visitor) PreVisitArrayRef(ar *ast.ArrayRef) bool {
	return true
}

func (v *Visitor) PostVisitArrayRef(ar *ast.ArrayRef) {}

func (v *Visitor) PreVisitArrayInitializer(ai *ast.ArrayInitializer) bool {
	return true
}

func (v *Visitor) PostArrayInitializer(ai *ast.ArrayInitializer) {}

func (v *Visitor) PreVisitNewExpr(ne *ast.NewExpr) bool {
	return true
}

func (v *Visitor) PostVisitNewExpr(ne *ast.NewExpr) {
}

func (v *Visitor) PreVisitThisRef(ref *ast.ThisRef) bool {
	return true
}

func (v *Visitor) PostVisitThisRef(ref *ast.ThisRef) {
}

func (v *Visitor) Errorf(si ast.HasSourceInfo, fmtStr string, args ...any) {
	v.Errors = append(v.Errors, ast.Error{
		SourceInfo: si.GetSourceInfo(),
		Desc:       fmt.Sprintf(fmtStr, args...),
	})
}
