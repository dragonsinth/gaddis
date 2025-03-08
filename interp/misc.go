package interp

import "github.com/dragonsinth/gaddis/ast"

type Dup struct {
	ast.SourceInfo
}

func (i Dup) Exec(p *Program) {
	v := p.Pop()
	p.Push(v)
	p.Push(v)
}

func (i Dup) String() string {
	return "dup"
}

type Pop struct {
	ast.SourceInfo
}

func (i Pop) Exec(p *Program) {
	p.Pop()
}

func (i Pop) String() string {
	return "pop"
}
