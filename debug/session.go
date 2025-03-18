package debug

import (
	"github.com/dragonsinth/gaddis"
	"github.com/dragonsinth/gaddis/asm"
	"io"
	"math/rand"
	"sync/atomic"
	"time"
)

type Session struct {
	Opts   Opts
	Host   EventHost
	Source Source

	Exec *asm.Execution

	// Private running state

	runState runState
	yield    atomic.Bool // ask the vm to yield so we can grab the mutex
	commands chan func(fromIoWait bool)
	done     chan struct{}

	exception *exceptionInfo

	lineBreaks []byte // pcs to break for lines
	instBreaks []byte // pcs to break for inst

	stepType  StepType
	stepGran  StepGran
	stepInst  int
	stepLine  int
	stepFrame int
}

type exceptionInfo struct {
	err    error
	trace  string
	frames []ErrFrame
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

	commands := make(chan func(bool))

	ec := &asm.ExecutionContext{
		Rng: rand.New(rand.NewSource(seed)),
		IoProvider: gaddis.IoAdapter{
			In: func() (string, error) {
				// TODO: an initial select that includes a short time delay, pause and unpause.
				for {
					select {
					case cmd, ok := <-commands:
						if !ok {
							// we're being asked to exit
							panic(sentinelIoExit{})
						}
						cmd(true)
					case in, ok := <-opts.Input:
						if !ok {
							return "", io.EOF
						}
						return in, nil
					}
				}
			},
			Out:     opts.Output, // TODO: should this also be a channel?
			WorkDir: ".",
		},
	}

	exec := source.Assembled.NewExecution(ec)

	lineBreaks := source.Breakpoints.ComputeLineBreaks(opts.LineBreaks)
	instBreaks := source.Breakpoints.ComputeInstBreaks(opts.InstBreaks)

	if opts.StopOnEntry && len(lineBreaks) > 0 {
		lineBreaks[0] = 1
	}

	return &Session{
		Opts:       opts,
		Host:       host,
		Source:     source,
		Exec:       exec,
		runState:   UNSTARTED,
		yield:      atomic.Bool{},
		commands:   commands,
		done:       make(chan struct{}),
		exception:  nil,
		lineBreaks: lineBreaks,
		instBreaks: instBreaks,
		stepType:   STEP_NONE,
		stepGran:   LineGran,
		stepInst:   0,
		stepLine:   0,
		stepFrame:  0,
	}
}
