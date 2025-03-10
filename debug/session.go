package debug

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/dragonsinth/gaddis"
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/gogen/builtins"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Session struct {
	Opts      Opts
	Host      EventHost
	Prog      *ast.Program
	Assembled *asm.Assembly
	File      string
	Source    string
	Asm       string

	Context *asm.ExecutionContext
	Exec    *asm.Execution

	NLines, NInst   int
	SourceToInst    []int // line mapping from source to instruction; -1 for lines that have no instruction
	InstToSource    []int // line mapping from instruction to source, cannot be empty
	ValidLineBreaks map[int]bool

	// Private running state

	yield    atomic.Bool // ask the vm to yield so we can grab the mutex
	running  atomic.Bool
	runMu    sync.Mutex
	runState RunState
	isDone   bool // terminated/exit/exception
	noDebug  bool

	exception      error
	exceptionTrace string

	lineBreaks []byte // pcs to break for lines
	instBreaks []byte // pcs to break for inst

	stopOnEntry bool
	stepType    StepType
	stepLine    int
	stepFrame   int
}

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
		Opts:      opts,
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

	// prefill with invalid
	ds.SourceToInst = make([]int, ds.NLines)
	for i := range ds.SourceToInst {
		ds.SourceToInst[i] = -1
	}
	ds.ValidLineBreaks = make(map[int]bool, ds.NLines)
	for i, inst := range assembled.Code {
		line := inst.GetSourceInfo().GetSourceInfo().Start.Line
		if ds.SourceToInst[line] < 0 {
			ds.SourceToInst[line] = i
			ds.ValidLineBreaks[line] = true
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
			Stdin:   bufio.NewScanner(bytes.NewReader(opts.Stdin)),
			Stdout:  &bufferedSyncWriter{out: opts.Stdout},
			WorkDir: opts.WorkDir,
		},
	}

	exec := ds.Assembled.NewExecution(ec)
	ds.Opts = opts
	ds.Context = ec
	ds.Exec = exec

	ds.yield.Store(false)
	ds.isDone = false
	ds.noDebug = false
	ds.runState = PAUSE
	ds.exception = nil
	if ds.instBreaks == nil {
		ds.lineBreaks = make([]byte, ds.NInst)
		ds.instBreaks = make([]byte, ds.NInst)
	}
	ds.stepType = STEP_NONE
}
