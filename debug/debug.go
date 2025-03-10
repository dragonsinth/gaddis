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
	Paused()
	Breakpoint()
	Panicked(error, ast.Position, string)
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

	yield    atomic.Bool // ask the vm to yield so we can grab the mutex
	running  atomic.Bool
	runMu    sync.Mutex
	runState RunState
	isDone   bool // terminated/exit/exception

	lineBreaks []byte // what source lines to break on
	instBreaks []byte // what instruction lines to break on
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
		return nil, fmt.Errorf("compile errors in %s: %w", filename, err)
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
	if ds.lineBreaks == nil {
		ds.lineBreaks = make([]byte, ds.NLines)
		ds.instBreaks = make([]byte, ds.NInst)
	}
}

func (ds *Session) Play() {
	if !ds.running.CompareAndSwap(false, true) {
		return
	}

	ds.runMu.Lock()
	defer ds.runMu.Unlock()
	if ds.runState > PAUSE {
		log.Println("invalid runstate:", ds.runState)
		return
	}
	ds.runState = RUN

	go func() {
		defer ds.running.Store(false)
		p := ds.Exec
		ds.runMu.Lock()
		defer ds.runMu.Unlock()

		defer func() {
			if r := recover(); r != nil {
				// TODO: allow an exception breakpoint halt here
				ds.isDone = true
				err := errors.New(fmt.Sprint(r))
				si := p.Code[p.PC].GetSourceInfo()
				trace := p.GetStackTrace(ds.File)
				ds.Host.Panicked(err, si.Start, trace)
			}
		}()

		instructionCount := 0
		for p.Frame != nil {
			// someone is asking for the lock; eventually we'll yield it
			if ds.yield.Load() {
				func() {
					ds.runMu.Unlock()
					defer ds.runMu.Lock()
				}()
			}

			switch ds.runState {
			case RUN:
				inst := p.Code[p.PC]
				inst.Exec(p)
				p.PC++
			case HALT:
				return
			case PAUSE:
				ds.Host.Paused()
				return
			case TERMINATE:
				ds.Host.Terminated()
				return
			default:
				panic(ds.runState)
			}

			// check breaks before looping
			if ds.instBreaks[p.PC] != 0 {
				ds.Host.Breakpoint()
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
		ds.instBreaks[0] = 1
		ds.lineBreaks[0] = 1
	})
}

func (ds *Session) GetStackFrames(f func(fr *asm.Frame, id int, inst asm.Inst)) {
	ds.withOuterLock(func() {
		ds.Exec.GetStackFrames(f)
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

func (ds *Session) MapSourceLine(i int) int {
	if i < 0 || i >= ds.NLines {
		return -1
	}
	return ds.SourceToInst[i]
}
