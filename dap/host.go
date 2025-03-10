package dap

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/debug"
	api "github.com/google/go-dap"
)

// eventHost receives events from the debugger
type eventHost struct {
	// send function is concurrency safe
	send func(api.Message)

	// copy variables from session to avoid memory races
	source  api.Source
	lineOff int
	colOff  int
}

func (eh eventHost) Paused(reason string) {
	eh.send(&api.StoppedEvent{
		Event: *newEvent("stopped"),
		Body:  api.StoppedEventBody{Reason: reason, ThreadId: 1, AllThreadsStopped: true},
	})
}

func (eh eventHost) Exception(err error) {
	eh.send(&api.StoppedEvent{
		Event: *newEvent("stopped"),
		Body:  api.StoppedEventBody{Reason: "exception", Description: "exception", Text: err.Error(), ThreadId: 1, AllThreadsStopped: true},
	})
}

func (eh eventHost) Panicked(err error, pos []ast.Position, trace []string) {
	// just send stack traces over stderr and exit
	eh.send(&api.OutputEvent{
		Event: *newEvent("output"),
		Body: api.OutputEventBody{
			Category: "stderr",
			Output:   "error: " + err.Error() + "\n",
			Source:   &eh.source,
		},
	})
	for i := range pos {
		eh.send(&api.OutputEvent{
			Event: *newEvent("output"),
			Body: api.OutputEventBody{
				Category: "stderr",
				Output:   "\tin " + trace[i] + "\n",
				Source:   &eh.source,
				Line:     pos[i].Line + eh.lineOff,
				Column:   pos[i].Column + eh.colOff,
			},
		})
	}
	eh.Exited(1)
}

func (eh eventHost) Exited(code int) {
	eh.send(&api.ExitedEvent{
		Event: *newEvent("exited"),
		Body:  api.ExitedEventBody{ExitCode: code},
	})
	eh.Terminated()
}

func (eh eventHost) Terminated() {
	eh.send(&api.TerminatedEvent{
		Event: *newEvent("terminated"),
	})
}

var _ debug.EventHost = eventHost{}
