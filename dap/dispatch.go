package dap

import (
	api "github.com/google/go-dap"
	"log"
)

// dispatchRequest processes each response.
func (h *Session) dispatchResponse(rsp api.ResponseMessage) {
	switch response := rsp.(type) {
	case *api.RunInTerminalResponse:
		h.onRunInTerminalResponse(response)
	case *api.StartDebuggingResponse:
		h.onStartDebuggingResponse(response)
	default:
		log.Printf("Unknown response type: %T", response)
	}
}

// dispatchRequest processes each request and sends back events and responses.
func (h *Session) dispatchRequest(req api.RequestMessage) {
	switch request := req.(type) {
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
	case *api.SetInstructionBreakpointsRequest:
		h.onSetInstructionBreakpointsRequest(request)
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
		h.unhandled(req.GetRequest())
	}
}
