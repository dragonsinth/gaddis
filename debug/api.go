package debug

import (
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/ast"
)

// EventHost is how the VM pushes out debug events.
type EventHost interface {
	Paused(reason string)
	Exception(err error)
	Panicked(error, []ast.Position, []string)
	Exited(code int)
	Terminated()
}

type Opts struct {
	IsTest  bool
	Stdin   []byte
	Stdout  func(string)
	WorkDir string
}

type RunState int

const (
	// keep running
	RUN RunState = iota
	// halt with a pause event
	PAUSE
	// halt with no event
	HALT
	// halt with a terminate event
	TERMINATE
)

type StepType int

const (
	STEP_NONE StepType = iota
	STEP_NEXT
	STEP_IN
	STEP_OUT
)

type StepGran bool

const (
	LineGran StepGran = false
	InstGran StepGran = true
)

func (ds *Session) Play() {
	ds.play()
}

func (ds *Session) Pause() {
	ds.yield.Store(true)
	ds.runMu.Lock()
	defer ds.runMu.Unlock()
	ds.yield.Store(false)
	ds.runState = PAUSE
}

// Halt without sending any events. Wait for halt.
func (ds *Session) Halt() {
	ds.withOuterLock(func() {
		ds.runState = HALT
	})
	for ds.running.Load() {
		ds.withOuterLock(func() {})
	}
}

func (ds *Session) Terminate() {
	ds.yield.Store(true)
	ds.runMu.Lock()
	defer ds.runMu.Unlock()
	ds.yield.Store(false)
	ds.runState = TERMINATE
}

func (ds *Session) StopOnEntry() {
	ds.withOuterLock(func() {
		ds.stopOnEntry = true
	})
}

func (ds *Session) Step(stepType StepType, stepGran StepGran) {
	ds.withOuterLock(func() {
		p := ds.Exec
		si := p.Code[p.PC].GetSourceInfo()
		ds.stepType = stepType
		ds.stepGran = stepGran
		ds.stepInst = p.PC
		ds.stepLine = si.Start.Line
		ds.stepFrame = len(p.Stack)
	})
}

func (ds *Session) RestartFrame(id int) {
	ds.withOuterLock(func() {
		p := ds.Exec
		if id < 1 || id > len(p.Stack) {
			return // client error
		}
		p.Stack = p.Stack[:id]
		p.Frame = &p.Stack[id-1]
		p.PC = p.Frame.Start
		clear(p.Frame.Locals)
		copy(p.Frame.Params, p.Frame.Args)
		p.Frame.Eval = p.Frame.Eval[:0]
	})
}

// StackFrameFunc receives successive stack frames, from newest to oldest.
type StackFrameFunc func(frame *asm.Frame, frameId int, inst asm.Inst, pc int)

func (ds *Session) GetStackFrames(f func(*asm.Frame, int, asm.Inst, int)) {
	ds.withOuterLock(func() {
		ds.Exec.GetStackFrames(f)
	})
}

func (ds *Session) GetCurrentException() (string, error) {
	var trace string
	var exception error
	ds.withOuterLock(func() {
		trace = ds.exceptionTrace
		exception = ds.exception
	})
	return trace, exception
}

func (ds *Session) SetNoDebug() {
	ds.withOuterLock(func() {
		ds.noDebug = true
	})
}

func (ds *Session) SetLineBreakpoints(bps []int) {
	ds.withOuterLock(func() {
		clear(ds.lineBreaks)
		for _, bp := range bps {
			pc := ds.Source.Breakpoints.InstFromSource(bp)
			if pc < 0 {
				continue
			}
			ds.lineBreaks[pc] = 1
		}
	})
}

func (ds *Session) SetInstBreakpoints(pcs []int) {
	ds.withOuterLock(func() {
		clear(ds.instBreaks)
		for _, pc := range pcs {
			if pc < 0 || pc >= ds.Source.Breakpoints.NInst {
				continue
			}
			ds.instBreaks[pc] = 1
		}
	})
}
