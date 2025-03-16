package dap

import (
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/debug"
	api "github.com/google/go-dap"
)

func (h *Session) onContinueRequest(request *api.ContinueRequest) {
	if h.pausedSessionRequiredError(request) {
		return
	}
	if request.Arguments.ThreadId != h.runId {
		h.send(newErrorResponse(request.Seq, request.Command, "unknown threadId"))
		return
	}
	response := &api.ContinueResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	h.Resume()
}

func (h *Session) onPauseRequest(request *api.PauseRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	if request.Arguments.ThreadId != h.runId {
		h.send(newErrorResponse(request.Seq, request.Command, "unknown threadId"))
		return
	}
	h.sess.Pause()
	response := &api.PauseResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)
}

func (h *Session) onNextRequest(request *api.NextRequest) {
	if h.pausedSessionRequiredError(request) {
		return
	}
	if request.Arguments.ThreadId != h.runId {
		h.send(newErrorResponse(request.Seq, request.Command, "unknown threadId"))
		return
	}
	response := &api.NextResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	gran := request.Arguments.Granularity == "instruction"
	h.sess.Step(debug.STEP_NEXT, debug.StepGran(gran))
	h.Resume()
}

func (h *Session) onStepInRequest(request *api.StepInRequest) {
	if h.pausedSessionRequiredError(request) {
		return
	}
	if request.Arguments.ThreadId != h.runId {
		h.send(newErrorResponse(request.Seq, request.Command, "unknown threadId"))
		return
	}
	response := &api.StepBackResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	// TODO: step granularity!
	gran := request.Arguments.Granularity == "instruction"
	h.sess.Step(debug.STEP_IN, debug.StepGran(gran))
	h.Resume()
}

func (h *Session) onStepOutRequest(request *api.StepOutRequest) {
	if h.pausedSessionRequiredError(request) {
		return
	}
	if request.Arguments.ThreadId != h.runId {
		h.send(newErrorResponse(request.Seq, request.Command, "unknown threadId"))
		return
	}
	response := &api.StepOutResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	gran := request.Arguments.Granularity == "instruction"
	h.sess.Step(debug.STEP_OUT, debug.StepGran(gran))
	h.Resume()
}

func (h *Session) onRestartFrameRequest(request *api.RestartFrameRequest) {
	if h.pausedSessionRequiredError(request) {
		return
	}

	found := false
	native := false
	targetFrameId := request.Arguments.FrameId
	h.sess.GetStackFrames(func(fr *asm.Frame, frameId int, _ asm.Inst, _ int) {
		if frameId != targetFrameId {
			return
		}
		found = true
		native = fr.Native != nil
	})

	if !found {
		h.send(newErrorResponse(request.Seq, request.Command, "no frame found"))
		return
	} else if native {
		h.send(newErrorResponse(request.Seq, request.Command, "cannot restart external frame"))
		return
	}

	response := &api.RestartFrameResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	h.sess.RestartFrame(targetFrameId)

	h.send(&api.StoppedEvent{
		Event: *newEvent("stopped"),
		Body:  api.StoppedEventBody{Reason: "restart", ThreadId: 1, AllThreadsStopped: true},
	})
}
