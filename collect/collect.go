package collect

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"slices"
)

// Collect constructs scopes, collects symbols.
func Collect(globalBlock *ast.Block) []ast.Error {
	v := New()
	globalBlock.Visit(v)
	slices.SortFunc(v.errors, func(a, b ast.Error) int {
		return a.Start.Pos - b.Start.Pos
	})
	return v.errors
}

func New() *Visitor {
	return &Visitor{}
}

type Visitor struct {
	currScope *ast.Scope
	errors    []ast.Error
}

var _ ast.Visitor = &Visitor{}

func (v *Visitor) PreVisitBlock(bl *ast.Block) bool {
	bl.Scope = ast.NewScope(v.currScope)
	v.currScope = bl.Scope
	return true
}

func (v *Visitor) PostVisitBlock(bl *ast.Block) {
	v.currScope = bl.Scope.Parent
}

func (v *Visitor) PreVisitVarDecl(vd *ast.VarDecl) bool {
	return true
}

func (v *Visitor) PostVisitVarDecl(vd *ast.VarDecl) {
	if existing, ok := v.currScope.Decls[vd.Name]; ok {
		v.Errorf(vd, "identifier %s %s already defined in this scope", existing.Type, existing.Name)
	}
	v.currScope.Decls[vd.Name] = vd
}

func (v *Visitor) PreVisitConstantStmt(cs *ast.ConstantStmt) bool {
	return true
}

func (v *Visitor) PostVisitConstantStmt(cs *ast.ConstantStmt) {
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

func (v *Visitor) PreVisitIntegerLiteral(il *ast.IntegerLiteral) bool {
	return true
}

func (v *Visitor) PostVisitIntegerLiteral(il *ast.IntegerLiteral) {}

func (v *Visitor) PreVisitRealLiteral(rl *ast.RealLiteral) bool {
	return true
}

func (v *Visitor) PostVisitRealLiteral(rl *ast.RealLiteral) {}

func (v *Visitor) PreVisitStringLiteral(sl *ast.StringLiteral) bool {
	return true
}

func (v *Visitor) PostVisitStringLiteral(sl *ast.StringLiteral) {}

func (v *Visitor) PreVisitCharacterLiteral(cl *ast.CharacterLiteral) bool {
	return true
}

func (v *Visitor) PostVisitCharacterLiteral(cl *ast.CharacterLiteral) {}

func (v *Visitor) PreVisitBooleanLiteral(bl *ast.BooleanLiteral) bool {
	return true
}

func (v *Visitor) PostVisitBooleanLiteral(bl *ast.BooleanLiteral) {}

func (v *Visitor) PreVisitUnaryOperation(uo *ast.UnaryOperation) bool {
	return true
}

func (v *Visitor) PostVisitUnaryOperation(uo *ast.UnaryOperation) {}

func (v *Visitor) PreVisitBinaryOperation(bo *ast.BinaryOperation) bool {
	return true
}

func (v *Visitor) PostVisitBinaryOperation(bo *ast.BinaryOperation) {}

func (v *Visitor) PreVisitVariableExpression(ve *ast.VariableExpression) bool {
	return true
}

func (v *Visitor) PostVisitVariableExpression(ve *ast.VariableExpression) {}

func (v *Visitor) Errorf(si ast.HasSourceInfo, fmtStr string, args ...any) {
	v.errors = append(v.errors, ast.Error{
		SourceInfo: si.GetSourceInfo(),
		Desc:       fmt.Sprintf("declaration error: "+fmtStr, args...),
	})
}
