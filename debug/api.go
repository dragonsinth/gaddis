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
func (ds *Session) Step(stepType StepType) {
	ds.withOuterLock(func() {
		p := ds.Exec
		si := p.Code[p.PC].GetSourceInfo()
		ds.stepType = stepType
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
		copy(p.Frame.Locals, p.Frame.Args)
		p.Frame.Eval = p.Frame.Eval[:0]
	})
}

func (ds *Session) GetStackFrames(f func(*asm.Frame, int, asm.Inst)) {
	ds.withOuterLock(func() {
		ds.Exec.GetStackFrames(f)
	})
}

func (ds *Session) SetNoDebug() {
	ds.withOuterLock(func() {
		ds.noDebug = true
	})
}

func (ds *Session) SetBreakpoints(bps []int) {
	ds.withOuterLock(func() {
		clear(ds.instBreaks)
		clear(ds.lineBreaks)
		for _, bp := range bps {
			if bp < 0 || bp >= len(ds.lineBreaks) {
				continue
			}
			pc := ds.SourceToInst[bp]
			if pc < 0 {
				continue
			}
			ds.lineBreaks[bp] = 1
			ds.instBreaks[pc] = 1
		}
	})
}
