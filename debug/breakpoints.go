package debug

import "github.com/dragonsinth/gaddis/asm"

type Breakpoints struct {
	NLines, NInst int
	sourceToInst  []int // line mapping from source to instruction; -1 for lines that have no instruction
	instToSource  []int // line mapping from instruction to source, cannot be empty
}

func NewBreakpoints(code []asm.Inst) *Breakpoints {
	var bps Breakpoints
	bps.NInst = len(code)
	bps.instToSource = make([]int, bps.NInst)
	for i, inst := range code {
		line := inst.GetSourceInfo().Start.Line
		bps.instToSource[i] = line
		bps.NLines = max(bps.NLines, line+1)
	}

	// prefill with invalid
	bps.sourceToInst = make([]int, bps.NLines)
	for i := range bps.sourceToInst {
		bps.sourceToInst[i] = -1
	}
	for i, inst := range code {
		line := inst.GetSourceInfo().GetSourceInfo().Start.Line
		if bps.sourceToInst[line] < 0 {
			bps.sourceToInst[line] = i
		}
	}
	return &bps
}

func (b Breakpoints) InstFromSource(srcLine int) int {
	if srcLine < 0 || srcLine >= b.NLines {
		return -1
	}
	return b.sourceToInst[srcLine]
}

func (b Breakpoints) SourceFromInst(instLine int) int {
	if instLine < 0 || instLine >= b.NInst {
		return -1
	}
	return b.instToSource[instLine]
}

func (b Breakpoints) ValidSrcLine(srcLine int) bool {
	if srcLine < 0 || srcLine >= b.NLines {
		return false
	}
	return b.sourceToInst[srcLine] >= 0
}
