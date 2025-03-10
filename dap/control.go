package dap

import (
	"github.com/dragonsinth/gaddis/debug"
	api "github.com/google/go-dap"
)

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

func (h *Session) onNextRequest(request *api.NextRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	response := &api.NextResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)

	// TODO: step granularity!
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

	// TODO: step granularity!
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

	// TODO: step granularity!
	h.sess.Step(debug.STEP_OUT)
	h.sess.Play()
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
