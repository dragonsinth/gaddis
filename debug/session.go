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

	runState       runState
	yield          atomic.Bool // ask the vm to yield so we can grab the mutex
	commands       chan func(fromIoWait bool)
	commandsClosed atomic.Bool
	done           chan struct{}

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

	var inputDelegate func() (string, error) // fill in below
	exec := source.Assembled.NewExecution(&asm.ExecutionContext{
		Rng: rand.New(rand.NewSource(seed)),
		IoProvider: gaddis.IoAdapter{
			In: func() (string, error) {
				return inputDelegate()
			},
			Out:     opts.Output,
			WorkDir: ".", // TODO
		},
	})

	lineBreaks := source.Breakpoints.ComputeLineBreaks(opts.LineBreaks)
	instBreaks := source.Breakpoints.ComputeInstBreaks(opts.InstBreaks)

	if opts.StopOnEntry && len(lineBreaks) > 0 {
		lineBreaks[0] = 1
	}

	ds := &Session{
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

	inputDelegate = ds.inputAdapter
	return ds
}

func (ds *Session) inputAdapter() (string, error) {
	timeout := time.After(100 * time.Millisecond)
	for {

		select {
		case cmd, ok := <-ds.commands:
			if !ok {
				// we're being asked to exit
				panic(sentinelIoExit{})
			}
			cmd(true)
		case in, ok := <-ds.Opts.Input:
			if !ok {
				return "", io.EOF
			}
			if timeout == nil && ds.runState == RUN {
				ds.Host.Continued()
			}
			return in, nil
		case <-timeout:
			timeout = nil // only timeout once
			ds.Host.Paused("i/o wait")
		}
	}
}
