package collect

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/base"
)

// TODO: collect modules, functions, classes here.

// Collect constructs scopes, collects global symbols.
func Collect(prog *ast.Program) []ast.Error {
	prog.Scope = ast.NewGlobalScope(prog.Block)
	v := &Visitor{currScope: prog.Scope}
	prog.Block.Visit(v)
	return v.Errors
}

type Visitor struct {
	base.Visitor
	currScope *ast.Scope
}

var _ ast.Visitor = &Visitor{}

func (v *Visitor) PreVisitModuleStmt(ms *ast.ModuleStmt) bool {
	ms.Scope = ast.NewModuleScope(ms, v.currScope)
	return true
}

func (v *Visitor) PostVisitModuleStmt(ms *ast.ModuleStmt) {
	v.currScope = ms.Scope.Parent
	if existing := v.currScope.Decls[ms.Name]; existing != nil {
		v.Errorf(ms, "symbol %s redeclared in this scope; previous declaration: %s", ms.Name, existing)
	} else {
		v.currScope.AddModule(ms)
	}
}

func (v *Visitor) PreVisitFunctionStmt(fs *ast.FunctionStmt) bool {
	fs.Scope = ast.NewFunctionScope(fs, v.currScope)
	return true
}

func (v *Visitor) PostVisitFunctionStmt(fs *ast.FunctionStmt) {
	v.currScope = fs.Scope.Parent
	if existing := v.currScope.Decls[fs.Name]; existing != nil {
		v.Errorf(fs, "symbol %s redeclared in this scope; previous declaration: %s", fs.Name, existing)
	} else {
		v.currScope.AddFunction(fs)
	}
}
