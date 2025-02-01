package ast

import (
	"fmt"
	"strings"
)

type Statement interface {
	Node
	fmt.Stringer
}

type ConstantStmt struct {
	Type  Type
	Decls []*VarDecl
}

func (cs *ConstantStmt) Visit(v Visitor) {
	if !v.PreVisitConstantStmt(cs) {
		return
	}
	for _, d := range cs.Decls {
		d.Visit(v)
	}
	v.PostVisitConstantStmt(cs)
}

func (cs *ConstantStmt) String() string {
	return fmt.Sprintf("Constant %s %s", cs.Type, strings.Join(stringArray(cs.Decls), ", "))
}

type DeclareStmt struct {
	Type  Type
	Decls []*VarDecl
}

func (ds *DeclareStmt) Visit(v Visitor) {
	if !v.PreVisitDeclareStmt(ds) {
		return
	}
	for _, d := range ds.Decls {
		d.Visit(v)
	}
	v.PostVisitDeclareStmt(ds)
}

func (ds *DeclareStmt) String() string {
	return fmt.Sprintf("Declare %s %s", ds.Type, strings.Join(stringArray(ds.Decls), ", "))
}

type DisplayStmt struct {
	Exprs []Expression
}

func (ds *DisplayStmt) Visit(v Visitor) {
	if !v.PreVisitDisplayStmt(ds) {
		return
	}
	for _, e := range ds.Exprs {
		e.Visit(v)
	}
	v.PostVisitDisplayStmt(ds)
}

func (ds DisplayStmt) String() string {
	var exprStr []string
	for _, expr := range ds.Exprs {
		exprStr = append(exprStr, expr.String())
	}
	return fmt.Sprintf("Display %s", strings.Join(exprStr, ", "))
}

type InputStmt struct {
	Name string
	Ref  *VarDecl
}

func (is *InputStmt) Visit(v Visitor) {
	if !v.PreVisitInputStmt(is) {
		return
	}
	v.PostVisitInputStmt(is)
}

func (is InputStmt) String() string {
	return fmt.Sprintf("Input %s", is.Name)
}

type SetStmt struct {
	Name string
	Ref  *VarDecl
	Expr Expression
}

func (ss *SetStmt) Visit(v Visitor) {
	if !v.PreVisitSetStmt(ss) {
		return
	}
	ss.Expr.Visit(v)
	v.PostVisitSetStmt(ss)
}

func (ss SetStmt) String() string {
	return fmt.Sprintf("Set %s = %s", ss.Name, ss.Expr)
}

type IfStmt struct {
	If     *CondBlock
	ElseIf []*CondBlock
	Else   *Block
}

func (is *IfStmt) Visit(v Visitor) {
	if !v.PreVisitIfStmt(is) {
		return
	}
	is.If.Visit(v)
	for _, cb := range is.ElseIf {
		cb.Visit(v)
	}
	if is.Else != nil {
		is.Else.Visit(v)
	}
	v.PostVisitIfStmt(is)
}

func (is IfStmt) String() string {
	return "If..." // TODO
}

type CondBlock struct {
	Expr  Expression
	Block *Block
}

func (cb *CondBlock) Visit(v Visitor) {
	if !v.PreVisitCondBlock(cb) {
		return
	}
	cb.Expr.Visit(v)
	cb.Block.Visit(v)
	v.PostVisitCondBlock(cb)
}

func (cb *CondBlock) String() string {
	return fmt.Sprintf("%s Then [..]", cb.Expr)
}
