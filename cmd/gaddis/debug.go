package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dragonsinth/gaddis/asm"
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

	ch := connHandler{
		rw:        bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		sendQueue: make(chan dap.Message, 1024),
		bps:       map[string][]int{},
		dbgLog:    dbgLog,
	}

	if err := ch.Run(); err != nil {
		log.Println("Error:", err)
	}
}

func (h *connHandler) Run() error {
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

func (h *connHandler) handleRequest() error {
	request, err := dap.ReadProtocolMessage(h.rw.Reader)
	if err != nil {
		return err
	}
	h.dbgLog.Printf("Received request: %s", toJson(request))
	h.dispatchRequest(request)
	return nil
}

// dispatchRequest launches a new goroutine to process each request
// and send back events and responses.
func (h *connHandler) dispatchRequest(request dap.Message) {
	switch request := request.(type) {
	case *dap.InitializeRequest:
		h.onInitializeRequest(request)
	case *dap.LaunchRequest:
		h.onLaunchRequest(request)
	case *dap.AttachRequest:
		h.onAttachRequest(request)
	case *dap.DisconnectRequest:
		h.onDisconnectRequest(request)
	case *dap.TerminateRequest:
		h.onTerminateRequest(request)
	case *dap.RestartRequest:
		h.onRestartRequest(request)
	case *dap.SetBreakpointsRequest:
		h.onSetBreakpointsRequest(request)
	case *dap.SetFunctionBreakpointsRequest:
		h.onSetFunctionBreakpointsRequest(request)
	case *dap.SetExceptionBreakpointsRequest:
		h.onSetExceptionBreakpointsRequest(request)
	case *dap.ConfigurationDoneRequest:
		h.onConfigurationDoneRequest(request)
	case *dap.ContinueRequest:
		h.onContinueRequest(request)
	case *dap.NextRequest:
		h.onNextRequest(request)
	case *dap.StepInRequest:
		h.onStepInRequest(request)
	case *dap.StepOutRequest:
		h.onStepOutRequest(request)
	case *dap.StepBackRequest:
		h.onStepBackRequest(request)
	case *dap.ReverseContinueRequest:
		h.onReverseContinueRequest(request)
	case *dap.RestartFrameRequest:
		h.onRestartFrameRequest(request)
	case *dap.GotoRequest:
		h.onGotoRequest(request)
	case *dap.PauseRequest:
		h.onPauseRequest(request)
	case *dap.StackTraceRequest:
		h.onStackTraceRequest(request)
	case *dap.ScopesRequest:
		h.onScopesRequest(request)
	case *dap.VariablesRequest:
		h.onVariablesRequest(request)
	case *dap.SetVariableRequest:
		h.onSetVariableRequest(request)
	case *dap.SetExpressionRequest:
		h.onSetExpressionRequest(request)
	case *dap.SourceRequest:
		h.onSourceRequest(request)
	case *dap.ThreadsRequest:
		h.onThreadsRequest(request)
	case *dap.TerminateThreadsRequest:
		h.onTerminateThreadsRequest(request)
	case *dap.EvaluateRequest:
		h.onEvaluateRequest(request)
	case *dap.StepInTargetsRequest:
		h.onStepInTargetsRequest(request)
	case *dap.GotoTargetsRequest:
		h.onGotoTargetsRequest(request)
	case *dap.CompletionsRequest:
		h.onCompletionsRequest(request)
	case *dap.ExceptionInfoRequest:
		h.onExceptionInfoRequest(request)
	case *dap.LoadedSourcesRequest:
		h.onLoadedSourcesRequest(request)
	case *dap.DataBreakpointInfoRequest:
		h.onDataBreakpointInfoRequest(request)
	case *dap.SetDataBreakpointsRequest:
		h.onSetDataBreakpointsRequest(request)
	case *dap.ReadMemoryRequest:
		h.onReadMemoryRequest(request)
	case *dap.DisassembleRequest:
		h.onDisassembleRequest(request)
	case *dap.CancelRequest:
		h.onCancelRequest(request)
	case *dap.BreakpointLocationsRequest:
		h.onBreakpointLocationsRequest(request)
	default:
		h.send(newErrorResponse(request.GetSeq(), "unknown", "unknown command"))
	}
}

type eventHost struct {
	handler *connHandler
}

func (eh eventHost) Panicked(err error, si ast.Position, trace string) {
	eh.handler.send(&dap.OutputEvent{
		Event: *newEvent("output"),
		Body: dap.OutputEventBody{
			Category: "stderr",
			Output:   err.Error() + "\n",
			Source:   &eh.handler.source,
			Line:     si.Line + eh.handler.lineOff,
			Column:   si.Column + eh.handler.colOff,
		},
	})
	eh.handler.send(&dap.OutputEvent{
		Event: *newEvent("output"),
		Body: dap.OutputEventBody{
			Category: "stderr",
			Output:   trace,
			Source:   &eh.handler.source,
		},
	})
	eh.Exited(1)
}

func (eh eventHost) Breakpoint() {
	eh.handler.send(&dap.StoppedEvent{
		Event: *newEvent("stopped"),
		Body:  dap.StoppedEventBody{Reason: "breakpoint", ThreadId: 1, AllThreadsStopped: true},
	})
}

func (eh eventHost) Paused() {
	eh.handler.send(&dap.StoppedEvent{
		Event: *newEvent("stopped"),
		Body:  dap.StoppedEventBody{Reason: "pause", ThreadId: 1, AllThreadsStopped: true},
	})
}

func (eh eventHost) Exited(code int) {
	eh.handler.send(&dap.ExitedEvent{
		Event: *newEvent("exited"),
		Body:  dap.ExitedEventBody{ExitCode: code},
	})
	eh.Terminated()
}

func (eh eventHost) Terminated() {
	eh.handler.send(&dap.TerminatedEvent{
		Event: *newEvent("terminated"),
	})
}

var _ debug.EventHost = eventHost{}

// send lets the sender goroutine know via a channel that there is
// a message to be sent to client. This is called by per-request
// goroutines to send events and responses for each request and
// to notify of events triggered by the fake debugger.
func (h *connHandler) send(message dap.Message) {
	select {
	case h.sendQueue <- message:
	default:
		// just drop messages if the queue is that backed up
	}
}

// sendFromQueue is to be run in a separate goroutine to listen on a
// channel for messages to send back to the client. It will
// return once the channel is closed.
func (h *connHandler) sendFromQueue() {
	seq := 1
	for message := range h.sendQueue {
		switch m := message.(type) {
		case dap.ResponseMessage:
			m.GetResponse().Seq = seq
		case dap.EventMessage:
			m.GetEvent().Seq = seq
		default:
			panic(m)
		}
		seq++
		if err := dap.WriteProtocolMessage(h.rw.Writer, message); err != nil {
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

// -----------------------------------------------------------------------
// Very Fake Debugger
//

// The debugging session will keep track of how many breakpoints
// have been set. Once start-up is done (i.e. configurationDone
// request is processed), it will "stop" at each breakpoint one by
// one, and once there are no more, it will trigger a terminated event.
type connHandler struct {
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
	bps             map[string][]int
	lineOff, colOff int
	stopOnEntry     bool
	noDebug         bool

	sess *debug.Session
}

// -----------------------------------------------------------------------
// Request Handlers
//
// Below is a dummy implementation of the request handlers.
// They take no action, but just return dummy responses.
// A real debug adaptor would call the debugger methods here
// and use their results to populate each response.

func (h *connHandler) onInitializeRequest(request *dap.InitializeRequest) {
	if h.sess != nil {
		h.send(newErrorResponse(request.Seq, request.Command, "already launched"))
		return
	}

	if request.Arguments.LinesStartAt1 {
		h.lineOff = 1
	}
	if request.Arguments.ColumnsStartAt1 {
		h.colOff = 1
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
	h.send(e)
	h.send(response)
}

type launchArgs struct {
	Name        string `json:"name"`
	Program     string `json:"program"`
	WorkingDir  string `json:"workingDir"`
	StopOnEntry bool   `json:"stopOnEntry"`
	NoDebug     bool   `json:"noDebug"`
}

func (h *connHandler) onLaunchRequest(request *dap.LaunchRequest) {
	if h.sess != nil {
		h.send(newErrorResponse(request.Seq, request.Command, "already launched"))
		return
	}

	// This is where a real debug adaptor would check the soundness of the
	// arguments (e.g. program from launch.json) and then use them to launch the
	// debugger and attach to the program.
	var args launchArgs
	if err := fromJson(request.Arguments, &args); err != nil {
		h.send(newErrorResponse(request.Seq, request.Command, "could not parse launch request"))
		return
	}
	h.source.Name = args.Name
	h.source.Path = args.Program
	h.stopOnEntry = args.StopOnEntry
	h.noDebug = args.NoDebug

	compileErr := func(err ast.Error) {
		h.send(&dap.OutputEvent{
			Event: *newEvent("output"),
			Body: dap.OutputEventBody{
				Category: "stderr",
				Output:   err.Desc + "\n",
				Source:   &h.source,
				Line:     err.Start.Line + h.lineOff,
				Column:   err.Start.Column + h.colOff,
			},
		})
	}

	opts := debug.Opts{
		Stdin:      tryReadInput(args.Program),
		Stdout:     h.stdout,
		WorkingDir: args.WorkingDir,
	}
	sess, err := debug.New(h.source.Path, compileErr, eventHost{handler: h}, opts)
	if err != nil {
		h.send(newErrorResponse(request.Seq, request.Command, err.Error()))
		return
	}
	h.sess = sess

	response := &dap.LaunchResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	if h.noDebug {
		// launch immediately, otherwise wait for configuration done.
		h.sess.SetBreakpoints(nil)
		h.sess.Play()
	}
}

func (h *connHandler) onAttachRequest(request *dap.AttachRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "AttachRequest is not yet supported"))
}

func (h *connHandler) onDisconnectRequest(request *dap.DisconnectRequest) {
	if h.sess != nil {
		h.sess.Terminate()
	}
	response := &dap.DisconnectResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)
}

func (h *connHandler) onTerminateRequest(request *dap.TerminateRequest) {
	if h.sess != nil {
		h.sess.Terminate()
	}
	response := &dap.TerminateResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(&dap.TerminatedEvent{
		Event: *newEvent("terminated"),
	})
	h.send(response)
}

func (h *connHandler) onRestartRequest(request *dap.RestartRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}

	var wrap struct {
		Arguments launchArgs `json:"arguments"`
	}
	if err := fromJson(request.Arguments, &wrap); err != nil {
		h.send(newErrorResponse(request.Seq, request.Command, "could not parse restart request"))
		return
	}

	// surely these don't change right?
	args := &wrap.Arguments
	h.source.Name = args.Name
	h.source.Path = args.Program
	h.stopOnEntry = args.StopOnEntry
	h.noDebug = args.NoDebug

	response := &dap.RestartResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	h.sess.Halt()
	h.sess.Reset(debug.Opts{
		IsTest:     false,
		Stdin:      tryReadInput(args.Program),
		Stdout:     h.stdout,
		WorkingDir: args.WorkingDir,
	})

	if h.noDebug {
		h.sess.SetBreakpoints(nil)
	} else if h.stopOnEntry {
		h.sess.StopOnEntry()
	}
	h.sess.Play()
}

func (h *connHandler) onSetBreakpointsRequest(request *dap.SetBreakpointsRequest) {
	response := &dap.SetBreakpointsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	if h.noDebug {
		h.send(response)
		return
	}

	newBps := make([]int, len(request.Arguments.Breakpoints))
	for i, b := range request.Arguments.Breakpoints {
		newBps[i] = b.Line - h.lineOff
	}

	path := request.Arguments.Source.Path
	var verified bool
	if h.sess == nil {
		newBps, verified = debug.ValidateBreakpoints(path, newBps)
		h.bps[path] = newBps
	} else if h.sess.File == path {
		newBps = h.sess.SetBreakpoints(newBps)
		verified = true
	} else {
		// cannot accept the breakpoints
		// TODO: breakpoints in disassembly?
		verified = false
	}

	for _, bp := range newBps {
		response.Body.Breakpoints = append(response.Body.Breakpoints, dap.Breakpoint{
			Line:     bp + h.lineOff,
			Verified: verified,
		})
	}
	h.send(response)
}

func (h *connHandler) onSetFunctionBreakpointsRequest(request *dap.SetFunctionBreakpointsRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "SetFunctionBreakpointsRequest is not yet supported"))
}

func (h *connHandler) onSetExceptionBreakpointsRequest(request *dap.SetExceptionBreakpointsRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "SetExceptionBreakpointsRequest is not yet supported"))
}

func (h *connHandler) onConfigurationDoneRequest(request *dap.ConfigurationDoneRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}

	// This would be the place to check if the session was configured to
	// stop on entry and if that is the case, to issue a
	// stopped-on-breakpoint event. This being a mock implementation,
	// we "let" the program continue after sending a successful response.
	response := &dap.ConfigurationDoneResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	if h.noDebug {
		return // we should already be running
	}
	h.sess.SetBreakpoints(h.bps[h.sess.File])
	if h.stopOnEntry {
		h.sess.StopOnEntry()
	}
	h.sess.Play()
}

func (h *connHandler) onContinueRequest(request *dap.ContinueRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	h.sess.Play()
	response := &dap.ContinueResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)
}

func (h *connHandler) onNextRequest(request *dap.NextRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	response := &dap.NextResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	h.sess.Step(debug.STEP_NEXT)
	h.sess.Play()
}

func (h *connHandler) onStepInRequest(request *dap.StepInRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	response := &dap.StepBackResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	h.sess.Step(debug.STEP_IN)
	h.sess.Play()
}

func (h *connHandler) onStepOutRequest(request *dap.StepOutRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	response := &dap.StepOutResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	h.sess.Step(debug.STEP_OUT)
	h.sess.Play()
}

func (h *connHandler) onStepBackRequest(request *dap.StepBackRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "StepBackRequest is not yet supported"))
}

func (h *connHandler) onReverseContinueRequest(request *dap.ReverseContinueRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "ReverseContinueRequest is not yet supported"))
}

func (h *connHandler) onRestartFrameRequest(request *dap.RestartFrameRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "RestartFrameRequest is not yet supported"))
}

func (h *connHandler) onGotoRequest(request *dap.GotoRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "GotoRequest is not yet supported"))
}

func (h *connHandler) onPauseRequest(request *dap.PauseRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	h.sess.Pause()
	response := &dap.PauseResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)
}

func (h *connHandler) onStackTraceRequest(request *dap.StackTraceRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	response := &dap.StackTraceResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.sess.GetStackFrames(func(fr *asm.Frame, id int, inst asm.Inst) {
		pos := inst.GetSourceInfo().Start
		response.Body.StackFrames = append(response.Body.StackFrames, dap.StackFrame{
			Id:     id,
			Source: &h.source,
			Line:   pos.Line + h.lineOff,
			Column: h.colOff, // don't do columns yet... it's too weird
			Name:   fr.Scope.Desc(),
		})
	})
	response.Body.TotalFrames = len(response.Body.StackFrames)
	h.send(response)
}

func (h *connHandler) onScopesRequest(request *dap.ScopesRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	response := &dap.ScopesResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	frameId := request.Arguments.FrameId
	h.sess.GetStackFrames(func(fr *asm.Frame, id int, inst asm.Inst) {
		si := fr.Scope.SourceInfo
		varId := id * 1024
		if fr.Scope.IsGlobal {
			response.Body.Scopes = append(response.Body.Scopes, dap.Scope{
				Name:               "Globals",
				PresentationHint:   "globals",
				VariablesReference: varId,
				NamedVariables:     len(fr.Locals),
				IndexedVariables:   0, // should also be len?
				Expensive:          false,
				Source:             &h.source,
				Line:               si.Start.Line + h.lineOff,
				Column:             si.Start.Column + h.colOff,
				EndLine:            si.End.Line + h.lineOff,
				EndColumn:          si.End.Column + h.colOff,
			})
		} else if id == frameId {
			response.Body.Scopes = append(response.Body.Scopes, dap.Scope{
				Name:               "Locals",
				PresentationHint:   "locals",
				VariablesReference: varId,
				NamedVariables:     len(fr.Locals),
				IndexedVariables:   0, // should also be len?
				Expensive:          false,
				Source:             &h.source,
				Line:               si.Start.Line + h.lineOff,
				Column:             si.Start.Column + h.colOff,
				EndLine:            si.End.Line + h.lineOff,
				EndColumn:          si.End.Column + h.colOff,
			})
			response.Body.Scopes = append(response.Body.Scopes, dap.Scope{
				Name:               "Arguments",
				PresentationHint:   "arguments",
				VariablesReference: varId + 1,
				NamedVariables:     len(fr.Args),
				IndexedVariables:   0, // should also be len?
				Expensive:          false,
				Source:             &h.source,
				Line:               si.Start.Line + h.lineOff,
				Column:             si.Start.Column + h.colOff,
				EndLine:            si.End.Line + h.lineOff,
				EndColumn:          si.End.Column + h.colOff,
			})
		}
	})
	h.send(response)
}

func (h *connHandler) onVariablesRequest(request *dap.VariablesRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	varId := request.Arguments.VariablesReference
	frameId := varId / 1024
	varType := varId % 1024
	response := &dap.VariablesResponse{}
	response.Response = *newResponse(request.Seq, request.Command)

	addVar := func(val any, vd *ast.VarDecl) {
		response.Body.Variables = append(response.Body.Variables, dap.Variable{
			Name:               vd.Name,
			Value:              asm.DebugStringVal(val),
			Type:               vd.Type.String(),
			PresentationHint:   nil,
			EvaluateName:       vd.Name,
			VariablesReference: 0, // list/map
			NamedVariables:     0, // map
			IndexedVariables:   0, // list
			MemoryReference:    "",
		})
	}

	h.sess.GetStackFrames(func(fr *asm.Frame, id int, inst asm.Inst) {
		if id != frameId {
			return
		}
		switch varType {
		case 0: // locals
			nArgs := len(fr.Args)
			for i, vd := range fr.Scope.Params {
				addVar(fr.Locals[i], vd)
			}
			for i, vd := range fr.Scope.Locals {
				addVar(fr.Locals[i+nArgs], vd)
			}
		case 1: // args
			for i, vd := range fr.Scope.Params {
				addVar(fr.Args[i], vd)
			}
		}
	})
	h.send(response)
}

func (h *connHandler) onSetVariableRequest(request *dap.SetVariableRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "setVariableRequest is not yet supported"))
}

func (h *connHandler) onSetExpressionRequest(request *dap.SetExpressionRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "SetExpressionRequest is not yet supported"))
}

func (h *connHandler) onSourceRequest(request *dap.SourceRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "SourceRequest is not yet supported"))
}

func (h *connHandler) onThreadsRequest(request *dap.ThreadsRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	response := &dap.ThreadsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	response.Body = dap.ThreadsResponseBody{Threads: []dap.Thread{{Id: 1, Name: "main"}}}
	h.send(response)

}

func (h *connHandler) onTerminateThreadsRequest(request *dap.TerminateThreadsRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "TerminateRequest is not yet supported"))
}

func (h *connHandler) onEvaluateRequest(request *dap.EvaluateRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "EvaluateRequest is not yet supported"))
}

func (h *connHandler) onStepInTargetsRequest(request *dap.StepInTargetsRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "StepInTargetRequest is not yet supported"))
}

func (h *connHandler) onGotoTargetsRequest(request *dap.GotoTargetsRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "GotoTargetRequest is not yet supported"))
}

func (h *connHandler) onCompletionsRequest(request *dap.CompletionsRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "CompletionRequest is not yet supported"))
}

func (h *connHandler) onExceptionInfoRequest(request *dap.ExceptionInfoRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "ExceptionRequest is not yet supported"))
}

func (h *connHandler) onLoadedSourcesRequest(request *dap.LoadedSourcesRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "LoadedRequest is not yet supported"))
}

func (h *connHandler) onDataBreakpointInfoRequest(request *dap.DataBreakpointInfoRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "DataBreakpointInfoRequest is not yet supported"))
}

func (h *connHandler) onSetDataBreakpointsRequest(request *dap.SetDataBreakpointsRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "SetDataBreakpointsRequest is not yet supported"))
}

func (h *connHandler) onReadMemoryRequest(request *dap.ReadMemoryRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "ReadMemoryRequest is not yet supported"))
}

func (h *connHandler) onDisassembleRequest(request *dap.DisassembleRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "DisassembleRequest is not yet supported"))
}

func (h *connHandler) onCancelRequest(request *dap.CancelRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "CancelRequest is not yet supported"))
}

func (h *connHandler) onBreakpointLocationsRequest(request *dap.BreakpointLocationsRequest) {
	response := &dap.BreakpointLocationsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	if h.sess == nil {
		// anything goes I guess
		for line := request.Arguments.Line; line < request.Arguments.Line; line++ {
			response.Body.Breakpoints = append(response.Body.Breakpoints, dap.BreakpointLocation{
				Line:      line,
				Column:    h.colOff,
				EndLine:   line,
				EndColumn: h.colOff,
			})
		}
	} else {
		// filter down only to lines that are actually executable
		line := request.Arguments.Line
		if h.sess.MapSourceLine(line-1) >= 0 {
			response.Body.Breakpoints = append(response.Body.Breakpoints, dap.BreakpointLocation{
				Line:      line,
				Column:    h.colOff,
				EndLine:   line,
				EndColumn: h.colOff,
			})
		}
	}
	h.send(response)
}

func (h *connHandler) stdout(line string) {
	h.send(&dap.OutputEvent{
		Event: *newEvent("output"),
		Body: dap.OutputEventBody{
			Category: "stdout",
			Output:   line,
			Source:   &h.source,
		},
	})
}

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
		debugSession := connHandler{
			rw:        bufio.NewReadWriter(bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout)),
			sendQueue: make(chan dap.Message),
			bps:       map[string][]int{},
			dbgLog:    dbgLog,
		}

		return debugSession.Run()
	}
}

func tryReadInput(program string) []byte {
	buf, _ := os.ReadFile(program + ".in")
	return buf
}

func fromJson(buf json.RawMessage, obj any) error {
	dec := json.NewDecoder(bytes.NewReader(buf))
	return dec.Decode(obj)
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
