package debug

import (
	"github.com/dragonsinth/gaddis/asm"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type Session struct {
	Opts   Opts
	Host   EventHost
	Source Source

	Exec *asm.Execution

	// Private running state

	yield    atomic.Bool // ask the vm to yield so we can grab the mutex
	running  atomic.Bool
	runMu    sync.Mutex
	runState RunState
	isDone   bool // terminated/exit/exception
	noDebug  bool

	exception       error
	exceptionTrace  string
	exceptionFrames []ErrFrame

	lineBreaks []byte // pcs to break for lines
	instBreaks []byte // pcs to break for inst

	stopOnEntry bool // remove this field in favor of "break before line" once debugger is single-execution
	stepType    StepType
	stepGran    StepGran
	stepInst    int
	stepLine    int
	stepFrame   int
}

func New(
	source Source,
	host EventHost,
	opts Opts,
) *Session {
	var seed int64
	if !opts.IsTest {
		seed = time.Now().UnixNano()
	}

	ec := &asm.ExecutionContext{
		Rng:        rand.New(rand.NewSource(seed)),
		IoProvider: opts.IoProvider,
	}

	exec := source.Assembled.NewExecution(ec)

	lineBreaks := source.Breakpoints.ComputeLineBreaks(opts.LineBreaks)
	instBreaks := source.Breakpoints.ComputeInstBreaks(opts.InstBreaks)

	stopOnEntry := opts.StopOnEntry ||
		(len(lineBreaks) > 0 && lineBreaks[0] != 0) ||
		(len(instBreaks) > 0 && instBreaks[0] != 0)

	return &Session{
		Opts:           opts,
		Host:           host,
		Source:         source,
		Exec:           exec,
		yield:          atomic.Bool{},
		running:        atomic.Bool{},
		runMu:          sync.Mutex{},
		runState:       PAUSE,
		isDone:         false,
		noDebug:        opts.NoDebug,
		exception:      nil,
		exceptionTrace: "",
		lineBreaks:     lineBreaks,
		instBreaks:     instBreaks,
		stopOnEntry:    stopOnEntry,
		stepType:       STEP_NONE,
		stepGran:       LineGran,
		stepInst:       0,
		stepLine:       0,
		stepFrame:      0,
	}
}
