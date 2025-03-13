package dap

import (
	"github.com/dragonsinth/gaddis/debug"
	api "github.com/google/go-dap"
	"sync/atomic"
)

// eventHost receives events from the debugger
type eventHost struct {
	// send function is concurrency safe
	sendFunc func(api.Message)

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

func (eh *eventHost) Exited(code int) {
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

var _ debug.EventHost = &eventHost{}
