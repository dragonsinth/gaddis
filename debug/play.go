package debug

import (
	"errors"
	"fmt"
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/ast"
	"log"
)

func (ds *Session) play() {
	defer close(ds.done)

	exitCode := 0
	defer func() {
		ds.Host.Exited(exitCode)
	}()

	p := ds.Exec
	instructionCount := 0

	for {
		ds.checkBreakpoints(p)

		for ds.runState == PAUSE || ds.yield.Load() {
			cmd, ok := <-ds.commands
			if !ok {
				return // we're being asked to exit
			}
			cmd(false)
		}

		if ds.exception != nil {
			// if we got back here and there's still an exception state, just exit immediately.
			ds.executePanic()
			exitCode = 1
			return
		}

		func() {
			// setup and catch any panics executing instructions
			defer func() {
				if r := recover(); r != nil {
					if _, ok := r.(sentinelIoInterrupt); ok {
						return // generally, restart frame with no exception
					}
					if _, ok := r.(sentinelIoExit); ok {
						p.Frame = nil // force a clean halt
						return
					}
					var err error
					if isErr, ok := r.(error); ok {
						err = isErr
					} else {
						err = errors.New(fmt.Sprint(r))
					}
					p.AddPanicFrames()

					log.Println("panicking:", err)

					var frames []ErrFrame
					ds.Exec.GetStackFrames(func(fr *asm.Frame, _ int, inst asm.Inst, _ int) {
						if fr.Native != nil {
							frames = append(frames, ErrFrame{
								File:     fr.Native.File,
								Desc:     fr.Native.Func,
								Pos:      ast.Position{Line: fr.Native.Line},
								IsNative: true,
							})
						} else {
							frames = append(frames, ErrFrame{
								Desc: asm.FormatFrameScope(fr),
								Pos:  inst.GetSourceInfo().Start,
							})
						}
					})

					ds.exception = &exceptionInfo{
						err:    err,
						trace:  ds.Exec.GetStackTrace(ds.Source.Path),
						frames: frames,
					}
					if !ds.Opts.NoDebug {
						// stop on exception, report error
						ds.runState = PAUSE
						ds.stepType = STEP_NONE
						ds.Host.Exception(err)
					}
				}
			}()

			// The actual part where we run instructions LOL.
			p.Code[p.PC].Exec(p)
			p.PC++
		}()

		if ds.exception != nil {
			if ds.Opts.NoDebug {
				ds.executePanic()
				exitCode = 1
				return
			} else {
				// stop on exception
				ds.runState = PAUSE
				ds.Host.Exception(ds.exception.err)
			}
		}

		if p.Frame == nil {
			return
		}

		instructionCount++
		if instructionCount > asm.MaxInstructions {
			panic("infinite loop detected")
		}
	}
}

func (ds *Session) checkBreakpoints(p *asm.Execution) {
	inst := p.Code[p.PC]
	if ds.stepType != STEP_NONE {
		stackDiff := len(p.Stack) - ds.stepFrame
		var ptrDiff bool
		if ds.stepGran == LineGran {
			ptrDiff = inst.GetSourceInfo().Start.Line != ds.stepLine
		} else {
			ptrDiff = p.PC != ds.stepInst
		}

		switch ds.stepType {
		case STEP_NEXT:
			if stackDiff < 0 || (stackDiff == 0 && ptrDiff) {
				ds.Host.Paused("step")
				ds.runState = PAUSE
				ds.stepType = STEP_NONE // reset step state
			}
		case STEP_IN:
			// break on any change
			if stackDiff != 0 || ptrDiff {
				ds.Host.Paused("step")
				ds.runState = PAUSE
				ds.stepType = STEP_NONE // reset step state
			}
		case STEP_OUT:
			if stackDiff < 0 {
				ds.Host.Paused("step")
				ds.runState = PAUSE
				ds.stepType = STEP_NONE // reset step state
			}
		default:
			panic(ds.stepType)
		}
	}

	if ds.lineBreaks[p.PC]+ds.instBreaks[p.PC] != 0 {
		ds.Host.Paused("breakpoint")
		ds.runState = PAUSE
		ds.stepType = STEP_NONE
	}
}

func (ds *Session) executePanic() {
	ds.Opts.Output(fmt.Sprintf("error: %s\n", ds.exception.err))
	ds.Opts.Output(ds.exception.trace)
	ds.Host.Panicked(ds.exception.err, ds.exception.frames)
}
