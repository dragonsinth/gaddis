package ast

type ClassStmt struct {
	SourceInfo
	Typ   *ClassType
	Block *Block

	Scope *Scope // collect
}

func (cs *ClassStmt) Visit(v Visitor) {
	if !v.PreVisitClassStmt(cs) {
		return
	}
	cs.Block.Visit(v)
	v.PostVisitClassStmt(cs)
}

func (cs *ClassStmt) isStatement() {}
