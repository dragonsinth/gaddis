package ast

type HasName interface {
	GetName() string
}

type VarDecl struct {
	SourceInfo
	Name    string
	Type    Type
	Expr    Expression
	IsConst bool
	IsParam bool
	IsRef   bool // TODO: should this be part of the type?

	Scope *Scope // collect
	Id    int
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

func (vd *VarDecl) GetName() string {
	return vd.Name
}
