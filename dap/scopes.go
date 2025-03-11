package dap

import (
	"github.com/dragonsinth/gaddis/asm"
	api "github.com/google/go-dap"
)

func (h *Session) onStackTraceRequest(request *api.StackTraceRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	response := &api.StackTraceResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.sess.GetStackFrames(func(fr *asm.Frame, id int, inst asm.Inst, pc int) {
		pos := inst.GetSourceInfo().Start
		response.Body.StackFrames = append(response.Body.StackFrames, api.StackFrame{
			Id:         id,
			Name:       fr.Scope.Desc(),
			Source:     h.source,
			Line:       pos.Line + h.lineOff,
			Column:     h.colOff, // don't do columns yet... it's too weird
			CanRestart: true,

			InstructionPointerReference: pcRef(pc),
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
	h.sess.GetStackFrames(func(fr *asm.Frame, id int, inst asm.Inst, _ int) {
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
				Source:             h.source,
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
				Source:             h.source,
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
				Source:             h.source,
				Line:               si.Start.Line + h.lineOff,
				Column:             si.Start.Column + h.colOff,
				EndLine:            si.End.Line + h.lineOff,
				EndColumn:          si.End.Column + h.colOff,
			})
		}
	})
	h.send(response)
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
