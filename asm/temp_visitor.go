package asm

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/base"
)

type TempVisitor struct {
	base.Visitor
	currScope *ast.Scope
}

var _ ast.Visitor = &TempVisitor{}

func (v *TempVisitor) PostVisitForEachStmt(fs *ast.ForEachStmt) {
	// must create a temp local to track index
	// attribute to top of loop
	si := fs.SourceInfo.Head()
	ref := fs.Ref.(*ast.VariableExpr)
	vd := &ast.VarDecl{
		SourceInfo: si,
		Name:       ref.Name + "$idx",
		Type:       ast.Integer,
	}
	// pre-assign the index to -1
	vd.Expr = &ast.Literal{SourceInfo: si, Type: ast.Integer, Val: int64(-1)}
	v.currScope.AddVariable(vd)
	fs.Index = vd
}

func (v *TempVisitor) PreVisitModuleStmt(ms *ast.ModuleStmt) bool {
	v.currScope = ms.Scope
	return true
}

func (v *TempVisitor) PostVisitModuleStmt(_ *ast.ModuleStmt) {
	v.currScope = v.currScope.Parent
}

func (v *TempVisitor) PreVisitFunctionStmt(fs *ast.FunctionStmt) bool {
	v.currScope = fs.Scope
	return true
}

func (v *TempVisitor) PostVisitFunctionStmt(_ *ast.FunctionStmt) {
	v.currScope = v.currScope.Parent
}
