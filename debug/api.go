package debug

import (
	"errors"
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/ast"
	"sync/atomic"
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
	Input       <-chan string
	Output      func(string)
	IsTest      bool
	NoDebug     bool
	StopOnEntry bool
	LineBreaks  []int
	InstBreaks  []int
}

type runState = int32

const (
	UNSTARTED  = runState(iota)
	RUN        // keep running
	PAUSE      // halt with a pause event
	TERMINATED // command channel is closed
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

func (ds *Session) IsStarted() bool {
	return atomic.LoadInt32(&ds.runState) != UNSTARTED
}

func (ds *Session) Play() {
	state := atomic.LoadInt32(&ds.runState)
	if state != UNSTARTED || !atomic.CompareAndSwapInt32(&ds.runState, state, RUN) {
		panic(state)
	}
	go ds.play()
}

// Wait for the interpreter to stop running.
func (ds *Session) Wait() {
	<-ds.done
}

func (ds *Session) Continue() {
	if ds.Opts.NoDebug {
		return
	}
	ds.runInVm(func(_ bool) {
		ds.runState = RUN
	})
}

func (ds *Session) Pause() {
	if ds.Opts.NoDebug {
		return
	}
	ds.runInVm(func(_ bool) {
		ds.Host.Paused("pause")
		ds.runState = PAUSE
	})
}

// Halt terminates without sending any events.
func (ds *Session) Halt() {
	ds.Host.SuppressAllEvents()
	ds.Terminate()
}

func (ds *Session) Terminate() {
	ds.yield.Store(true)
	state := atomic.LoadInt32(&ds.runState)
	if state != TERMINATED && atomic.CompareAndSwapInt32(&ds.runState, state, TERMINATED) {
		close(ds.commands)
	}
}

func (ds *Session) Step(stepType StepType, stepGran StepGran) {
	if ds.Opts.NoDebug {
		return
	}
	ds.runInVm(func(_ bool) {
		p := ds.Exec
		si := p.Code[p.PC].GetSourceInfo()
		ds.stepType = stepType
		ds.stepGran = stepGran
		ds.stepInst = p.PC
		ds.stepLine = si.Start.Line
		ds.stepFrame = len(p.Stack)
	})
}

func (ds *Session) RestartFrame(id int) (err error) {
	if ds.Opts.NoDebug {
		return
	}
	ds.runInVm(func(fromIoWait bool) {
		p := ds.Exec
		if id < 1 || id > len(p.Stack) {
			err = errors.New("invalid frame id")
			return // client error
		}
		fr := &p.Stack[id-1]
		if fr.Native != nil {
			err = errors.New("cannot restart native frame")
			return
		}

		p.Stack = p.Stack[:id]
		p.Frame = fr
		p.PC = p.Frame.Start
		p.Frame.Params = nil
		p.Frame.Locals = nil
		p.Frame.Eval = nil

		// clear the exception state
		ds.exception = nil

		if fromIoWait {
			// fix the real Go call stack back to the interpreter loop
			panic(sentinelIoInterrupt{})
		}
	})
	return
}

// StackFrameFunc receives successive stack frames, from newest to oldest.
type StackFrameFunc func(frame *asm.Frame, frameId int, inst asm.Inst, pc int)

func (ds *Session) GetStackFrames(f func(*asm.Frame, int, asm.Inst, int)) {
	if ds.Opts.NoDebug {
		return
	}
	ds.runInVm(func(_ bool) {
		ds.Exec.GetStackFrames(f)
	})
}

func (ds *Session) GetCurrentException() (trace string, err error) {
	if ds.Opts.NoDebug {
		return
	}
	ds.runInVm(func(_ bool) {
		if ds.exception != nil {
			trace, err = ds.exception.trace, ds.exception.err
		}
	})
	return
}

func (ds *Session) UpdateLineBreakpoints(bps []int) {
	if ds.Opts.NoDebug {
		return
	}
	if atomic.CompareAndSwapInt32(&ds.runState, UNSTARTED, PAUSE) {
		ds.lineBreaks = ds.Source.Breakpoints.ComputeLineBreaks(bps)
		atomic.StoreInt32(&ds.runState, UNSTARTED)
	} else {
		ds.runInVm(func(_ bool) {
			ds.lineBreaks = ds.Source.Breakpoints.ComputeLineBreaks(bps)
		})
	}
}

func (ds *Session) UpdateInstBreakpoints(pcs []int) {
	if ds.Opts.NoDebug {
		return
	}
	if atomic.CompareAndSwapInt32(&ds.runState, UNSTARTED, PAUSE) {
		ds.instBreaks = ds.Source.Breakpoints.ComputeInstBreaks(pcs)
		atomic.StoreInt32(&ds.runState, UNSTARTED)
	} else {
		ds.runInVm(func(_ bool) {
			ds.instBreaks = ds.Source.Breakpoints.ComputeInstBreaks(pcs)
		})
	}
}

func (ds *Session) EvaluateExpressionInFrame(targetFrameId int, expr string) (val any, typ ast.Type, err error) {
	if ds.Opts.NoDebug {
		return nil, nil, errors.New("no debug")
	}

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
