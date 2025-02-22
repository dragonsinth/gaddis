package ast

type Program struct {
	Block *Block
	Scope *Scope
}

type Block struct {
	SourceInfo
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
