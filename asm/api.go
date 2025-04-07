package asm

import (
	"github.com/dragonsinth/gaddis/ast"
)

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
