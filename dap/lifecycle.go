package dap

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/debug"
	api "github.com/google/go-dap"
)

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
	host := eventHost{
		send:    h.send,
		source:  h.source,
		lineOff: h.lineOff,
		colOff:  h.colOff,
	}

	sess, err := debug.New(h.source.Path, compileErr, host, opts)
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
		h.sess.SetNoDebug()
		h.sess.SetBreakpoints(nil)
		h.sess.Play()
	}
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
		h.sess.SetNoDebug()
		h.sess.SetBreakpoints(nil)
	} else if h.stopOnEntry {
		h.sess.StopOnEntry()
	}
	h.sess.Play()
}
