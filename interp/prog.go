package interp

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
)

type Compilation struct {
	GlobalScope *ast.Scope
	Code        []Inst
}

func zeroValue(typ ast.Type) any {
	switch typ {
	case ast.Integer:
		return int64(0)
	case ast.Real:
		return float64(0)
	case ast.String:
		return []byte{}
	case ast.Character:
		return byte(0)
	case ast.Boolean:
		return false
	default:
		panic(typ)
	}
}

func (c *Compilation) NewProgram(ec *ExecutionContext) *Program {
	var globals []any
	for _, decl := range c.GlobalScope.Locals {
		globals = append(globals, zeroValue(decl.Type))
	}

	p := &Program{
		PC:   0,
		Code: c.Code,
		Stack: []Frame{{
			Scope:  c.GlobalScope,
			Return: 0,
			Locals: nil,
			Eval:   make([]any, 0, 16),
		}},
		Frame:   nil,
		Globals: globals,
		Lib:     ec.CreateLibrary(),
	}
	p.Frame = &p.Stack[0]
	return p
}

type Program struct {
	PC      int
	Code    []Inst
	Stack   []Frame
	Frame   *Frame
	Globals []any
	Lib     []LibFunc
}

type Frame struct {
	Scope  *ast.Scope
	Return int
	Args   []any // original function args
	Locals []any // current params+locals
	Eval   []any
	// try/catch stack?
}

type Label struct {
	Name string
	PC   int
}

func (l *Label) String() string {
	return fmt.Sprintf("%s(%d)", l.Name, l.PC)
}

func (p *Program) Push(val any) {
	if val == nil {
		panic("here")
	}
	p.Frame.Eval = append(p.Frame.Eval, val)
}

func (p *Program) Pop() any {
	tip := len(p.Frame.Eval) - 1
	ret := p.Frame.Eval[tip]
	p.Frame.Eval = p.Frame.Eval[:tip]
	return ret
}

func (p *Program) PopN(n int) []any {
	tip := len(p.Frame.Eval) - n
	ret := p.Frame.Eval[tip:]
	p.Frame.Eval = p.Frame.Eval[:tip]
	return ret
}

type Inst interface {
	ast.HasSourceInfo
	Exec(p *Program)
	String() string
}
