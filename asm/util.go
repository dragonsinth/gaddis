package asm

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func toFloat64(val any) float64 {
	switch v := val.(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	default:
		panic(v)
	}
}

func AsmDump(source string, code []Inst) string {
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
	return sb.String()
}

func PcRef(pc int) string {
	pc = pc * 4
	pc += 0x1000 // start program here
	if pc < 0 {
		return "0x0000"
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
