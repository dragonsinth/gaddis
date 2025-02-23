package resolve

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/base"
)

// TODO: ensure the resolved thing is the correct type of thing.

// Resolve resolves symbols.
func Resolve(prog *ast.Program) []ast.Error {
	v := &Visitor{currScope: prog.Scope}
	prog.Block.Visit(v)
	return v.Errors
}

type Visitor struct {
	base.Visitor
	currScope *ast.Scope
}

var _ ast.Visitor = &Visitor{}

func (v *Visitor) PostVisitVarDecl(vd *ast.VarDecl) {
	if existing := v.currScope.Decls[vd.Name]; existing != nil {
		v.Errorf(vd, "symbol %s redeclared in this scope; previous declaration: %s", vd.Name, existing.String())
	} else {
		v.currScope.AddVariable(vd)
	}
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
