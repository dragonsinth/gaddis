package ast

type OpenStmt struct {
	SourceInfo
	File Expression
	Name Expression
}

func (os *OpenStmt) Visit(v Visitor) {
	if !v.PreVisitOpenStmt(os) {
		return
	}
	os.File.Visit(v)
	os.Name.Visit(v)
	v.PostVisitOpenStmt(os)
}

func (*OpenStmt) isStatement() {}

type CloseStmt struct {
	SourceInfo
	File Expression
}

func (cs *CloseStmt) Visit(v Visitor) {
	if !v.PreVisitCloseStmt(cs) {
		return
	}
	cs.File.Visit(v)
	v.PostVisitCloseStmt(cs)
}

func (*CloseStmt) isStatement() {}

type ReadStmt struct {
	SourceInfo
	File  Expression
	Exprs []Expression
}

func (rs *ReadStmt) Visit(v Visitor) {
	if !v.PreVisitReadStmt(rs) {
		return
	}
	rs.File.Visit(v)
	for _, expr := range rs.Exprs {
		expr.Visit(v)
	}
	v.PostVisitReadStmt(rs)
}

func (*ReadStmt) isStatement() {}

type WriteStmt struct {
	SourceInfo
	File  Expression
	Exprs []Expression
}

func (ws *WriteStmt) Visit(v Visitor) {
	if !v.PreVisitWriteStmt(ws) {
		return
	}
	ws.File.Visit(v)
	for _, expr := range ws.Exprs {
		expr.Visit(v)
	}
	v.PostVisitWriteStmt(ws)
}

func (*WriteStmt) isStatement() {}
