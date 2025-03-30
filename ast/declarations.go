package ast

type HasName interface {
	GetName() string
}

type VarDecl struct {
	SourceInfo
	Name      string
	Type      Type
	DimExprs  []Expression
	Expr      Expression
	IsConst   bool
	IsField   bool
	IsPrivate bool
	IsParam   bool
	IsRef     bool

	Scope *Scope // collect
	Id    int
	Dims  []int // typecheck
}

func (vd *VarDecl) Visit(v Visitor) {
	if !v.PreVisitVarDecl(vd) {
		return
	}
	for _, d := range vd.DimExprs {
		d.Visit(v)
	}
	if vd.Expr != nil {
		vd.Expr.Visit(v)
	}
	v.PostVisitVarDecl(vd)
}

func (vd *VarDecl) GetName() string {
	return vd.Name
}
