package asm

import (
	"fmt"
	"slices"
)

type class struct {
	vtable *vtable
	fields []any
}

type vtable []int

type ObjNew struct {
	baseInst
	Name    string
	Vtable  *vtable
	NFields int
}

func (i ObjNew) Exec(p *Execution) {
	p.Push(&class{
		vtable: i.Vtable,
		fields: make([]any, i.NFields),
	})
}

func (i ObjNew) String() string {
	return fmt.Sprintf("object new %s", i.Name)
}

func (i ObjNew) Sym() string {
	return i.Name
}

type FieldRef struct {
	baseInst
	Name  string
	Index int
}

func (i FieldRef) Exec(p *Execution) {
	ref := p.Pop().(*class)
	p.Push(&ref.fields[i.Index])
}

func (i FieldRef) String() string {
	return fmt.Sprintf("&field[%d] #%s", i.Index, i.Name)
}

func (i FieldRef) Sym() string {
	return i.Name
}

type FieldVal struct {
	baseInst
	Class string
	Name  string
	Index int
}

func (i FieldVal) Exec(p *Execution) {
	ref := p.Pop().(*class)
	p.Push(ref.fields[i.Index])
}

func (i FieldVal) String() string {
	return fmt.Sprintf("field[%d] %s#%s", i.Index, i.Class, i.Name)
}

func (i FieldVal) Sym() string {
	return i.Class + "#" + i.Name
}

type VCall struct {
	baseInst
	Class string
	Name  string
	Index int
	NArgs int
}

func (i VCall) Exec(p *Execution) {
	if len(p.Stack) >= MaxStack {
		panic("stack overflow")
	}

	args := slices.Clone(p.PopN(i.NArgs))
	this := args[0].(*class)
	pc := (*this.vtable)[i.Index]

	inst := p.Code[pc]
	be := inst.(Begin)

	p.Stack = append(p.Stack, Frame{
		Scope:  be.Scope,
		Start:  pc,
		Return: p.PC,
		Args:   args,
	})
	p.Frame = &p.Stack[len(p.Stack)-1]
	p.PC = pc - 1 // will advance to next instruction
}

func (i VCall) String() string {
	return fmt.Sprintf("vcall[%d](%d) %s.%s", i.Index, i.NArgs, i.Class, i.Name)
}

func (i VCall) Sym() string {
	return i.Class + "." + i.Name
}
