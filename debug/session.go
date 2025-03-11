package debug

import (
	"bufio"
	"bytes"
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/gogen/builtins"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type Session struct {
	Opts   Opts
	Host   EventHost
	Source Source

	Context *asm.ExecutionContext
	Exec    *asm.Execution

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
	source Source,
	host EventHost,
	opts Opts,
) *Session {
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

	exec := source.Assembled.NewExecution(ec)

	return &Session{
		Opts:           opts,
		Host:           host,
		Source:         source,
		Context:        ec,
		Exec:           exec,
		yield:          atomic.Bool{},
		running:        atomic.Bool{},
		runMu:          sync.Mutex{},
		runState:       PAUSE,
		isDone:         false,
		noDebug:        false,
		exception:      nil,
		exceptionTrace: "",
		lineBreaks:     make([]byte, source.Breakpoints.NInst),
		instBreaks:     make([]byte, source.Breakpoints.NInst),
		stopOnEntry:    false,
		stepType:       STEP_NONE,
		stepLine:       0,
		stepFrame:      0,
	}
}
