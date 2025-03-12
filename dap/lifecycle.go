package dap

import (
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
	h.canTerminal = request.Arguments.SupportsRunInTerminalRequest

	response := &api.InitializeResponse{
		Response: *newResponse(request.Seq, request.Command),
		Body: api.Capabilities{
			SupportsConfigurationDoneRequest:   true,
			SupportsSetVariable:                true,
			SupportsRestartFrame:               true,
			SupportsRestartRequest:             true,
			SupportsExceptionInfoRequest:       true,
			SupportTerminateDebuggee:           true,
			SupportSuspendDebuggee:             true,
			SupportsLoadedSourcesRequest:       true,
			SupportsTerminateRequest:           true,
			SupportsDisassembleRequest:         true,
			SupportsBreakpointLocationsRequest: true,
			SupportsInstructionBreakpoints:     true,
			SupportsSteppingGranularity:        true,
			SupportedChecksumAlgorithms:        []api.ChecksumAlgorithm{"SHA256"},

			SupportsStepInTargetsRequest: false, // what is this
			SupportsSetExpression:        false, // support later

			SupportsFunctionBreakpoints:           false,
			SupportsConditionalBreakpoints:        false,
			SupportsHitConditionalBreakpoints:     false,
			SupportsEvaluateForHovers:             false,
			ExceptionBreakpointFilters:            nil,
			SupportsStepBack:                      false,
			SupportsGotoTargetsRequest:            false,
			SupportsCompletionsRequest:            false,
			CompletionTriggerCharacters:           nil,
			SupportsModulesRequest:                false,
			AdditionalModuleColumns:               nil,
			SupportsExceptionOptions:              false,
			SupportsValueFormattingOptions:        false,
			SupportsDelayedStackTraceLoading:      false,
			SupportsLogPoints:                     false,
			SupportsTerminateThreadsRequest:       false,
			SupportsDataBreakpoints:               false,
			SupportsReadMemoryRequest:             false,
			SupportsWriteMemoryRequest:            false,
			SupportsCancelRequest:                 false,
			SupportsClipboardContext:              false,
			SupportsExceptionFilterOptions:        false,
			SupportsSingleThreadExecutionRequests: false,
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

	if !h.tryStartSession(args, request.GetRequest()) {
		return // already notified
	}

	response := &api.LaunchResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	if h.noDebug {
		// launch immediately, otherwise wait for configuration done.
		h.sess.SetNoDebug()
		h.sess.SetLineBreakpoints(nil)
		h.sess.SetInstBreakpoints(nil)
		h.sess.Play()
	}
}

func (h *Session) onRestartRequest(request *api.RestartRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	h.sess.Host.SuppressAllEvents()
	if h.terminal != nil {
		h.terminal.Interrupt()
	}
	h.sess.Halt()
	h.sess.Wait()
	if h.terminal != nil {
		h.terminal.Continue()
	}
	h.sess = nil

	// clear any lingering client state for the old session
	h.send(&api.ExitedEvent{
		Event: *newEvent("exited"),
		Body:  api.ExitedEventBody{ExitCode: 1},
	})

	var wrap struct {
		Arguments launchArgs `json:"arguments"`
	}
	if err := fromJson(request.Arguments, &wrap); err != nil {
		h.send(newErrorResponse(request.Seq, request.Command, "could not parse restart request"))
		return
	}

	if !h.tryStartSession(wrap.Arguments, request.GetRequest()) {
		return // already notified
	}

	response := &api.RestartResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	if h.noDebug {
		h.sess.SetNoDebug()
		h.sess.SetLineBreakpoints(nil)
		h.sess.SetInstBreakpoints(nil)
	} else {
		if h.stopOnEntry {
			h.sess.StopOnEntry()
		}
		h.sess.SetLineBreakpoints(h.bpsBySum[h.sess.Source.Sum])
		h.sess.SetInstBreakpoints(h.instBps)
	}
	h.sess.Play()
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
	if h.stopOnEntry {
		h.sess.StopOnEntry()
	}
	h.sess.SetLineBreakpoints(h.bpsBySum[h.sess.Source.Sum])
	h.sess.SetInstBreakpoints(h.instBps)
	h.sess.Play()
}

func (h *Session) onDisconnectRequest(request *api.DisconnectRequest) {
	if h.terminal != nil {
		h.terminal.Interrupt()
	}
	if h.sess != nil {
		h.sess.Terminate()
	}
	response := &api.DisconnectResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)
}

func (h *Session) onTerminateRequest(request *api.TerminateRequest) {
	if h.terminal != nil {
		h.terminal.Interrupt()
	}
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

func (h *Session) onRunInTerminalResponse(response *api.RunInTerminalResponse) {
	h.terminalPid = response.Body.ProcessId
}
