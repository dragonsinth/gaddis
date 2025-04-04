package asm

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/base"
)

type TempVisitor struct {
	base.Visitor
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
	vd.Expr = &ast.Literal{SourceInfo: si, Type: ast.Integer, Val: int64(0)}
	v.Scope().AddVariable(vd)
	fs.Index = vd
	fs.IndexExpr = &ast.VariableExpr{
		SourceInfo: si,
		Name:       vd.Name,
		Qualifier:  nil,
		Ref:        vd,
		Type:       vd.Type,
	}
}

func (v *TempVisitor) PreVisitModuleStmt(ms *ast.ModuleStmt) bool {
	v.PushScope(ms.Scope)
	return true
}

func (v *TempVisitor) PostVisitModuleStmt(_ *ast.ModuleStmt) {
	v.PopScope()
}

func (v *TempVisitor) PreVisitFunctionStmt(fs *ast.FunctionStmt) bool {
	v.PushScope(fs.Scope)
	return true
}

func (v *TempVisitor) PostVisitFunctionStmt(_ *ast.FunctionStmt) {
	v.PopScope()
}

func (v *TempVisitor) PreVisitClassStmt(cs *ast.ClassStmt) bool {
	v.PushScope(cs.Scope)
	return true
}

func (v *TempVisitor) PostVisitClassStmt(cs *ast.ClassStmt) {
	v.PopScope()
}
