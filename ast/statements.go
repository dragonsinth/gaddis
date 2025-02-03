package ast

type Statement interface {
	Node
}

type Block struct {
	SourceInfo
	Statements []Statement
	Scope      *Scope
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

type ConstantStmt struct {
	SourceInfo
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

type DeclareStmt struct {
	SourceInfo
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

type DisplayStmt struct {
	SourceInfo
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

type InputStmt struct {
	SourceInfo
	Name string
	Ref  *VarDecl
}

func (is *InputStmt) Visit(v Visitor) {
	if !v.PreVisitInputStmt(is) {
		return
	}
	v.PostVisitInputStmt(is)
}

type SetStmt struct {
	SourceInfo
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

type IfStmt struct {
	SourceInfo
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

type CondBlock struct {
	SourceInfo
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
