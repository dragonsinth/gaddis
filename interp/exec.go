package interp

import (
	"errors"
	"fmt"
	"github.com/dragonsinth/gaddis/gogen/builtins"
	"reflect"
	"strings"
)

func (p *Program) Run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()
	for p.Frame != nil {
		inst := p.Code[p.PC]
		inst.Exec(p)
		p.PC++
	}
	return nil
}

func (p *Program) GetStackTrace(filename string) string {
	var sb strings.Builder
	pc := p.PC
	for i := len(p.Stack) - 1; i >= 0; i-- {
		fr := p.Stack[i]
		inst := p.Code[pc]
		line := inst.GetSourceInfo().Start.Line + 1
		scopeName := fr.Scope.String()
		if !fr.Scope.IsGlobal {
			var sb strings.Builder
			sb.WriteRune('(')
			for i, arg := range fr.Args {
				if i > 0 {
					sb.WriteRune(',')
				}
				sb.WriteString(display(arg))
			}
			sb.WriteRune(')')
			scopeName += sb.String()
		}
		_, _ = fmt.Fprintf(&sb, "%s:%d: in %s\n", filename, line, scopeName)
		pc = fr.Return
	}
	return sb.String()
}

func display(arg any) string {
	if reflect.TypeOf(arg).Kind() == reflect.Pointer {
		return "<ref>"
	}
	if arg == builtins.TabDisplay {
		return "tab"
	}

	var sb strings.Builder
	switch typedArg := arg.(type) {
	case bool:
		if typedArg {
			sb.WriteString("True")
		} else {
			sb.WriteString("False")
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
