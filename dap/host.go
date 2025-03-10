package dap

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/debug"
	api "github.com/google/go-dap"
)

// eventHost receives events from the debugger
type eventHost struct {
	handler *Session
	noDebug bool
}

func (eh eventHost) Paused(reason string) {
	eh.handler.send(&api.StoppedEvent{
		Event: *newEvent("stopped"),
		Body:  api.StoppedEventBody{Reason: reason, ThreadId: 1, AllThreadsStopped: true},
	})
}

func (eh eventHost) BreakOnException() bool {
	return !eh.noDebug
}

func (eh eventHost) Exception(err error) {
	eh.handler.send(&api.StoppedEvent{
		Event: *newEvent("stopped"),
		Body:  api.StoppedEventBody{Reason: "exception", Description: "exception", Text: err.Error(), ThreadId: 1, AllThreadsStopped: true},
	})
}

func (eh eventHost) Panicked(err error, pos []ast.Position, trace []string) {
	// just send stack traces over stderr and exit
	eh.handler.send(&api.OutputEvent{
		Event: *newEvent("output"),
		Body: api.OutputEventBody{
			Category: "stderr",
			Output:   "error: " + err.Error() + "\n",
			Source:   &eh.handler.source,
		},
	})
	for i := range pos {
		eh.handler.send(&api.OutputEvent{
			Event: *newEvent("output"),
			Body: api.OutputEventBody{
				Category: "stderr",
				Output:   "\tin " + trace[i] + "\n",
				Source:   &eh.handler.source,
				Line:     pos[i].Line + eh.handler.lineOff,
				Column:   pos[i].Column + eh.handler.colOff,
			},
		})
	}
	eh.Exited(1)
}

func (eh eventHost) Exited(code int) {
	eh.handler.send(&api.ExitedEvent{
		Event: *newEvent("exited"),
		Body:  api.ExitedEventBody{ExitCode: code},
	})
	eh.Terminated()
}

func (eh eventHost) Terminated() {
	eh.handler.send(&api.TerminatedEvent{
		Event: *newEvent("terminated"),
	})
}

var _ debug.EventHost = eventHost{}
