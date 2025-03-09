package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dragonsinth/gaddis"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/debug"
	"github.com/google/go-dap"
	"io"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
)

// server starts a server that listens on a specified port
// and blocks indefinitely. This server can accept multiple
// client connections at the same time.
func debugServer(port int, dbgLog *log.Logger) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return err
	}
	defer func() {
		_ = listener.Close()
	}()
	log.Println("Started server at", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection failed:", err)
			continue
		}
		log.Println("Accepted connection from", conn.RemoteAddr())
		// Handle multiple client connections concurrently
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("panic:", r)
					buf := make([]byte, 1<<16)
					runtime.Stack(buf, false)
					log.Println(string(buf))
				}
			}()
			handleConnection(conn, dbgLog)
		}()
	}
}

// handleConnection handles a connection from a single client.
// It reads and decodes the incoming data and dispatches it
// to per-request processing goroutines. It also launches the
// sender goroutine to send resulting messages over the connection
// back to the client.
func handleConnection(conn net.Conn, dbgLog *log.Logger) {
	defer func() {
		log.Println("Closing connection from", conn.RemoteAddr())
		_ = conn.Close()
	}()

	debugSession := fakeDebugSession{
		rw:        bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		sendQueue: make(chan dap.Message),
		bps:       map[string][]int{},
		dbgLog:    dbgLog,
	}

	if err := debugSession.Run(); err != nil {
		log.Println("Error:", err)
	}
}

func (ds *fakeDebugSession) Run() error {
	var wg sync.WaitGroup
	defer wg.Wait()
	defer close(ds.sendQueue)

	wg.Add(1)
	go func() {
		defer wg.Done()
		ds.sendFromQueue()
	}()

	for {
		if err := ds.handleRequest(); errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return err
		}
	}
}

func (ds *fakeDebugSession) handleRequest() error {
	request, err := dap.ReadProtocolMessage(ds.rw.Reader)
	if err != nil {
		return err
	}
	ds.dbgLog.Printf("Received request: %s", toJson(request))
	ds.dispatchRequest(request)
	return nil
}

// dispatchRequest launches a new goroutine to process each request
// and send back events and responses.
func (ds *fakeDebugSession) dispatchRequest(request dap.Message) {
	switch request := request.(type) {
	case *dap.InitializeRequest:
		ds.onInitializeRequest(request)
	case *dap.LaunchRequest:
		ds.onLaunchRequest(request)
	case *dap.AttachRequest:
		ds.onAttachRequest(request)
	case *dap.DisconnectRequest:
		ds.onDisconnectRequest(request)
	case *dap.TerminateRequest:
		ds.onTerminateRequest(request)
	case *dap.RestartRequest:
		ds.onRestartRequest(request)
	case *dap.SetBreakpointsRequest:
		ds.onSetBreakpointsRequest(request)
	case *dap.SetFunctionBreakpointsRequest:
		ds.onSetFunctionBreakpointsRequest(request)
	case *dap.SetExceptionBreakpointsRequest:
		ds.onSetExceptionBreakpointsRequest(request)
	case *dap.ConfigurationDoneRequest:
		ds.onConfigurationDoneRequest(request)
	case *dap.ContinueRequest:
		ds.onContinueRequest(request)
	case *dap.NextRequest:
		ds.onNextRequest(request)
	case *dap.StepInRequest:
		ds.onStepInRequest(request)
	case *dap.StepOutRequest:
		ds.onStepOutRequest(request)
	case *dap.StepBackRequest:
		ds.onStepBackRequest(request)
	case *dap.ReverseContinueRequest:
		ds.onReverseContinueRequest(request)
	case *dap.RestartFrameRequest:
		ds.onRestartFrameRequest(request)
	case *dap.GotoRequest:
		ds.onGotoRequest(request)
	case *dap.PauseRequest:
		ds.onPauseRequest(request)
	case *dap.StackTraceRequest:
		ds.onStackTraceRequest(request)
	case *dap.ScopesRequest:
		ds.onScopesRequest(request)
	case *dap.VariablesRequest:
		ds.onVariablesRequest(request)
	case *dap.SetVariableRequest:
		ds.onSetVariableRequest(request)
	case *dap.SetExpressionRequest:
		ds.onSetExpressionRequest(request)
	case *dap.SourceRequest:
		ds.onSourceRequest(request)
	case *dap.ThreadsRequest:
		ds.onThreadsRequest(request)
	case *dap.TerminateThreadsRequest:
		ds.onTerminateThreadsRequest(request)
	case *dap.EvaluateRequest:
		ds.onEvaluateRequest(request)
	case *dap.StepInTargetsRequest:
		ds.onStepInTargetsRequest(request)
	case *dap.GotoTargetsRequest:
		ds.onGotoTargetsRequest(request)
	case *dap.CompletionsRequest:
		ds.onCompletionsRequest(request)
	case *dap.ExceptionInfoRequest:
		ds.onExceptionInfoRequest(request)
	case *dap.LoadedSourcesRequest:
		ds.onLoadedSourcesRequest(request)
	case *dap.DataBreakpointInfoRequest:
		ds.onDataBreakpointInfoRequest(request)
	case *dap.SetDataBreakpointsRequest:
		ds.onSetDataBreakpointsRequest(request)
	case *dap.ReadMemoryRequest:
		ds.onReadMemoryRequest(request)
	case *dap.DisassembleRequest:
		ds.onDisassembleRequest(request)
	case *dap.CancelRequest:
		ds.onCancelRequest(request)
	case *dap.BreakpointLocationsRequest:
		ds.onBreakpointLocationsRequest(request)
	default:
		ds.send(newErrorResponse(request.GetSeq(), "unknown", "unknown command"))
	}
}

// send lets the sender goroutine know via a channel that there is
// a message to be sent to client. This is called by per-request
// goroutines to send events and responses for each request and
// to notify of events triggered by the fake debugger.
func (ds *fakeDebugSession) send(message dap.Message) {
	ds.sendQueue <- message
}

// sendFromQueue is to be run in a separate goroutine to listen on a
// channel for messages to send back to the client. It will
// return once the channel is closed.
func (ds *fakeDebugSession) sendFromQueue() {
	seq := 1
	for message := range ds.sendQueue {
		switch m := message.(type) {
		case dap.ResponseMessage:
			m.GetResponse().Seq = seq
		case dap.EventMessage:
			m.GetEvent().Seq = seq
		default:
			panic(m)
		}
		seq++
		if err := dap.WriteProtocolMessage(ds.rw.Writer, message); err != nil {
			log.Println("Error writing message:", err)
			return
		}
		ds.dbgLog.Printf("Message sent\n%s", toJson(message))
		if err := ds.rw.Flush(); err != nil {
			log.Println("Error writing message:", err)
			return
		}
	}
}

// -----------------------------------------------------------------------
// Very Fake Debugger
//

// The debugging session will keep track of how many breakpoints
// have been set. Once start-up is done (i.e. configurationDone
// request is processed), it will "stop" at each breakpoint one by
// one, and once there are no more, it will trigger a terminated event.
type fakeDebugSession struct {
	// rw is used to read requests and write events/responses
	rw     *bufio.ReadWriter
	dbgLog *log.Logger

	// sendQueue is used to capture messages from multiple request
	// processing goroutines while writing them to the client connection
	// from a single goroutine via sendFromQueue. We must keep track of
	// the multiple channel senders with a wait group to make sure we do
	// not close this channel prematurely. Closing this channel will signal
	// the sendFromQueue goroutine that it can exit.
	sendQueue chan dap.Message

	source          dap.Source
	line            int
	bps             map[string][]int
	lineOff, colOff int
	stopOnEntry     bool
}

// doContinue allows fake program execution to continue when the program
// is started or unpaused. It simulates events from the debug session
// by "stopping" on a breakpoint or terminating if there are no more
// breakpoints. Safe to use concurrently.
func (ds *fakeDebugSession) doContinue() {
	if ds.line < 0 && ds.stopOnEntry {
		e := &dap.ThreadEvent{Event: *newEvent("thread"), Body: dap.ThreadEventBody{Reason: "started", ThreadId: 1}}
		ds.send(e)
		ds.line = 0
		if ds.stopOnEntry {
			ds.send(&dap.StoppedEvent{
				Event: *newEvent("stopped"),
				Body:  dap.StoppedEventBody{Reason: "breakpoint", ThreadId: 1, AllThreadsStopped: true},
			})
			return
		}
	}

	ds.line++
	bps := ds.bps[ds.source.Path]
	for _, bp := range bps {
		if ds.line < bp {
			// stop here
			ds.line = bp
			ds.send(&dap.StoppedEvent{
				Event: *newEvent("stopped"),
				Body:  dap.StoppedEventBody{Reason: "breakpoint", ThreadId: 1, AllThreadsStopped: true},
			})
			return
		}
	}

	// no breakpoints hit
	ds.send(&dap.TerminatedEvent{
		Event: *newEvent("terminated"),
	})
}

// -----------------------------------------------------------------------
// Request Handlers
//
// Below is a dummy implementation of the request handlers.
// They take no action, but just return dummy responses.
// A real debug adaptor would call the debugger methods here
// and use their results to populate each response.

func (ds *fakeDebugSession) onInitializeRequest(request *dap.InitializeRequest) {
	if request.Arguments.LinesStartAt1 {
		ds.lineOff = 1
	}
	if request.Arguments.ColumnsStartAt1 {
		ds.colOff = 1
	}
	response := &dap.InitializeResponse{
		Response: *newResponse(request.Seq, request.Command),
		Body: dap.Capabilities{
			SupportsConfigurationDoneRequest:   true,
			SupportsSetVariable:                true,
			SupportsRestartFrame:               true,
			SupportsStepInTargetsRequest:       false, // what is this
			SupportsRestartRequest:             true,
			SupportTerminateDebuggee:           true,
			SupportSuspendDebuggee:             true,
			SupportsLoadedSourcesRequest:       true,
			SupportsTerminateRequest:           true,
			SupportsDisassembleRequest:         true,
			SupportsBreakpointLocationsRequest: true,
			SupportsSteppingGranularity:        false, // support later
		},
	}
	e := &dap.InitializedEvent{Event: *newEvent("initialized")}
	ds.send(e)
	ds.send(response)
}

type launchArgs struct {
	Name        string `json:"name"`
	Program     string `json:"program"`
	StopOnEntry bool   `json:"stopOnEntry"`
	NoDebug     bool   `json:"noDebug"`
}

func (ds *fakeDebugSession) onLaunchRequest(request *dap.LaunchRequest) {
	// This is where a real debug adaptor would check the soundness of the
	// arguments (e.g. program from launch.json) and then use them to launch the
	// debugger and attach to the program.
	var args launchArgs
	if err := json.Unmarshal(request.Arguments, &args); err != nil {
		ds.send(newErrorResponse(request.Seq, request.Command, "could not parse launch request"))
		return
	}
	ds.source.Name = args.Name
	ds.source.Path = args.Program
	ds.stopOnEntry = args.StopOnEntry
	ds.line = -1

	compileErr := func(err ast.Error) {
		ds.send(&dap.OutputEvent{
			Event: *newEvent("output"),
			Body: dap.OutputEventBody{
				Category: "stderr",
				Output:   err.Desc,
				Source:   &ds.source,
				Line:     err.Start.Line + ds.lineOff,
				Column:   err.Start.Column + ds.colOff,
			},
		})
	}
	opts := debug.DebugOpts{
		Stdout: ds.makeStream("stdout"),
		Stderr: ds.makeStream("stderr"),
	}
	sess, err := debug.New(ds.source.Path, compileErr, ds, opts)
	if err != nil {
		ds.send(newErrorResponse(request.Seq, request.Command, err.Error()))
		return
	}

	if args.NoDebug {
		// just run
		go func() {
			if err := sess.Exec.Run(); err != nil {
				// TODO: map sess.Exec.PC -> source line
				si := sess.Exec.Code[sess.Exec.PC].GetSourceInfo().Start
				ds.send(&dap.OutputEvent{
					Event: *newEvent("output"),
					Body: dap.OutputEventBody{
						Category: "stderr",
						Output:   err.Error(),
						Source:   &ds.source,
						Line:     si.Line + ds.lineOff,
						Column:   si.Column + ds.colOff,
					},
				})
			}
			ds.send(&dap.TerminatedEvent{
				Event: *newEvent("terminated"),
			})
		}()
		return
	}

	response := &dap.LaunchResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	ds.send(response)
}

func (ds *fakeDebugSession) onAttachRequest(request *dap.AttachRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "AttachRequest is not yet supported"))
}

func (ds *fakeDebugSession) onDisconnectRequest(request *dap.DisconnectRequest) {
	response := &dap.DisconnectResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	ds.send(response)
}

func (ds *fakeDebugSession) onTerminateRequest(request *dap.TerminateRequest) {
	response := &dap.TerminateResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	ds.send(&dap.TerminatedEvent{
		Event: *newEvent("terminated"),
	})
	ds.send(response)
}

func (ds *fakeDebugSession) onRestartRequest(request *dap.RestartRequest) {
	var wrap struct {
		Arguments launchArgs `json:"arguments"`
	}
	args := &wrap.Arguments
	if err := json.Unmarshal(request.Arguments, &args); err != nil {
		ds.send(newErrorResponse(request.Seq, request.Command, "could not parse restart request"))
		return
	}
	ds.source.Name = args.Name
	ds.source.Path = args.Program
	ds.stopOnEntry = args.StopOnEntry
	ds.line = -1
	response := &dap.RestartResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	ds.send(response)
	if args.NoDebug {
		// just run
		ds.send(&dap.TerminatedEvent{
			Event: *newEvent("terminated"),
		})
	} else {
		ds.doContinue()
	}
}

func (ds *fakeDebugSession) onSetBreakpointsRequest(request *dap.SetBreakpointsRequest) {
	response := &dap.SetBreakpointsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	response.Body.Breakpoints = make([]dap.Breakpoint, len(request.Arguments.Breakpoints))
	newBps := make([]int, len(request.Arguments.Breakpoints))
	for i, b := range request.Arguments.Breakpoints {
		response.Body.Breakpoints[i].Line = b.Line
		response.Body.Breakpoints[i].Verified = true
		newBps[i] = b.Line - ds.lineOff
	}
	ds.bps[request.Arguments.Source.Path] = newBps
	ds.send(response)
}

func (ds *fakeDebugSession) onSetFunctionBreakpointsRequest(request *dap.SetFunctionBreakpointsRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "SetFunctionBreakpointsRequest is not yet supported"))
}

func (ds *fakeDebugSession) onSetExceptionBreakpointsRequest(request *dap.SetExceptionBreakpointsRequest) {
	response := &dap.SetExceptionBreakpointsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	ds.send(response)
}

func (ds *fakeDebugSession) onConfigurationDoneRequest(request *dap.ConfigurationDoneRequest) {
	// This would be the place to check if the session was configured to
	// stop on entry and if that is the case, to issue a
	// stopped-on-breakpoint event. This being a mock implementation,
	// we "let" the program continue after sending a successful response.
	response := &dap.ConfigurationDoneResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	ds.send(response)
	ds.doContinue()
}

func (ds *fakeDebugSession) onContinueRequest(request *dap.ContinueRequest) {
	response := &dap.ContinueResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	ds.send(response)
	ds.doContinue()
}

func (ds *fakeDebugSession) onNextRequest(request *dap.NextRequest) {
	response := &dap.NextResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	ds.send(response)
	ds.line++
	ds.send(&dap.StoppedEvent{
		Event: *newEvent("stopped"),
		Body:  dap.StoppedEventBody{Reason: "breakpoint", ThreadId: 1, AllThreadsStopped: true},
	})
	ds.send(response)
}

func (ds *fakeDebugSession) onStepInRequest(request *dap.StepInRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "StepInRequest is not yet supported"))
}

func (ds *fakeDebugSession) onStepOutRequest(request *dap.StepOutRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "StepOutRequest is not yet supported"))
}

func (ds *fakeDebugSession) onStepBackRequest(request *dap.StepBackRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "StepBackRequest is not yet supported"))
}

func (ds *fakeDebugSession) onReverseContinueRequest(request *dap.ReverseContinueRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "ReverseContinueRequest is not yet supported"))
}

func (ds *fakeDebugSession) onRestartFrameRequest(request *dap.RestartFrameRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "RestartFrameRequest is not yet supported"))
}

func (ds *fakeDebugSession) onGotoRequest(request *dap.GotoRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "GotoRequest is not yet supported"))
}

func (ds *fakeDebugSession) onPauseRequest(request *dap.PauseRequest) {
	response := &dap.PauseResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	ds.send(response)
}

func (ds *fakeDebugSession) onStackTraceRequest(request *dap.StackTraceRequest) {
	response := &dap.StackTraceResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	response.Body = dap.StackTraceResponseBody{
		StackFrames: []dap.StackFrame{
			{
				Id:     1000,
				Source: &ds.source,
				Line:   ds.line + ds.lineOff,
				Column: ds.colOff,
				Name:   "main()",
			},
		},
		TotalFrames: 1,
	}
	ds.send(response)
}

func (ds *fakeDebugSession) onScopesRequest(request *dap.ScopesRequest) {
	response := &dap.ScopesResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	response.Body = dap.ScopesResponseBody{
		Scopes: []dap.Scope{
			{Name: "Local", VariablesReference: 1000, Expensive: false},
			{Name: "Global", VariablesReference: 1001, Expensive: false},
		},
	}
	ds.send(response)
}

func (ds *fakeDebugSession) onVariablesRequest(request *dap.VariablesRequest) {
	response := &dap.VariablesResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	response.Body = dap.VariablesResponseBody{
		Variables: []dap.Variable{{Name: "i", Value: "18434528", EvaluateName: "i", VariablesReference: 0}},
	}
	ds.send(response)
}

func (ds *fakeDebugSession) onSetVariableRequest(request *dap.SetVariableRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "setVariableRequest is not yet supported"))
}

func (ds *fakeDebugSession) onSetExpressionRequest(request *dap.SetExpressionRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "SetExpressionRequest is not yet supported"))
}

func (ds *fakeDebugSession) onSourceRequest(request *dap.SourceRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "SourceRequest is not yet supported"))
}

func (ds *fakeDebugSession) onThreadsRequest(request *dap.ThreadsRequest) {
	response := &dap.ThreadsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	response.Body = dap.ThreadsResponseBody{Threads: []dap.Thread{{Id: 1, Name: "main"}}}
	ds.send(response)

}

func (ds *fakeDebugSession) onTerminateThreadsRequest(request *dap.TerminateThreadsRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "TerminateRequest is not yet supported"))
}

func (ds *fakeDebugSession) onEvaluateRequest(request *dap.EvaluateRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "EvaluateRequest is not yet supported"))
}

func (ds *fakeDebugSession) onStepInTargetsRequest(request *dap.StepInTargetsRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "StepInTargetRequest is not yet supported"))
}

func (ds *fakeDebugSession) onGotoTargetsRequest(request *dap.GotoTargetsRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "GotoTargetRequest is not yet supported"))
}

func (ds *fakeDebugSession) onCompletionsRequest(request *dap.CompletionsRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "CompletionRequest is not yet supported"))
}

func (ds *fakeDebugSession) onExceptionInfoRequest(request *dap.ExceptionInfoRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "ExceptionRequest is not yet supported"))
}

func (ds *fakeDebugSession) onLoadedSourcesRequest(request *dap.LoadedSourcesRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "LoadedRequest is not yet supported"))
}

func (ds *fakeDebugSession) onDataBreakpointInfoRequest(request *dap.DataBreakpointInfoRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "DataBreakpointInfoRequest is not yet supported"))
}

func (ds *fakeDebugSession) onSetDataBreakpointsRequest(request *dap.SetDataBreakpointsRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "SetDataBreakpointsRequest is not yet supported"))
}

func (ds *fakeDebugSession) onReadMemoryRequest(request *dap.ReadMemoryRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "ReadMemoryRequest is not yet supported"))
}

func (ds *fakeDebugSession) onDisassembleRequest(request *dap.DisassembleRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "DisassembleRequest is not yet supported"))
}

func (ds *fakeDebugSession) onCancelRequest(request *dap.CancelRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "CancelRequest is not yet supported"))
}

func (ds *fakeDebugSession) onBreakpointLocationsRequest(request *dap.BreakpointLocationsRequest) {
	response := &dap.BreakpointLocationsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	response.Body = dap.BreakpointLocationsResponseBody{
		Breakpoints: []dap.BreakpointLocation{
			{
				Line:      request.Arguments.Line,
				Column:    ds.colOff,
				EndLine:   request.Arguments.EndLine,
				EndColumn: ds.colOff,
			},
		},
	}
	ds.send(response)
}

func (ds *fakeDebugSession) makeStream(typ string) gaddis.SyncWriter {
	return &bufferedSyncWriter{
		typ: typ,
		ds:  ds,
	}
}

type bufferedSyncWriter struct {
	buffer bytes.Buffer
	typ    string
	ds     *fakeDebugSession
}

func (b *bufferedSyncWriter) Write(p []byte) (n int, err error) {
	return b.buffer.Write(p)
}

func (b *bufferedSyncWriter) Sync() error {
	output := b.buffer.String()
	b.buffer.Reset()
	b.ds.send(&dap.OutputEvent{
		Event: *newEvent("output"),
		Body: dap.OutputEventBody{
			Category: "stderr",
			Output:   output,
			Source:   &b.ds.source,
		},
	})
	return nil
}

var _ gaddis.SyncWriter = &bufferedSyncWriter{}

func newEvent(event string) *dap.Event {
	return &dap.Event{
		ProtocolMessage: dap.ProtocolMessage{
			Seq:  0,
			Type: "event",
		},
		Event: event,
	}
}

func newResponse(requestSeq int, command string) *dap.Response {
	return &dap.Response{
		ProtocolMessage: dap.ProtocolMessage{
			Seq:  0,
			Type: "response",
		},
		Command:    command,
		RequestSeq: requestSeq,
		Success:    true,
	}
}

func newErrorResponse(requestSeq int, command string, message string) *dap.ErrorResponse {
	er := &dap.ErrorResponse{}
	er.Response = *newResponse(requestSeq, command)
	er.Success = false
	er.Message = "unsupported"
	er.Body.Error = &dap.ErrorMessage{Format: message}
	return er
}

func debugCmd(port int, verbose bool) error {
	dbgLog := log.New(io.Discard, "", log.LstdFlags)

	if port < 0 {
		// we have to run on stdout, so create a log file.
		_ = os.Chdir(os.Getenv("PWD"))
		logFile, err := os.OpenFile("gaddis-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err == nil {
			log.SetOutput(logFile)
			defer func() {
				_ = logFile.Close()
			}()
		}
		if verbose {
			dbgLog.SetOutput(logFile)
		}
		log.SetOutput(logFile)
	} else {
		// setup normal stdout/stderr logging
		if verbose {
			dbgLog.SetOutput(os.Stdout)
		}
		log.SetOutput(os.Stderr)

		// if running as a server within vscode, don't emit timestamps (IDE will do this).
		if os.Getenv("VSCODE_CLI") != "" {
			dbgLog.SetFlags(0)
			log.SetFlags(0)
		}
	}

	if verbose {
		dbgLog.Println(os.Getwd())
		dbgLog.Println(os.Args)
		for _, ev := range os.Environ() {
			dbgLog.Println(ev)
		}
	}

	if port >= 0 {
		return debugServer(port, dbgLog)
	} else {
		debugSession := fakeDebugSession{
			rw:        bufio.NewReadWriter(bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout)),
			sendQueue: make(chan dap.Message),
			bps:       map[string][]int{},
			dbgLog:    dbgLog,
		}

		return debugSession.Run()
	}
}

func toJson(obj any) string {
	var buf strings.Builder
	e := json.NewEncoder(&buf)
	e.SetIndent("", "  ")
	e.SetEscapeHTML(false)
	if err := e.Encode(obj); err != nil {
		panic(err)
	}
	return buf.String()
}
