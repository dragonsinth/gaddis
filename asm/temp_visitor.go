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
	// must create temp locals to track index and array
	// attribute to top of loop
	si := fs.SourceInfo.Head()
	ref := fs.Ref.(*ast.VariableExpr)
	fs.IndexTemp = &ast.VarDecl{
		SourceInfo: si,
		Name:       ref.Name + "$idx",
		Type:       ast.Integer,
		Expr:       &ast.Literal{SourceInfo: si, Type: ast.Integer, Val: int64(0)},
	}
	v.Scope().AddTempLocal(fs.IndexTemp)

	fs.ArrayTemp = &ast.VarDecl{
		SourceInfo: si,
		Name:       ref.Name + "$arr",
		Type:       fs.ArrayExpr.GetType(),
		Expr:       fs.ArrayExpr,
	}
	v.Scope().AddTempLocal(fs.ArrayTemp)
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
