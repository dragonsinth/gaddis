package ast

type Program struct {
	Block *Block
	Scope *Scope
}

func (p *Program) GetSourceInfo() SourceInfo {
	return p.Block.GetSourceInfo()
}

func (p *Program) Visit(v Visitor) {
	if sv, ok := v.(ScopeVisitor); ok {
		sv.PushScope(p.Scope)
	}
	p.Block.Visit(v)
}

var _ Node = &Program{}

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

func (bl *Block) isStatement() {}
