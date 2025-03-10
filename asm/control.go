package asm

import (
	"fmt"
)

type Jump struct {
	baseInst
	Label *Label
}

func (i Jump) Exec(p *Execution) {
	p.PC = i.Label.PC - 1 // will advance to next instruction
}

func (i Jump) String() string {
	return fmt.Sprintf("jump %s", i.Label)
}

type JumpFalse struct {
	baseInst
	Label *Label
}

func (i JumpFalse) Exec(p *Execution) {
	v := p.Pop().(bool)
	if !v {
		p.PC = i.Label.PC - 1 // will advance to next instruction
	}
}

func (i JumpFalse) String() string {
	return fmt.Sprintf("jump_false %s", i.Label)
}

type JumpTrue struct {
	baseInst
	Label *Label
}

func (i JumpTrue) Exec(p *Execution) {
	v := p.Pop().(bool)
	if v {
		p.PC = i.Label.PC - 1 // will advance to next instruction
	}
}

func (i JumpTrue) String() string {
	return fmt.Sprintf("jump_true %s", i.Label)
}

type ForInt struct {
	baseInst
}

func (i ForInt) Exec(p *Execution) {
	step := p.Pop().(int64)
	stop := p.Pop().(int64)
	val := p.Pop().(int64)
	ref := p.Pop().(*any)
	*ref = val
	if step < 0 {
		p.Push(val >= stop)
	} else {
		p.Push(val <= stop)
	}
}

func (i ForInt) String() string {
	return "for_int"
}

type ForReal struct {
	baseInst
}

func (i ForReal) Exec(p *Execution) {
	step := p.Pop().(float64)
	stop := p.Pop().(float64)
	val := p.Pop().(float64)
	ref := p.Pop().(*any)
	*ref = val
	if step < 0 {
		p.Push(val >= stop)
	} else {
		p.Push(val <= stop)
	}
}

func (i ForReal) String() string {
	return "for_real"
}

type StepInt struct {
	baseInst
}

func (i StepInt) Exec(p *Execution) {
	step := p.Pop().(int64)
	stop := p.Pop().(int64)
	ref := p.Pop().(*any)
	val := (*ref).(int64)
	val += step
	*ref = val
	if step < 0 {
		p.Push(val >= stop)
	} else {
		p.Push(val <= stop)
	}
}

func (i StepInt) String() string {
	return "step_int"
}

type StepReal struct {
	baseInst
}

func (i StepReal) Exec(p *Execution) {
	step := p.Pop().(float64)
	stop := p.Pop().(float64)
	ref := p.Pop().(*any)
	val := (*ref).(float64)
	val += step
	*ref = val
	if step < 0 {
		p.Push(val >= stop)
	} else {
		p.Push(val <= stop)
	}
}

func (i StepReal) String() string {
	return "step_real"
}
