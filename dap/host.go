package dap

import (
	"bufio"
	"fmt"
	"github.com/dragonsinth/gaddis/debug"
	api "github.com/google/go-dap"
	"strings"
	"sync/atomic"
)

// eventHost receives events from the debugger
type eventHost struct {
	// send function is concurrency safe
	sendFunc func(api.Message)

	// for test mode
	remainingInput *bufio.Scanner
	capturedOutput strings.Builder
	wantOutput     string
	isTest         bool

	// copy variables from session to avoid memory races
	source  *api.Source
	lineOff int
	colOff  int

	suppressed int32
}

func (eh *eventHost) Paused(reason string) {
	eh.send(&api.StoppedEvent{
		Event: *newEvent("stopped"),
		Body:  api.StoppedEventBody{Reason: reason, ThreadId: 1, AllThreadsStopped: true},
	})
}

func (eh *eventHost) Exception(err error) {
	eh.send(&api.StoppedEvent{
		Event: *newEvent("stopped"),
		Body:  api.StoppedEventBody{Reason: "exception", Description: "exception", Text: err.Error(), ThreadId: 1, AllThreadsStopped: true},
	})
}

func (eh *eventHost) Panicked(err error, errFrames []debug.ErrFrame) {
	// just send stack traces over stderr and exit
	eh.send(&api.OutputEvent{
		Event: *newEvent("output"),
		Body: api.OutputEventBody{
			Category: "stderr",
			Output:   "error: " + err.Error() + "\n",
			Source:   eh.source,
		},
	})
	for _, fr := range errFrames {
		body := api.OutputEventBody{
			Category: "stderr",
			Output:   "in " + fr.Desc + "\n",
		}
		if fr.IsNative {
			body.Source = &api.Source{Path: fr.File} // use full path here not abstract
			body.Line = fr.Pos.Line + eh.lineOff
		} else {
			body.Source = eh.source
			body.Line = fr.Pos.Line + eh.lineOff
			body.Column = fr.Pos.Column + eh.colOff
		}
		eh.send(&api.OutputEvent{
			Event: *newEvent("output"),
			Body:  body,
		})
	}
	eh.Exited(1)
}

const failFmt = `
=== FAILED ===

-- WANT
%s--
++ GOT
%s++

=== FAILED ===
`

const inputUnreadFmt = `
=== FAILED ===

ERROR: %s

-- REMAINING INPUT
%s
=== FAILED ===
`

func (eh *eventHost) Exited(code int) {
	if eh.isTest && code == 0 {
		// check the output!
		gotOutput := eh.capturedOutput.String()
		if eh.wantOutput != gotOutput {
			code = 1
			msg := fmt.Sprintf(failFmt, eh.wantOutput, gotOutput)
			eh.send(&api.OutputEvent{
				Event: *newEvent("output"),
				Body:  api.OutputEventBody{Category: "stderr", Output: msg, Source: eh.source},
			})
		} else if rem, err := eh.drainStdin(); rem != "" || err != nil {
			code = 2
			failLine := "not all input was read"
			if err != nil {
				failLine = err.Error()
			}
			msg := fmt.Sprintf(inputUnreadFmt, failLine, rem)
			eh.send(&api.OutputEvent{
				Event: *newEvent("output"),
				Body:  api.OutputEventBody{Category: "stderr", Output: msg, Source: eh.source},
			})
		} else {
			eh.send(&api.OutputEvent{
				Event: *newEvent("output"),
				Body:  api.OutputEventBody{Category: "stdout", Output: "\n=== PASSED ===\n", Source: eh.source},
			})
		}
	}

	eh.send(&api.ExitedEvent{
		Event: *newEvent("exited"),
		Body:  api.ExitedEventBody{ExitCode: code},
	})
	eh.Terminated()
}

func (eh *eventHost) Terminated() {
	eh.send(&api.TerminatedEvent{
		Event: *newEvent("terminated"),
	})
}

func (eh *eventHost) SuppressAllEvents() {
	atomic.StoreInt32(&eh.suppressed, 1)
}

func (eh *eventHost) send(m api.Message) {
	if atomic.LoadInt32(&eh.suppressed) == 0 {
		eh.sendFunc(m)
	}
}

func (eh *eventHost) drainStdin() (string, error) {
	var sb strings.Builder
	for eh.remainingInput.Scan() {
		sb.WriteString(eh.remainingInput.Text())
		sb.WriteRune('\n')
	}
	return sb.String(), eh.remainingInput.Err()
}

var _ debug.EventHost = &eventHost{}
