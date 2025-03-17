package debug

import (
	"errors"
	"github.com/dragonsinth/gaddis"
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/ast"
)

// EventHost is how the VM pushes out debug events.
type EventHost interface {
	Paused(reason string)
	Exception(err error)
	Panicked(error, []ErrFrame)
	Exited(code int)
	Terminated()
	SuppressAllEvents()
}

type ErrFrame struct {
	File     string
	Desc     string
	Pos      ast.Position
	IsNative bool
}

type Opts struct {
	IoProvider  gaddis.IoProvider
	IsTest      bool
	NoDebug     bool
	StopOnEntry bool
	LineBreaks  []int
	InstBreaks  []int
}

type RunState int

const (
	// keep running
	RUN RunState = iota
	// halt with a pause event
	PAUSE
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

func (ds *Session) IsRunning() bool {
	return ds.running.Load()
}

// Wait for the interpreter to stop running.
func (ds *Session) Wait() {
	for ds.running.Load() {
		ds.withOuterLock(func() {})
	}
}

func (ds *Session) Pause() {
	ds.withOuterLock(func() {
		if ds.noDebug {
			return
		}
		ds.runState = PAUSE
	})
}

// Halt terminates without sending any events.
func (ds *Session) Halt() {
	ds.withOuterLock(func() {
		ds.runState = TERMINATE
		ds.Host.SuppressAllEvents()
	})
}

func (ds *Session) Terminate() {
	ds.withOuterLock(func() {
		if ds.noDebug {
			return
		}
		ds.runState = TERMINATE
	})
}

func (ds *Session) Step(stepType StepType, stepGran StepGran) {
	ds.withOuterLock(func() {
		if ds.noDebug {
			return
		}
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
		if ds.noDebug {
			return
		}
		p := ds.Exec
		if id < 1 || id > len(p.Stack) {
			return // client error
		}
		p.Stack = p.Stack[:id]
		p.Frame = &p.Stack[id-1]
		p.PC = p.Frame.Start
		p.Frame.Params = nil
		p.Frame.Locals = nil
		p.Frame.Eval = nil

		// clear the exception state
		ds.exception = nil
		ds.exceptionTrace = ""
		ds.exceptionFrames = nil
	})
}

// StackFrameFunc receives successive stack frames, from newest to oldest.
type StackFrameFunc func(frame *asm.Frame, frameId int, inst asm.Inst, pc int)

func (ds *Session) GetStackFrames(f func(*asm.Frame, int, asm.Inst, int)) {
	ds.withOuterLock(func() {
		if ds.noDebug {
			return
		}
		ds.Exec.GetStackFrames(f)
	})
}

func (ds *Session) GetCurrentException() (string, error) {
	var trace string
	var exception error
	ds.withOuterLock(func() {
		if ds.noDebug {
			return
		}
		trace = ds.exceptionTrace
		exception = ds.exception
	})
	return trace, exception
}

func (ds *Session) UpdateLineBreakpoints(bps []int) {
	ds.withOuterLock(func() {
		if ds.noDebug {
			return
		}
		ds.lineBreaks = ds.Source.Breakpoints.ComputeLineBreaks(bps)
	})
}

func (ds *Session) UpdateInstBreakpoints(pcs []int) {
	ds.withOuterLock(func() {
		if ds.noDebug {
			return
		}
		ds.instBreaks = ds.Source.Breakpoints.ComputeInstBreaks(pcs)
	})
}

func (ds *Session) EvaluateExpressionInFrame(targetFrameId int, expr string) (val any, typ ast.Type, err error) {
	found := false
	ds.GetStackFrames(func(fr *asm.Frame, frameId int, inst asm.Inst, _ int) {
		if fr.Native != nil {
			return
		}
		if frameId == targetFrameId || fr.Scope.IsGlobal && targetFrameId == 0 {
			found = true
			val, typ, err = ds.evaluateExprInFrame(fr, expr)
		}
	})
	if !found {
		err = errors.New("frame not found")
	}
	return
}
