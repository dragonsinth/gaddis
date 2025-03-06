package interp

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
)

type UnaryOpInt struct {
	ast.SourceInfo
	Op ast.Operator
}

func (i UnaryOpInt) Exec(p *Program) {
	a := p.Pop().(int64)
	switch i.Op {
	case ast.NEG:
		p.Push(-a)
	default:
		panic(i.Op)
	}
}

func (i UnaryOpInt) String() string {
	return fmt.Sprintf("%s_int", i.Op.Name())
}

type UnaryOpFloat struct {
	ast.SourceInfo
	Op ast.Operator
}

func (i UnaryOpFloat) Exec(p *Program) {
	a := toFloat64(p.Pop())
	switch i.Op {
	case ast.NEG:
		p.Push(-a)
	default:
		panic(i.Op)
	}
}

func (i UnaryOpFloat) String() string {
	return fmt.Sprintf("%s_real", i.Op.Name())
}

type UnaryOpBool struct {
	ast.SourceInfo
	Op ast.Operator
}

func (i UnaryOpBool) Exec(p *Program) {
	a := p.Pop().(bool)
	switch i.Op {
	case ast.NOT:
		p.Push(!a)
	default:
		panic(i.Op)
	}
}

func (i UnaryOpBool) String() string {
	return i.Op.Name()
}
