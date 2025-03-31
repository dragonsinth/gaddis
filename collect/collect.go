package collect

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/base"
)

// Collect constructs scopes, collects global symbols.
// We need to do this in two passes to correctly chain super scopes.
func Collect(prog *ast.Program) []ast.Error {
	prog.Scope = ast.NewGlobalScope(prog.Block)

	cc := &ClassCollector{classes: map[string]*ast.ClassStmt{}}
	prog.Visit(cc)
	for _, stmt := range cc.classes {
		createClassScope(cc.classes, stmt, prog.Scope)
	}

	v := &Visitor{}
	prog.Visit(v)
	return v.Errors
}

func createClassScope(classes map[string]*ast.ClassStmt, stmt *ast.ClassStmt, globalScope *ast.Scope) {
	if stmt.Scope != nil {
		return
	}
	parentScope := globalScope
	if stmt.Extends != "" {
		parent := classes[stmt.Extends]
		createClassScope(classes, parent, globalScope)
		parentScope = parent.Scope
	}
	stmt.Scope = ast.NewClassScope(stmt, parentScope)
	stmt.Type.Class = stmt
	stmt.Type.Scope = stmt.Scope
}

type ClassCollector struct {
	base.Visitor
	classes map[string]*ast.ClassStmt
}

var _ ast.Visitor = &Visitor{}

func (v *ClassCollector) PostVisitClassStmt(cs *ast.ClassStmt) {
	v.classes[cs.Name] = cs
}

type Visitor struct {
	base.Visitor
}

var _ ast.Visitor = &Visitor{}

func (v *Visitor) PostVisitVarDecl(vd *ast.VarDecl) {
	if existing := v.Scope().Decls[vd.Name]; existing != nil {
		v.Errorf(vd, "symbol %s redeclared in this scope; previous declaration: %s", vd.Name, existing.String())
	} else if nameMatchesClass(vd, vd.IsField) {
		v.Errorf(vd, "only a constructor is allowed to use the name of the enclosing class")
	} else {
		v.Scope().AddVariable(vd)
	}
}

func (v *Visitor) PreVisitModuleStmt(ms *ast.ModuleStmt) bool {
	ms.Scope = ast.NewModuleScope(ms, v.Scope())
	v.PushScope(ms.Scope)
	return true
}

func (v *Visitor) PostVisitModuleStmt(ms *ast.ModuleStmt) {
	v.PopScope()
	if existing := v.Scope().Decls[ms.Name]; existing != nil {
		v.Errorf(ms, "symbol %s redeclared in this scope; previous declaration: %s", ms.Name, existing)
	} else {
		if nameMatchesClass(ms, ms.IsMethod) {
			ms.IsConstructor = true
		}
		v.Scope().AddModule(ms)
	}
}

func (v *Visitor) PreVisitFunctionStmt(fs *ast.FunctionStmt) bool {
	fs.Scope = ast.NewFunctionScope(fs, v.Scope())
	v.PushScope(fs.Scope)
	return true
}

func (v *Visitor) PostVisitFunctionStmt(fs *ast.FunctionStmt) {
	v.PopScope()
	if existing := v.Scope().Decls[fs.Name]; existing != nil {
		v.Errorf(fs, "symbol %s redeclared in this scope; previous declaration: %s", fs.Name, existing)
	} else if nameMatchesClass(fs, fs.IsMethod) {
		v.Errorf(fs, "only a constructor is allowed to use the name of the enclosing class")
	} else {
		v.Scope().AddFunction(fs)
	}
}

func (v *Visitor) PreVisitClassStmt(cs *ast.ClassStmt) bool {
	v.PushScope(cs.Scope)
	return true
}

func (v *Visitor) PostVisitClassStmt(cs *ast.ClassStmt) {
	v.PopScope()
	if existing := v.Scope().Decls[cs.Name]; existing != nil {
		v.Errorf(cs, "symbol %s redeclared in this scope; previous declaration: %s", cs.Name, existing)
	} else {
		v.Scope().AddClass(cs)
	}
}

func nameMatchesClass(named ast.HasName, class *ast.ClassType) bool {
	return class != nil && class.GetName() == named.GetName()
}
