package asm

import "github.com/dragonsinth/gaddis/ast"

type Assembly struct {
	GlobalScope *ast.Scope
	Code        []Inst
	Labels      map[string]*Label
	Strings     []string
	Classes     []string
	Vtables     [][]*Label
}

type Inst interface {
	ast.HasSourceInfo
	Exec(p *Execution)
	String() string
	Sym() string
}

func (as *Assembly) AssembleExpression(expr ast.Expression) []Inst {
	v := &Visitor{
		labels: as.Labels,
	}
	expr.Visit(v)
	v.code = append(v.code, Halt{baseInst{expr.GetSourceInfo().Tail()}, 1})
	return v.code
}
