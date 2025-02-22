package resolve

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"slices"
)

// TODO: ensure the resolved thing is the correct type of thing.

// Resolve resolves symbols.
func Resolve(prog *ast.Program) []ast.Error {
	v := New()
	v.currScope = prog.Scope
	prog.Block.Visit(v)
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
	return true
}

func (v *Visitor) PostVisitBlock(bl *ast.Block) {}

func (v *Visitor) PreVisitVarDecl(vd *ast.VarDecl) bool {
	return true
}

func (v *Visitor) PostVisitVarDecl(vd *ast.VarDecl) {
	if existing := v.currScope.Decls[vd.Name]; existing != nil {
		v.Errorf(vd, "symbol %s redeclared in this scope; previous declaration: %s", vd.Name, existing.String())
	} else {
		v.currScope.AddVariable(vd)
	}
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

func (v *Visitor) PreVisitCallStmt(cs *ast.CallStmt) bool {
	return true
}

func (v *Visitor) PostVisitCallStmt(cs *ast.CallStmt) {
	ref := v.currScope.Lookup(cs.Name)
	if ref == nil {
		v.Errorf(cs, "unresolved symbol: %s", cs.Name)
	} else if ref.ModuleStmt != nil {
		cs.Ref = ref.ModuleStmt
	} else {
		v.Errorf(cs, "expected Module ref, got: %s", ref)
	}
}

func (v *Visitor) PreVisitModuleStmt(ms *ast.ModuleStmt) bool {
	// eagerly set the current scope for parameter scoping.
	v.currScope = ms.Scope
	return true
}

func (v *Visitor) PostVisitModuleStmt(ms *ast.ModuleStmt) {
	v.currScope = ms.Scope.Parent
}

func (v *Visitor) PreVisitReturnStmt(rs *ast.ReturnStmt) bool {
	return true
}

func (v *Visitor) PostVisitReturnStmt(rs *ast.ReturnStmt) {
	ref := v.currScope.FunctionStmt
	if ref == nil {
		v.Errorf(rs, "return statement without enclosing Function")
	} else {
		rs.Ref = ref
	}
}

func (v *Visitor) PreVisitFunctionStmt(fs *ast.FunctionStmt) bool {
	// eagerly set the current scope for parameter scoping.
	v.currScope = fs.Scope
	return true
}

func (v *Visitor) PostVisitFunctionStmt(fs *ast.FunctionStmt) {
	v.currScope = fs.Scope.Parent
}

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

func (v *Visitor) PreVisitVariableExpr(ve *ast.VariableExpr) bool {
	return true
}

func (v *Visitor) PostVisitVariableExpr(ve *ast.VariableExpr) {
	ref := v.currScope.Lookup(ve.Name)
	if ref == nil {
		v.Errorf(ve, "unresolved symbol: %s", ve.Name)
	} else if ref.VarDecl != nil {
		ve.Ref = ref.VarDecl
	} else {
		v.Errorf(ve, "expected variable ref, got: %s", ref)
	}
}

func (v *Visitor) PreVisitCallExpr(ce *ast.CallExpr) bool {
	return true
}

func (v *Visitor) PostVisitCallExpr(ce *ast.CallExpr) {
	ref := v.currScope.Lookup(ce.Name)
	if ref == nil {
		v.Errorf(ce, "unresolved symbol: %s", ce.Name)
	} else if ref.FunctionStmt != nil {
		ce.Ref = ref.FunctionStmt
	} else {
		v.Errorf(ce, "expected Function ref, got: %s", ref)
	}
}

func (v *Visitor) Errorf(si ast.HasSourceInfo, fmtStr string, args ...any) {
	v.errors = append(v.errors, ast.Error{
		SourceInfo: si.GetSourceInfo(),
		Desc:       fmt.Sprintf(fmtStr, args...),
	})
}
