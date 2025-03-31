package ast

type ClassStmt struct {
	SourceInfo
	Name    string
	Extends string
	Type    *ClassType
	Block   *Block

	Scope *Scope // collect
}

func (cs *ClassStmt) Visit(v Visitor) {
	if !v.PreVisitClassStmt(cs) {
		return
	}
	cs.Block.Visit(v)
	v.PostVisitClassStmt(cs)
}

func (cs *ClassStmt) GetName() string {
	return cs.Name
}

func (cs *ClassStmt) isStatement() {}

type ThisRef struct {
	SourceInfo
	baseExpression
	Type *ClassType
}

func (ref *ThisRef) Visit(v Visitor) {
	if !v.PreVisitThisRef(ref) {
		return
	}
	v.PostVisitThisRef(ref)
}

func (ref *ThisRef) GetType() Type {
	return ref.Type
}
