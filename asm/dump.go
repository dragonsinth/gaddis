package asm

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func (as *Assembly) Dump(source string) string {
	code := as.Code
	lines := strings.Split(source, "\n")
	lastLine := -1
	var sb bytes.Buffer
	for i, inst := range code {
		line := inst.GetSourceInfo().Start.Line
		if lastLine != line {
			lastLine = line
			_, _ = fmt.Fprintf(&sb, "; %d: %s\n", line+1, lines[line])
		}
		_, _ = fmt.Fprintf(&sb, "%s\t\t\t%s\n", PcRef(i), inst)
	}
	// Now dump all symbol tables.
	sb.WriteString("; -- functions\n")
	var endPcs []int
	for pc, inst := range code {
		if _, ok := inst.(End); ok {
			endPcs = append(endPcs, pc)
		}
	}

	for pc, inst := range code {
		if be, ok := inst.(Begin); ok {
			sb.WriteRune(';')
			sb.WriteString(PcRef(pc))
			sb.WriteRune('-')
			sb.WriteString(PcRef(endPcs[0]))
			sb.WriteRune(':')
			sb.WriteString(be.Label.Name)
			endPcs = endPcs[1:]
			sb.WriteRune('(')
			for i, vd := range be.Scope.Params {
				if i > 0 {
					sb.WriteRune('|')
				}
				sb.WriteString(vd.Type.String())
				sb.WriteRune('#')
				sb.WriteString(vd.Name)
			}
			sb.WriteRune(')')
			for i, vd := range be.Scope.Locals {
				if i > 0 {
					sb.WriteRune('|')
				}
				sb.WriteString(vd.Type.String())
				sb.WriteRune('#')
				sb.WriteString(vd.Name)
			}
			sb.WriteRune('\n')
		}
	}

	// Dump the string table
	sb.WriteString("; -- strings\n")
	for i, str := range as.Strings {
		_, _ = fmt.Fprintf(&sb, ";[%d] %q\n", i, str)
	}
	return sb.String()
}

func PcRef(pc int) string {
	pc = pc * 4
	pc += 0x1000 // start program here
	if pc < 0 {
		return ""
	}
	if pc > 0xffff {
		return "0xFFFF"
	}
	ret := fmt.Sprintf("0x%04X", pc)
	return ret
}

func RefPc(ref string) int {
	pc, err := strconv.ParseUint(ref, 0, 64)
	if err != nil {
		return -1
	}
	pc -= 0x1000
	if pc > 0xffff {
		return -1
	}
	pc = pc / 4
	return int(pc)
}
