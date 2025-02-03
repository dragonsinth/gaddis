package ast

type VarDecl struct {
	SourceInfo
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
