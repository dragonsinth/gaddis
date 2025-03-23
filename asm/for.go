package asm

import "strconv"

type IncrInt struct {
	baseInst
	Val int64
}

func (i IncrInt) Exec(p *Execution) {
	ref := p.Pop().(*any)
	refVal := (*ref).(int64)
	*ref = refVal + i.Val
}

func (i IncrInt) String() string {
	return "incr int " + strconv.FormatInt(i.Val, 10)
}

type IncrReal struct {
	baseInst
	Val float64
}

func (i IncrReal) Exec(p *Execution) {
	ref := p.Pop().(*any)
	refVal := (*ref).(float64)
	*ref = refVal + i.Val
}

func (i IncrReal) String() string {
	return "incr real " + strconv.FormatFloat(i.Val, 'g', -1, 64)
}
