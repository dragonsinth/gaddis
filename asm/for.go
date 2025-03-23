package asm

type ForInt struct {
	baseInst
}

func (i ForInt) Exec(p *Execution) {
	step := p.Pop().(int64)
	stop := p.Pop().(int64)
	val := p.Pop().(int64)
	ref := p.Pop().(*any)
	*ref = val
	if step < 0 {
		p.Push(val >= stop)
	} else {
		p.Push(val <= stop)
	}
}

func (i ForInt) String() string {
	return "for int"
}

type ForReal struct {
	baseInst
}

func (i ForReal) Exec(p *Execution) {
	step := p.Pop().(float64)
	stop := p.Pop().(float64)
	val := p.Pop().(float64)
	ref := p.Pop().(*any)
	*ref = val
	if step < 0 {
		p.Push(val >= stop)
	} else {
		p.Push(val <= stop)
	}
}

func (i ForReal) String() string {
	return "for real"
}

type StepInt struct {
	baseInst
}

func (i StepInt) Exec(p *Execution) {
	step := p.Pop().(int64)
	stop := p.Pop().(int64)
	ref := p.Pop().(*any)
	val := (*ref).(int64)
	val += step
	*ref = val
	if step < 0 {
		p.Push(val >= stop)
	} else {
		p.Push(val <= stop)
	}
}

func (i StepInt) String() string {
	return "step int"
}

type StepReal struct {
	baseInst
}

func (i StepReal) Exec(p *Execution) {
	step := p.Pop().(float64)
	stop := p.Pop().(float64)
	ref := p.Pop().(*any)
	val := (*ref).(float64)
	val += step
	*ref = val
	if step < 0 {
		p.Push(val >= stop)
	} else {
		p.Push(val <= stop)
	}
}

func (i StepReal) String() string {
	return "step real"
}

type ForEach struct {
	baseInst
}

func (i ForEach) Exec(p *Execution) {
	arr := p.Pop().([]any)
	idx := p.Pop().(*any)
	ref := p.Pop().(*any)

	idxVal := (*idx).(int64)
	idxVal++
	*idx = idxVal

	if idxVal < int64(len(arr)) {
		*ref = arr[idxVal]
		p.Push(true)
	} else {
		p.Push(false)
	}
}

func (i ForEach) String() string {
	return "foreach"
}
