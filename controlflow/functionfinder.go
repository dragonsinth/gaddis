package controlflow

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/base"
)

type FunctionFinder struct {
	base.Visitor
	functions []*ast.FunctionStmt
}

var _ ast.Visitor = &FunctionFinder{}

func (v *FunctionFinder) PreVisitFunctionStmt(fs *ast.FunctionStmt) bool {
	return true
}

func (v *FunctionFinder) PostVisitFunctionStmt(fs *ast.FunctionStmt) {
	v.functions = append(v.functions, fs)
}
