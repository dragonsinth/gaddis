package asm

import (
	"errors"
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/gogen/builtins"
	"reflect"
	"strings"
)

const MAX_INSTRUCTIONS = 1 << 30
const MAX_STACK = 1024

func (a *Assembly) NewExecution(ec *ExecutionContext) *Execution {
	p := &Execution{
		PC:   0,
		Code: a.Code,
		Stack: []Frame{{
			Scope:  a.GlobalScope,
			Start:  0,
			Return: 0,
			Locals: make([]any, len(a.GlobalScope.Locals)),
			Eval:   make([]any, 0, 16),
		}},
		Frame: nil,
		Lib:   ec.CreateLibrary(),
	}
	p.Frame = &p.Stack[0]
	return p
}

type Execution struct {
	PC    int
	Code  []Inst
	Stack []Frame
	Frame *Frame
	Lib   []LibFunc
}

type Frame struct {
	Scope  *ast.Scope
	Start  int   // starting pc for this function
	Return int   // return address
	Args   []any // original function args
	Locals []any // current params+locals
	Eval   []any

	// try/catch stack?
}

func (p *Execution) Run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()

	instructionCount := 0
	for p.Frame != nil {
		inst := p.Code[p.PC]
		inst.Exec(p)
		p.PC++

		instructionCount++
		if instructionCount > MAX_INSTRUCTIONS {
			panic("infinite loop detected")
		}
	}
	return nil
}

func (p *Execution) GetStackFrames(f func(fr *Frame, id int, inst Inst)) {
	pc := p.PC
	for i := len(p.Stack) - 1; i >= 0; i-- {
		fr := p.Stack[i]
		inst := p.Code[pc]
		f(&fr, i+1, inst)
		pc = fr.Return
	}
}

func (p *Execution) GetStackTrace(filename string) string {
	var sb strings.Builder
	p.GetStackFrames(func(fr *Frame, _ int, inst Inst) {
		line := inst.GetSourceInfo().Start.Line + 1
		scope := FormatFrameScope(fr)
		_, _ = fmt.Fprintf(&sb, "%s:%d: in %s\n", filename, line, scope)
	})
	return sb.String()
}

func FormatFrameScope(fr *Frame) string {
	var sb strings.Builder
	sb.WriteString(fr.Scope.String())
	if !fr.Scope.IsGlobal {
		sb.WriteRune('(')
		for i, arg := range fr.Args {
			if i > 0 {
				sb.WriteRune(',')
			}
			sb.WriteString(DebugStringVal(arg))
		}
		sb.WriteRune(')')
	}
	return sb.String()
}

func (p *Execution) Push(val any) {
	if val == nil {
		panic("here")
	}
	p.Frame.Eval = append(p.Frame.Eval, val)
}

func (p *Execution) Pop() any {
	tip := len(p.Frame.Eval) - 1
	ret := p.Frame.Eval[tip]
	p.Frame.Eval = p.Frame.Eval[:tip]
	return ret
}

func (p *Execution) PopN(n int) []any {
	tip := len(p.Frame.Eval) - n
	ret := p.Frame.Eval[tip:]
	p.Frame.Eval = p.Frame.Eval[:tip]
	return ret
}

func DebugStringVal(arg any) string {
	if arg == nil {
		return "<nil>"
	}
	if reflect.TypeOf(arg).Kind() == reflect.Pointer {
		return "<ref>"
	}
	if arg == builtins.TabDisplay {
		return "Tab"
	}

	var sb strings.Builder
	switch typedArg := arg.(type) {
	case bool:
		if typedArg {
			return "True"
		} else {
			return "False"
		}
	case string:
		panic(typedArg) // should be impossible
	case []byte:
		_, _ = fmt.Fprintf(&sb, "%#v", string(typedArg))
	case byte:
		_, _ = fmt.Fprintf(&sb, "%#v", rune(typedArg))
	default:
		_, _ = fmt.Fprint(&sb, arg)
	}
	return sb.String()
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

var _ = zeroValue
