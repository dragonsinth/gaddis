package dap

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/debug"
	"github.com/dragonsinth/gaddis/parse"
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

// -----------------------------------------------------------------------
// Request Handlers
// -----------------------------------------------------------------------

func (h *Session) onInitializeRequest(request *api.InitializeRequest) {
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
	response := &api.InitializeResponse{
		Response: *newResponse(request.Seq, request.Command),
		Body: api.Capabilities{
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
	e := &api.InitializedEvent{Event: *newEvent("initialized")}
	h.send(e)
	h.send(response)
}

func (h *Session) onLaunchRequest(request *api.LaunchRequest) {
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
		h.send(&api.OutputEvent{
			Event: *newEvent("output"),
			Body: api.OutputEventBody{
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
	sess, err := debug.New(h.source.Path, compileErr, eventHost{handler: h, noDebug: h.noDebug}, opts)
	if err != nil {
		h.send(newErrorResponse(request.Seq, request.Command, err.Error()))
		return
	}
	h.sess = sess

	response := &api.LaunchResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	if h.noDebug {
		// launch immediately, otherwise wait for configuration done.
		h.sess.SetBreakpoints(nil)
		h.sess.Play()
	}
}

func (h *Session) onAttachRequest(request *api.AttachRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "AttachRequest is not yet supported"))
}

func (h *Session) onDisconnectRequest(request *api.DisconnectRequest) {
	if h.sess != nil {
		h.sess.Terminate()
	}
	response := &api.DisconnectResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)
}

func (h *Session) onTerminateRequest(request *api.TerminateRequest) {
	if h.sess != nil {
		h.sess.Terminate()
	}
	response := &api.TerminateResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(&api.TerminatedEvent{
		Event: *newEvent("terminated"),
	})
	h.send(response)
}

func (h *Session) onRestartRequest(request *api.RestartRequest) {
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

	response := &api.RestartResponse{}
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

func (h *Session) onSetBreakpointsRequest(request *api.SetBreakpointsRequest) {
	source := request.Arguments.Source

	response := &api.SetBreakpointsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	if h.noDebug {
		h.send(response)
		return
	}

	var found map[int]bool
	if h.sess != nil && h.sess.File == source.Path {
		found = h.sess.ValidLineBreaks
	} else if h.sess == nil {
		found = debug.FindBreakpoints(source.Path)
	} else {
		// cannot accept the breakpoints
		found = nil
	}

	var bps []int
	for _, bp := range request.Arguments.Breakpoints {
		var msg string
		srcLine := bp.Line - h.lineOff
		verified := found[srcLine]
		if verified {
			bps = append(bps, srcLine)
		} else {
			msg = "failed"
		}
		response.Body.Breakpoints = append(response.Body.Breakpoints, api.Breakpoint{
			Verified: verified,
			Message:  msg,
			Source:   &source,
			Line:     bp.Line,
			Column:   h.colOff,
		})
	}

	if h.sess != nil && h.sess.File == source.Path {
		h.sess.SetBreakpoints(bps)
	} else if h.sess == nil {
		h.bps[source.Path] = bps
	}
	h.send(response)
}

func (h *Session) onSetFunctionBreakpointsRequest(request *api.SetFunctionBreakpointsRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "SetFunctionBreakpointsRequest is not yet supported"))
}

func (h *Session) onSetExceptionBreakpointsRequest(request *api.SetExceptionBreakpointsRequest) {
	response := &api.SetExceptionBreakpointsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)
}

func (h *Session) onConfigurationDoneRequest(request *api.ConfigurationDoneRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}

	// This would be the place to check if the session was configured to
	// stop on entry and if that is the case, to issue a
	// stopped-on-breakpoint event. This being a mock implementation,
	// we "let" the program continue after sending a successful response.
	response := &api.ConfigurationDoneResponse{}
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

func (h *Session) onContinueRequest(request *api.ContinueRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	response := &api.ContinueResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	h.sess.Play()
}

func (h *Session) onNextRequest(request *api.NextRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	response := &api.NextResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	h.sess.Step(debug.STEP_NEXT)
	h.sess.Play()
}

func (h *Session) onStepInRequest(request *api.StepInRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	response := &api.StepBackResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	h.sess.Step(debug.STEP_IN)
	h.sess.Play()
}

func (h *Session) onStepOutRequest(request *api.StepOutRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	response := &api.StepOutResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	h.sess.Step(debug.STEP_OUT)
	h.sess.Play()
}

func (h *Session) onStepBackRequest(request *api.StepBackRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "StepBackRequest is not yet supported"))
}

func (h *Session) onReverseContinueRequest(request *api.ReverseContinueRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "ReverseContinueRequest is not yet supported"))
}

func (h *Session) onRestartFrameRequest(request *api.RestartFrameRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}

	response := &api.RestartFrameResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	h.sess.RestartFrame(request.Arguments.FrameId)

	h.send(&api.StoppedEvent{
		Event: *newEvent("stopped"),
		Body:  api.StoppedEventBody{Reason: "restart", ThreadId: 1, AllThreadsStopped: true},
	})
}

func (h *Session) onGotoRequest(request *api.GotoRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "GotoRequest is not yet supported"))
}

func (h *Session) onPauseRequest(request *api.PauseRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	h.sess.Pause()
	response := &api.PauseResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)
}

func (h *Session) onStackTraceRequest(request *api.StackTraceRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	response := &api.StackTraceResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.sess.GetStackFrames(func(fr *asm.Frame, id int, inst asm.Inst) {
		pos := inst.GetSourceInfo().Start
		response.Body.StackFrames = append(response.Body.StackFrames, api.StackFrame{
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

func (h *Session) onScopesRequest(request *api.ScopesRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	response := &api.ScopesResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	frameId := request.Arguments.FrameId
	h.sess.GetStackFrames(func(fr *asm.Frame, id int, inst asm.Inst) {
		si := fr.Scope.SourceInfo
		scopeId := id * 1024
		if fr.Scope.IsGlobal {
			response.Body.Scopes = append(response.Body.Scopes, api.Scope{
				Name:               "Globals",
				PresentationHint:   "globals",
				VariablesReference: scopeId,
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
			response.Body.Scopes = append(response.Body.Scopes, api.Scope{
				Name:               "Locals",
				PresentationHint:   "locals",
				VariablesReference: scopeId,
				NamedVariables:     len(fr.Locals),
				IndexedVariables:   0, // should also be len?
				Expensive:          false,
				Source:             &h.source,
				Line:               si.Start.Line + h.lineOff,
				Column:             si.Start.Column + h.colOff,
				EndLine:            si.End.Line + h.lineOff,
				EndColumn:          si.End.Column + h.colOff,
			})
			response.Body.Scopes = append(response.Body.Scopes, api.Scope{
				Name:               "Arguments",
				PresentationHint:   "arguments",
				VariablesReference: scopeId + 512,
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

func (h *Session) onVariablesRequest(request *api.VariablesRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	varId := request.Arguments.VariablesReference
	scopeId := varId / 1024
	isParamScope := varId&512 != 0
	_ = varId % 512 // TODO
	response := &api.VariablesResponse{}
	response.Response = *newResponse(request.Seq, request.Command)

	addVar := func(val any, vd *ast.VarDecl, id int) {
		response.Body.Variables = append(response.Body.Variables, api.Variable{
			Name:               vd.Name,
			Value:              asm.DebugStringVal(val),
			Type:               vd.Type.String(),
			PresentationHint:   nil,
			EvaluateName:       vd.Name,
			VariablesReference: 0, // TODO: map/list
			NamedVariables:     0, // map
			IndexedVariables:   0, // list
			MemoryReference:    "",
		})
	}

	h.sess.GetStackFrames(func(fr *asm.Frame, id int, inst asm.Inst) {
		if id != scopeId {
			return
		}
		if isParamScope {
			for i, vd := range fr.Scope.Params {
				addVar(fr.Args[i], vd, i)
			}
		} else {
			nArgs := len(fr.Args)
			for i, vd := range fr.Scope.Params {
				addVar(fr.Locals[i], vd, i)
			}
			for i, vd := range fr.Scope.Locals {
				addVar(fr.Locals[i+nArgs], vd, i)
			}
		}
	})
	h.send(response)
}

func (h *Session) onSetVariableRequest(request *api.SetVariableRequest) {
	name := request.Arguments.Name
	value := request.Arguments.Value
	varId := request.Arguments.VariablesReference
	scopeId := varId / 1024
	isParamScope := varId&512 != 0
	_ = varId % 512 // TODO

	var err error
	var typStr string
	var valStr string
	h.sess.GetStackFrames(func(fr *asm.Frame, id int, inst asm.Inst) {
		if id != scopeId {
			return
		}

		decl, ref := func() (*ast.VarDecl, *any) {
			if isParamScope {
				for i, vd := range fr.Scope.Params {
					if vd.Name == name {
						return vd, &fr.Args[i]
					}
				}
			} else {
				nArgs := len(fr.Args)
				for i, vd := range fr.Scope.Params {
					if vd.Name == name {
						return vd, &fr.Locals[i]
					}
				}
				for i, vd := range fr.Scope.Locals {
					if vd.Name == name {
						return vd, &fr.Locals[i+nArgs]
					}
				}
			}
			return nil, nil
		}()

		if decl == nil {
			err = fmt.Errorf("unknown variable %s", name)
			return
		}
		if !decl.Type.IsPrimitive() {
			err = fmt.Errorf("variable %s of type %s is not primitive", name, decl.Type)
			return
		}

		var val any
		if value == "<nil>" {
			val = nil
		} else if lit := parse.ParseLiteral(value, ast.SourceInfo{}, decl.Type.AsPrimitive()); lit != nil {
			val = lit.Val
			if str, ok := val.(string); ok {
				val = []byte(str)
			}
		} else {
			err = fmt.Errorf("failed to parse value %q of type %s", value, decl.Type)
			return
		}

		*ref = val
		typStr = decl.Type.String()
		valStr = asm.DebugStringVal(val)
	})
	if err != nil {
		h.send(newErrorResponse(request.Seq, request.Command, err.Error()))
		return
	}

	response := &api.SetVariableResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	response.Body = api.SetVariableResponseBody{
		Value:              valStr,
		Type:               typStr,
		VariablesReference: request.Arguments.VariablesReference,
		NamedVariables:     0,
		IndexedVariables:   0,
	}
	h.send(response)
}

func (h *Session) onSetExpressionRequest(request *api.SetExpressionRequest) {
	// TODO: support
	h.send(newErrorResponse(request.Seq, request.Command, "SetExpressionRequest is not yet supported"))
}

func (h *Session) onSourceRequest(request *api.SourceRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "SourceRequest is not yet supported"))
}

func (h *Session) onThreadsRequest(request *api.ThreadsRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	response := &api.ThreadsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	response.Body = api.ThreadsResponseBody{Threads: []api.Thread{{Id: 1, Name: "main"}}}
	h.send(response)

}

func (h *Session) onTerminateThreadsRequest(request *api.TerminateThreadsRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "TerminateRequest is not yet supported"))
}

func (h *Session) onEvaluateRequest(request *api.EvaluateRequest) {
	// TODO: support
	h.send(newErrorResponse(request.Seq, request.Command, "EvaluateRequest is not yet supported"))
}

func (h *Session) onStepInTargetsRequest(request *api.StepInTargetsRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "StepInTargetRequest is not yet supported"))
}

func (h *Session) onGotoTargetsRequest(request *api.GotoTargetsRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "GotoTargetRequest is not yet supported"))
}

func (h *Session) onCompletionsRequest(request *api.CompletionsRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "CompletionRequest is not yet supported"))
}

func (h *Session) onExceptionInfoRequest(request *api.ExceptionInfoRequest) {
	// TODO: support
	h.send(newErrorResponse(request.Seq, request.Command, "ExceptionRequest is not yet supported"))
}

func (h *Session) onLoadedSourcesRequest(request *api.LoadedSourcesRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "LoadedRequest is not yet supported"))
}

func (h *Session) onDataBreakpointInfoRequest(request *api.DataBreakpointInfoRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "DataBreakpointInfoRequest is not yet supported"))
}

func (h *Session) onSetDataBreakpointsRequest(request *api.SetDataBreakpointsRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "SetDataBreakpointsRequest is not yet supported"))
}

func (h *Session) onReadMemoryRequest(request *api.ReadMemoryRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "ReadMemoryRequest is not yet supported"))
}

func (h *Session) onDisassembleRequest(request *api.DisassembleRequest) {
	// TODO: support
	h.send(newErrorResponse(request.Seq, request.Command, "DisassembleRequest is not yet supported"))
}

func (h *Session) onCancelRequest(request *api.CancelRequest) {
	h.send(newErrorResponse(request.Seq, request.Command, "CancelRequest is not yet supported"))
}

func (h *Session) onBreakpointLocationsRequest(request *api.BreakpointLocationsRequest) {
	// TODO(scottb: debug this more
	path := request.Arguments.Source.Path
	response := &api.BreakpointLocationsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	startLine := request.Arguments.Line
	endLine := max(startLine, request.Arguments.EndLine)
	if h.sess != nil && h.sess.File == path {
		// filter down only to lines that are actually executable
		for line := startLine; line <= endLine; line++ {
			if h.sess.MapSourceLine(line-h.lineOff) >= 0 {
				response.Body.Breakpoints = append(response.Body.Breakpoints, api.BreakpointLocation{
					Line:   line,
					Column: h.colOff,
				})
			}
		}
	} else {
		found := debug.FindBreakpoints(request.Arguments.Source.Path)
		for line := startLine; line <= endLine; line++ {
			if found[line-h.lineOff] {
				response.Body.Breakpoints = append(response.Body.Breakpoints, api.BreakpointLocation{
					Line:   line,
					Column: h.colOff,
				})
			}
		}
	}
	h.send(response)
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
