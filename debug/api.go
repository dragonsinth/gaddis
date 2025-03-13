package debug

import (
	"errors"
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/ast"
	"io"
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
	IsTest  bool
	Stdin   io.Reader
	Stdout  func(string)
	WorkDir string
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

func (ds *Session) SetNoDebug() {
	ds.withOuterLock(func() {
		ds.noDebug = true
		ds.stopOnEntry = false
		ds.stepType = STEP_NONE
		clear(ds.instBreaks)
		clear(ds.instBreaks)
	})
}

func (ds *Session) StopOnEntry() {
	ds.withOuterLock(func() {
		if ds.noDebug {
			return
		}
		ds.stopOnEntry = true
	})
}

func (ds *Session) SetLineBreakpoints(bps []int) {
	ds.withOuterLock(func() {
		if ds.noDebug {
			return
		}
		clear(ds.lineBreaks)
		for _, bp := range bps {
			pc := ds.Source.Breakpoints.InstFromSource(bp)
			if pc < 0 {
				continue
			}
			ds.lineBreaks[pc] = 1
			if pc == 0 {
				ds.stopOnEntry = true // only way to actually break on inst 0
			}
		}
	})
}

func (ds *Session) SetInstBreakpoints(pcs []int) {
	ds.withOuterLock(func() {
		if ds.noDebug {
			return
		}
		clear(ds.instBreaks)
		for _, pc := range pcs {
			if pc < 0 || pc >= ds.Source.Breakpoints.NInst {
				continue
			}
			ds.instBreaks[pc] = 1
			if pc == 0 {
				ds.stopOnEntry = true // only way to actually break on inst 0
			}
		}
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
