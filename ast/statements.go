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

func (c *ConstantStmt) Visit(v Visitor) {
	if !v.PreVisitConstantStmt(c) {
		return
	}
	for _, d := range c.Decls {
		d.Visit(v)
	}
	v.PostVisitConstantStmt(c)
}

func (c *ConstantStmt) String() string {
	return fmt.Sprintf("Constant %s %s", c.Type, strings.Join(stringArray(c.Decls), ", "))
}

type DeclareStmt struct {
	Type  Type
	Decls []*VarDecl
}

func (d *DeclareStmt) Visit(v Visitor) {
	if !v.PreVisitDeclareStmt(d) {
		return
	}
	for _, d := range d.Decls {
		d.Visit(v)
	}
	v.PostVisitDeclareStmt(d)
}

func (d *DeclareStmt) String() string {
	return fmt.Sprintf("Declare %s %s", d.Type, strings.Join(stringArray(d.Decls), ", "))
}

type DisplayStmt struct {
	Exprs []Expression
}

func (d *DisplayStmt) Visit(v Visitor) {
	if !v.PreVisitDisplayStmt(d) {
		return
	}
	for _, e := range d.Exprs {
		e.Visit(v)
	}
	v.PostVisitDisplayStmt(d)
}

func (d DisplayStmt) String() string {
	var exprStr []string
	for _, expr := range d.Exprs {
		exprStr = append(exprStr, expr.String())
	}
	return fmt.Sprintf("Display %s", strings.Join(exprStr, ", "))
}

type InputStmt struct {
	Name string
}

func (i *InputStmt) Visit(v Visitor) {
	if !v.PreVisitInputStmt(i) {
		return
	}
	v.PostVisitInputStmt(i)
}

func (i InputStmt) String() string {
	return fmt.Sprintf("Input %s", i.Name)
}

type SetStmt struct {
	Name string
	Expr Expression
}

func (s *SetStmt) Visit(v Visitor) {
	if !v.PreVisitSetStmt(s) {
		return
	}
	s.Expr.Visit(v)
	v.PostVisitSetStmt(s)
}

func (s SetStmt) String() string {
	return fmt.Sprintf("Set %s = %s", s.Name, s.Expr)
}
