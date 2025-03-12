package asm

import (
	"errors"
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/lib"
	"math/rand"
	"reflect"
	"strings"
)

const MaxInstructions = 1 << 30
const MaxStack = 1024

type ExecutionContext struct {
	Rng *rand.Rand
	lib.IoContext
}

func (a *Assembly) NewExecution(ec *ExecutionContext) *Execution {
	extlib := lib.CreateLibrary(ec.IoContext, lib.RandContext{Rng: ec.Rng})

	p := &Execution{
		PC:   0,
		Code: a.Code,
		Stack: []Frame{{
			Scope:  a.GlobalScope,
			Start:  0,
			Return: 0,
			Params: nil,
		}},
		Frame: nil,
		Lib:   extlib,
	}
	p.Frame = &p.Stack[0]
	return p
}

type Execution struct {
	PC    int
	Code  []Inst
	Stack []Frame
	Frame *Frame
	Lib   []lib.Func
}

type Frame struct {
	Scope  *ast.Scope
	Start  int   // starting pc for this function
	Return int   // return address
	Args   []any // original function args
	Params []any // current params
	Locals []any // current locals
	Eval   []any // eval stack

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
		if instructionCount > MaxInstructions {
			panic("infinite loop detected")
		}
	}
	return nil
}

func (p *Execution) GetStackFrames(f func(fr *Frame, id int, inst Inst, pc int)) {
	pc := p.PC
	for i := len(p.Stack) - 1; i >= 0; i-- {
		fr := p.Stack[i]
		inst := p.Code[pc]
		f(&fr, i+1, inst, pc)
		pc = fr.Return
	}
}

func (p *Execution) GetStackTrace(filename string) string {
	var sb strings.Builder
	p.GetStackFrames(func(fr *Frame, _ int, inst Inst, _ int) {
		line := inst.GetSourceInfo().Start.Line + 1
		if lc, ok := inst.(LibCall); ok {
			// if the top of stack is a libcall with a native exception, generate a synthetic frame
			libScope := FormatLibCall(lc)
			_, _ = fmt.Fprintf(&sb, "%s:%d: in %s\n", filename, line, libScope)
		}
		scope := FormatFrameScope(fr)
		_, _ = fmt.Fprintf(&sb, "%s:%d: in %s\n", filename, line, scope)
	})
	return sb.String()
}

func FormatLibCall(lc LibCall) string {
	var sb strings.Builder
	// if the top of stack is a libcall with a native exception, generate a synthetic frame
	if lc.Type == ast.UnresolvedType {
		_, _ = fmt.Fprintf(&sb, "External %s", lc.Name)
	} else {
		_, _ = fmt.Fprintf(&sb, "External %s %s", lc.Type, lc.Name)
	}
	sb.WriteRune('(')
	if lc.NArg > 0 {
		_, _ = fmt.Fprintf(&sb, "[%d]", lc.NArg)
	}
	sb.WriteRune(')')
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
	if arg == lib.TabDisplay {
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
