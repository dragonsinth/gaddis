package dap

import (
	"bufio"
	"errors"
	"github.com/dragonsinth/gaddis/debug"
	api "github.com/google/go-dap"
	"io"
	"log"
	"sync"
)

// NewSession return a new DAP Session.
func NewSession(rw *bufio.ReadWriter, dbgLog *log.Logger) *Session {
	return &Session{
		rw:        rw,
		sendQueue: make(chan api.Message, 1024),
		bps:       map[string][]int{},
		dbgLog:    dbgLog,
	}
}

// Session manages the debug session and implements DAP.
type Session struct {
	// rw is used to read requests and write events/responses
	rw     *bufio.ReadWriter
	dbgLog *log.Logger

	// sendQueue is used to capture messages from multiple request
	// processing goroutines while writing them to the client connection
	// from a single goroutine via sendFromQueue. We must keep track of
	// the multiple channel senders with a wait group to make sure we do
	// not close this channel prematurely. Closing this channel will signal
	// the sendFromQueue goroutine that it can exit.
	sendQueue chan api.Message

	source          api.Source
	bps             map[string][]int
	lineOff, colOff int
	stopOnEntry     bool
	noDebug         bool

	sess *debug.Session
}

// Run runs the session for as long as it last.
func (h *Session) Run() error {
	var wg sync.WaitGroup
	defer wg.Wait()
	defer close(h.sendQueue)

	wg.Add(1)
	go func() {
		defer wg.Done()
		h.sendFromQueue()
	}()

	defer func() {
		if h.sess != nil {
			h.sess.Halt()
		}
	}()

	for {
		if err := h.handleRequest(); errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return err
		}
	}
}

func (h *Session) handleRequest() error {
	request, err := api.ReadProtocolMessage(h.rw.Reader)
	if err != nil {
		return err
	}
	h.dbgLog.Printf("Received request: %s", toJson(request))
	h.dispatchRequest(request)
	return nil
}

// dispatchRequest processes each request and sends back events and responses.
func (h *Session) dispatchRequest(request api.Message) {
	switch request := request.(type) {
	case *api.InitializeRequest:
		h.onInitializeRequest(request)
	case *api.LaunchRequest:
		h.onLaunchRequest(request)
	case *api.AttachRequest:
		h.onAttachRequest(request)
	case *api.DisconnectRequest:
		h.onDisconnectRequest(request)
	case *api.TerminateRequest:
		h.onTerminateRequest(request)
	case *api.RestartRequest:
		h.onRestartRequest(request)
	case *api.SetBreakpointsRequest:
		h.onSetBreakpointsRequest(request)
	case *api.SetFunctionBreakpointsRequest:
		h.onSetFunctionBreakpointsRequest(request)
	case *api.SetExceptionBreakpointsRequest:
		h.onSetExceptionBreakpointsRequest(request)
	case *api.ConfigurationDoneRequest:
		h.onConfigurationDoneRequest(request)
	case *api.ContinueRequest:
		h.onContinueRequest(request)
	case *api.NextRequest:
		h.onNextRequest(request)
	case *api.StepInRequest:
		h.onStepInRequest(request)
	case *api.StepOutRequest:
		h.onStepOutRequest(request)
	case *api.StepBackRequest:
		h.onStepBackRequest(request)
	case *api.ReverseContinueRequest:
		h.onReverseContinueRequest(request)
	case *api.RestartFrameRequest:
		h.onRestartFrameRequest(request)
	case *api.GotoRequest:
		h.onGotoRequest(request)
	case *api.PauseRequest:
		h.onPauseRequest(request)
	case *api.StackTraceRequest:
		h.onStackTraceRequest(request)
	case *api.ScopesRequest:
		h.onScopesRequest(request)
	case *api.VariablesRequest:
		h.onVariablesRequest(request)
	case *api.SetVariableRequest:
		h.onSetVariableRequest(request)
	case *api.SetExpressionRequest:
		h.onSetExpressionRequest(request)
	case *api.SourceRequest:
		h.onSourceRequest(request)
	case *api.ThreadsRequest:
		h.onThreadsRequest(request)
	case *api.TerminateThreadsRequest:
		h.onTerminateThreadsRequest(request)
	case *api.EvaluateRequest:
		h.onEvaluateRequest(request)
	case *api.StepInTargetsRequest:
		h.onStepInTargetsRequest(request)
	case *api.GotoTargetsRequest:
		h.onGotoTargetsRequest(request)
	case *api.CompletionsRequest:
		h.onCompletionsRequest(request)
	case *api.ExceptionInfoRequest:
		h.onExceptionInfoRequest(request)
	case *api.LoadedSourcesRequest:
		h.onLoadedSourcesRequest(request)
	case *api.DataBreakpointInfoRequest:
		h.onDataBreakpointInfoRequest(request)
	case *api.SetDataBreakpointsRequest:
		h.onSetDataBreakpointsRequest(request)
	case *api.ReadMemoryRequest:
		h.onReadMemoryRequest(request)
	case *api.DisassembleRequest:
		h.onDisassembleRequest(request)
	case *api.CancelRequest:
		h.onCancelRequest(request)
	case *api.BreakpointLocationsRequest:
		h.onBreakpointLocationsRequest(request)
	default:
		h.send(newErrorResponse(request.GetSeq(), "unknown", "unknown command"))
	}
}

// send lets the sender goroutine know via a channel that there is
// a message to be sent to client.
func (h *Session) send(message api.Message) {
	select {
	case h.sendQueue <- message:
	default:
		// just drop messages if the queue is that backed up
	}
}

// sendFromQueue run in a separate goroutine to listen on a
// channel for messages to send back to the client. It will
// return once the channel is closed or the outbound conn
// becomes unwritable.
func (h *Session) sendFromQueue() {
	seq := 1
	for message := range h.sendQueue {
		switch m := message.(type) {
		case api.ResponseMessage:
			m.GetResponse().Seq = seq
		case api.EventMessage:
			m.GetEvent().Seq = seq
		default:
			panic(m)
		}
		seq++
		if err := api.WriteProtocolMessage(h.rw.Writer, message); err != nil {
			log.Println("Error writing message:", err)
			return
		}
		h.dbgLog.Printf("Message sent\n%s", toJson(message))
		if err := h.rw.Flush(); err != nil {
			log.Println("Error writing message:", err)
			return
		}
	}
}

func (h *Session) stdout(line string) {
	h.send(&api.OutputEvent{
		Event: *newEvent("output"),
		Body: api.OutputEventBody{
			Category: "stdout",
			Output:   line,
			Source:   &h.source,
		},
	})
}
