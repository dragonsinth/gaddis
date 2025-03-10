package debug

import (
	"errors"
	"fmt"
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/ast"
	"log"
	"strings"
)

func (ds *Session) play() {
	if !ds.running.CompareAndSwap(false, true) {
		return
	}

	ds.runMu.Lock()
	defer ds.runMu.Unlock()
	if ds.runState > PAUSE {
		log.Println("ERROR: invalid runstate:", ds.runState)
		return
	}
	if ds.isDone {
		log.Println("ERROR: already done")
		return
	}

	// clear state on a fresh run
	ds.runState = RUN
	ds.exception = nil
	ds.exceptionTrace = ""
	log.Println("running")

	go func() {
		defer ds.running.Store(false)
		p := ds.Exec
		ds.runMu.Lock()
		defer ds.runMu.Unlock()

		defer func() {
			// if we exit for any reason, reset all step state
			ds.stepType = STEP_NONE

			if r := recover(); r != nil {
				err := errors.New(fmt.Sprint(r))
				log.Println("panicking:", err)
				if ds.noDebug {
					// execute the panic; send trace to stderr
					var lines []ast.Position
					var trace []string
					ds.Exec.GetStackFrames(func(fr *asm.Frame, _ int, inst asm.Inst, _ int) {
						if lc, ok := inst.(asm.LibCall); ok {
							// if the top of stack is a libcall with a native exception, generate a synthetic frame
							lines = append(lines, inst.GetSourceInfo().Start)
							trace = append(trace, asm.FormatLibCall(lc))
						}
						lines = append(lines, inst.GetSourceInfo().Start)
						trace = append(trace, asm.FormatFrameScope(fr))
					})
					ds.Host.Panicked(err, lines, trace)
					ds.isDone = true
				} else {
					// stop on exception
					ds.runState = PAUSE
					ds.exception = err
					file := strings.TrimPrefix(ds.File, ds.Opts.WorkDir)
					ds.exceptionTrace = ds.Exec.GetStackTrace(file)
					ds.Host.Exception(err)
				}
			}
		}()

		runStateEvent := func() {
			switch ds.runState {
			case HALT:
				log.Println("halting")
			case PAUSE:
				ds.Host.Paused("pause")
				log.Println("pausing")
			case TERMINATE:
				ds.Host.Terminated()
				log.Println("terminating")
			default:
				panic(ds.runState)
			}
		}

		if ds.runState != RUN {
			runStateEvent()
			return
		}

		if ds.stopOnEntry {
			ds.stopOnEntry = false // just once
			ds.Host.Paused("entry")
			return
		}

		instructionCount := 0
		for p.Frame != nil {
			// someone is asking for the lock; eventually we'll yield it
			if ds.yield.Load() {
				func() {
					ds.runMu.Unlock()
					defer ds.runMu.Lock()
				}()

				if ds.runState != RUN {
					runStateEvent()
					return
				}
			}

			inst := p.Code[p.PC]
			if ds.stepType != STEP_NONE {
				stackDiff := len(p.Stack) - ds.stepFrame
				line := inst.GetSourceInfo().Start.Line
				switch ds.stepType {
				case STEP_NEXT:
					if stackDiff < 0 || (stackDiff == 0 && ds.stepLine != line) {
						ds.Host.Paused("step")
						ds.runState = PAUSE
						return
					}
				case STEP_IN:
					// break on any change
					if stackDiff != 0 || ds.stepLine != line {
						ds.Host.Paused("step")
						ds.runState = PAUSE
						return
					}
				case STEP_OUT:
					if stackDiff < 0 {
						ds.Host.Paused("step")
						ds.runState = PAUSE
						return
					}
				default:
					panic(ds.stepType)
				}
			}

			inst.Exec(p)
			p.PC++

			// must be after exec to avoid reentry
			if ds.lineBreaks[p.PC]+ds.instBreaks[p.PC] != 0 {
				ds.Host.Paused("breakpoint")
				ds.runState = PAUSE
				return
			}

			instructionCount++
			if instructionCount > asm.MAX_INSTRUCTIONS {
				panic("infinite loop detected")
			}
		}
		ds.isDone = true
		ds.Host.Exited(0)
	}()
}
