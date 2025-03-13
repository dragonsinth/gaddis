package dap

import (
	"fmt"
	api "github.com/google/go-dap"
	"log"
)

func (h *Session) unhandled(request api.RequestMessage) {
	cmd := request.GetRequest().Command
	h.send(newErrorResponse(request.GetSeq(), cmd, fmt.Sprintf("%s is not yet supported", cmd)))
}

func (h *Session) onAttachRequest(request *api.AttachRequest) {
	h.unhandled(request)
}

func (h *Session) onSetFunctionBreakpointsRequest(request *api.SetFunctionBreakpointsRequest) {
	h.unhandled(request)
}

func (h *Session) onSetExpressionRequest(request *api.SetExpressionRequest) {
	// TODO: support
	h.unhandled(request)
}

func (h *Session) onStepBackRequest(request *api.StepBackRequest) {
	h.unhandled(request)
}

func (h *Session) onReverseContinueRequest(request *api.ReverseContinueRequest) {
	h.unhandled(request)
}
func (h *Session) onGotoRequest(request *api.GotoRequest) {
	h.unhandled(request)
}

func (h *Session) onTerminateThreadsRequest(request *api.TerminateThreadsRequest) {
	h.unhandled(request)
}

func (h *Session) onStepInTargetsRequest(request *api.StepInTargetsRequest) {
	h.unhandled(request)
}

func (h *Session) onGotoTargetsRequest(request *api.GotoTargetsRequest) {
	h.unhandled(request)
}

func (h *Session) onCompletionsRequest(request *api.CompletionsRequest) {
	h.unhandled(request)
}

func (h *Session) onLoadedSourcesRequest(request *api.LoadedSourcesRequest) {
	h.unhandled(request)
}

func (h *Session) onDataBreakpointInfoRequest(request *api.DataBreakpointInfoRequest) {
	h.unhandled(request)
}

func (h *Session) onSetDataBreakpointsRequest(request *api.SetDataBreakpointsRequest) {
	h.unhandled(request)
}

func (h *Session) onReadMemoryRequest(request *api.ReadMemoryRequest) {
	h.unhandled(request)
}

func (h *Session) onCancelRequest(request *api.CancelRequest) {
	h.send(newErrorResponse(request.GetSeq(), request.Command, ""))
}

func (h *Session) onStartDebuggingResponse(response *api.StartDebuggingResponse) {
	log.Printf("Unexpected response type: %T", response)
}
