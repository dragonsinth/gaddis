package ast

import "fmt"

type Scope struct {
	Parent *Scope
	Decls  map[string]*VarDecl
}

func (s *Scope) Lookup(name string) *VarDecl {
	if v := s.Decls[name]; v != nil {
		return v
	}
	if s.Parent != nil {
		return s.Parent.Lookup(name)
	}
	return nil
}

type VarDecl struct {
	Name    string
	Type    Type
	Expr    Expression
	IsConst bool
}

func (vd *VarDecl) Visit(v Visitor) {
	if !v.PreVisitVarDecl(vd) {
		return
	}
	if vd.Expr != nil {
		vd.Expr.Visit(v)
	}
	v.PostVisitVarDecl(vd)
}

func (vd *VarDecl) String() string {
	if vd.Expr != nil {
		return fmt.Sprintf("%s = %s", vd.Name, vd.Expr)
	}
	return fmt.Sprintf("%s", vd.Name)
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		Parent: nil,
		Decls:  map[string]*VarDecl{},
	}
}

type Block struct {
	Scope      *Scope
	Statements []Statement
}

func (bl *Block) Visit(v Visitor) {
	if !v.PreVisitBlock(bl) {
		return
	}
	for _, stmt := range bl.Statements {
		stmt.Visit(v)
	}
	v.PostVisitBlock(bl)
}
