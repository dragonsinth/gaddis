package debug

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/dragonsinth/gaddis"
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/gogen/builtins"
	"log"
	"math/rand"
	"os"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

type EventHost interface {
	Paused(reason string)
	BreakOnException() bool
	Exception(err error)
	Panicked(error, []ast.Position, []string)
	Exited(code int)
	Terminated()
}

type Opts struct {
	IsTest     bool
	Stdin      []byte
	Stdout     func(string)
	WorkingDir string // TODO
}

type Session struct {
	Host      EventHost
	Prog      *ast.Program
	Assembled *asm.Assembly
	File      string
	Source    string
	Asm       string

	Context *asm.ExecutionContext
	Exec    *asm.Execution

	NLines, NInst int
	SourceToInst  []int // line mapping from source to instruction; -1 for lines that have no instruction
	InstToSource  []int // line mapping from instruction to source, cannot be empty

	// Private running state

	yield     atomic.Bool // ask the vm to yield so we can grab the mutex
	running   atomic.Bool
	runMu     sync.Mutex
	runState  RunState
	isDone    bool // terminated/exit/exception
	exception error

	lineBreaks []byte // what source lines to break on
	instBreaks []byte // what instruction lines to break on

	stopOnEntry bool
	stepType    StepType
	stepLine    int
	stepFrame   int
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

func New(
	filename string,
	compileErr func(err ast.Error),
	host EventHost,
	opts Opts,
) (*Session, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", filename, err)
	}
	src := string(buf)

	prog, _, errs := gaddis.Compile(src)
	if len(errs) > 0 {
		for _, err := range errs {
			compileErr(err)
		}
		return nil, fmt.Errorf("compile errors in %s", filename)
	}

	assembled := asm.Assemble(prog)
	asmSrc := assembled.AsmDump(src)

	ds := &Session{
		Host:      host,
		Prog:      prog,
		Assembled: assembled,
		File:      filename,
		Source:    src,
		Asm:       asmSrc,
	}

	// source/asm line mapping.
	ds.NInst = len(assembled.Code)
	ds.InstToSource = make([]int, ds.NInst)
	for i, inst := range assembled.Code {
		line := inst.GetSourceInfo().Start.Line
		ds.InstToSource[i] = line
		ds.NLines = max(ds.NLines, line+1)
	}
	ds.SourceToInst = make([]int, ds.NLines)
	for i := range ds.SourceToInst {
		ds.SourceToInst[i] = -1
	}
	for i, inst := range assembled.Code {
		line := inst.GetSourceInfo().GetSourceInfo().Start.Line
		if ds.SourceToInst[line] < 0 {
			ds.SourceToInst[line] = i
		}
	}

	ds.Reset(opts)
	return ds, nil
}

func (ds *Session) MapSourceLine(i int) int {
	if i < 0 || i >= ds.NLines {
		return -1
	}
	return ds.SourceToInst[i]
}

func (ds *Session) Reset(opts Opts) {
	var seed int64
	if !opts.IsTest {
		seed = time.Now().UnixNano()
	}

	ec := &asm.ExecutionContext{
		Rng: rand.New(rand.NewSource(seed)),
		IoContext: builtins.IoContext{
			Stdin:  bufio.NewScanner(bytes.NewReader(opts.Stdin)),
			Stdout: &bufferedSyncWriter{out: opts.Stdout},
		},
	}

	exec := ds.Assembled.NewExecution(ec)

	ds.Context = ec
	ds.Exec = exec

	ds.yield.Store(false)
	ds.isDone = false
	ds.runState = PAUSE
	ds.exception = nil
	if ds.lineBreaks == nil {
		ds.lineBreaks = make([]byte, ds.NLines)
		ds.instBreaks = make([]byte, ds.NInst)
	}
	ds.stepType = STEP_NONE
}

func (ds *Session) Play() {
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

	ds.runState = RUN
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
				if ds.Host.BreakOnException() {
					// pause the debugger
					ds.runState = PAUSE
					ds.exception = err
					ds.Host.Exception(err)
				} else {
					// send trace to stderr
					var lines []ast.Position
					var trace []string
					ds.Exec.GetStackFrames(func(fr *asm.Frame, _ int, inst asm.Inst) {
						lines = append(lines, inst.GetSourceInfo().Start)
						trace = append(trace, asm.FormatFrameScope(fr))
					})
					ds.Host.Panicked(err, lines, trace)
					ds.isDone = true
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
			if ds.instBreaks[p.PC] != 0 {
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

func (ds *Session) Terminate() {
	ds.yield.Store(true)
	ds.runMu.Lock()
	defer ds.runMu.Unlock()
	ds.yield.Store(false)
	ds.runState = TERMINATE
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

func (ds *Session) SetBreakpoints(bps []int) []int {
	bps = slices.DeleteFunc(bps, func(bp int) bool {
		return ds.MapSourceLine(bp) < 0
	})
	ds.withOuterLock(func() {
		clear(ds.instBreaks)
		clear(ds.lineBreaks)
		for _, bp := range bps {
			ds.lineBreaks[bp] = 1
			idx := ds.SourceToInst[bp]
			ds.instBreaks[idx] = 1
		}
	})
	return bps
}

func (ds *Session) StopOnEntry() {
	ds.withOuterLock(func() {
		ds.stopOnEntry = true
	})
}

func (ds *Session) GetStackFrames(f func(*asm.Frame, int, asm.Inst)) {
	ds.withOuterLock(func() {
		ds.Exec.GetStackFrames(f)
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

func (ds *Session) withOuterLock(f func()) {
	// force a yield, acquire the lock
	ds.yield.Store(true)
	ds.runMu.Lock()
	defer ds.runMu.Unlock()
	ds.yield.Store(false)
	f()
}
