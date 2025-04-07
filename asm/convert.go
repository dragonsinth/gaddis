package asm

import "github.com/dragonsinth/gaddis/ast"

type IntToReal struct {
	baseInst
	ast.SourceInfo
}

func (i IntToReal) Exec(p *Execution) {
	tip := len(p.Frame.Eval) - 1
	p.Frame.Eval[tip] = float64(p.Frame.Eval[tip].(int64))
}

func (i IntToReal) String() string {
	return "conv int real"
}

type RealToInt struct {
	baseInst
	ast.SourceInfo
}

func (i RealToInt) Exec(p *Execution) {
	tip := len(p.Frame.Eval) - 1
	p.Frame.Eval[tip] = int64(p.Frame.Eval[tip].(float64))
}

func (i RealToInt) String() string {
	return "conv real int"
}
