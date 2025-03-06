package interp

import (
	"errors"
	"fmt"
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
		line := inst.GetSourceInfo().Start.Line
		scopeName := fr.Scope.String()
		_, _ = fmt.Fprintf(&sb, "\tat %s(%s:%d)\n", scopeName, filename, line)
		pc = fr.Return
	}
	return sb.String()
}
