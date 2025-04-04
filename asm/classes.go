package asm

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"slices"
)

type Object struct {
	Type   *ast.ClassType
	Vtable *vtable
	Fields []any
}

// For toString() debug.
func (obj *Object) String() string {
	if obj == nil {
		return "<nil>"
	}
	return "<" + obj.Type.String() + ">"
}

type vtable []int

type ObjNew struct {
	baseInst
	Type    *ast.ClassType
	Vtable  *vtable
	NFields int
}

func (i ObjNew) Exec(p *Execution) {
	p.Push(&Object{
		Type:   i.Type,
		Vtable: i.Vtable,
		Fields: make([]any, i.NFields),
	})
}

func (i ObjNew) String() string {
	return fmt.Sprintf("object new %s", i.Type)
}

func (i ObjNew) Sym() string {
	return i.Type.String()
}

type FieldRef struct {
	baseInst
	Name  string
	Index int
}

func (i FieldRef) Exec(p *Execution) {
	ref := p.Pop().(*Object)
	p.Push(&ref.Fields[i.Index])
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
	ref := p.Pop().(*Object)
	val := ref.Fields[i.Index]
	if val == nil {
		panic(fmt.Sprintf("field %s read before assignment", i.Name)) // TODO: zero-init classes
	}
	p.Push(val)
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
	this := args[0].(*Object)
	pc := (*this.Vtable)[i.Index]

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
