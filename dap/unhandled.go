package dap

import (
	"fmt"
	api "github.com/google/go-dap"
)

func (h *Session) onAttachRequest(request *api.AttachRequest) {
	h.unhandled(request.Request)
}

func (h *Session) onSetFunctionBreakpointsRequest(request *api.SetFunctionBreakpointsRequest) {
	h.unhandled(request.Request)
}

func (h *Session) onSetExpressionRequest(request *api.SetExpressionRequest) {
	// TODO: support
	h.unhandled(request.Request)
}

func (h *Session) onEvaluateRequest(request *api.EvaluateRequest) {
	// TODO: support
	h.unhandled(request.Request)
}

func (h *Session) onStepBackRequest(request *api.StepBackRequest) {
	h.unhandled(request.Request)
}

func (h *Session) onReverseContinueRequest(request *api.ReverseContinueRequest) {
	h.unhandled(request.Request)
}
func (h *Session) onGotoRequest(request *api.GotoRequest) {
	h.unhandled(request.Request)
}

func (h *Session) onSourceRequest(request *api.SourceRequest) {
	h.unhandled(request.Request)
}

func (h *Session) onTerminateThreadsRequest(request *api.TerminateThreadsRequest) {
	h.unhandled(request.Request)
}

func (h *Session) onStepInTargetsRequest(request *api.StepInTargetsRequest) {
	h.unhandled(request.Request)
}

func (h *Session) onGotoTargetsRequest(request *api.GotoTargetsRequest) {
	h.unhandled(request.Request)
}

func (h *Session) onCompletionsRequest(request *api.CompletionsRequest) {
	h.unhandled(request.Request)
}

func (h *Session) onExceptionInfoRequest(request *api.ExceptionInfoRequest) {
	// TODO: support
	h.unhandled(request.Request)
}

func (h *Session) onLoadedSourcesRequest(request *api.LoadedSourcesRequest) {
	h.unhandled(request.Request)
}

func (h *Session) onDataBreakpointInfoRequest(request *api.DataBreakpointInfoRequest) {
	h.unhandled(request.Request)
}

func (h *Session) onSetDataBreakpointsRequest(request *api.SetDataBreakpointsRequest) {
	h.unhandled(request.Request)
}

func (h *Session) onReadMemoryRequest(request *api.ReadMemoryRequest) {
	h.unhandled(request.Request)
}

func (h *Session) onDisassembleRequest(request *api.DisassembleRequest) {
	// TODO: support
	h.unhandled(request.Request)
}

func (h *Session) onCancelRequest(request *api.CancelRequest) {
	h.send(newErrorResponse(request.GetSeq(), request.Command, ""))
}

func (h *Session) unhandled(request api.Request) {
	h.send(newErrorResponse(request.GetSeq(), request.Command, fmt.Sprintf("%s is not yet supported", request.Command)))
}
