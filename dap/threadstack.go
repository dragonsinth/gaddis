package dap

import (
	"github.com/dragonsinth/gaddis/asm"
	api "github.com/google/go-dap"
)

func (h *Session) onThreadsRequest(request *api.ThreadsRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	response := &api.ThreadsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	response.Body = api.ThreadsResponseBody{Threads: []api.Thread{{Id: h.runId, Name: "main"}}}
	h.send(response)
}

func (h *Session) onStackTraceRequest(request *api.StackTraceRequest) {
	if h.pausedSessionRequiredError(request) {
		return
	}
	if request.Arguments.ThreadId != h.runId {
		h.send(newErrorResponse(request.Seq, request.Command, "unknown threadId"))
		return
	}
	response := &api.StackTraceResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.sess.GetStackFrames(func(fr *asm.Frame, id int, inst asm.Inst, pc int) {
		if fr.Native != nil {
			n := fr.Native
			response.Body.StackFrames = append(response.Body.StackFrames, api.StackFrame{
				Id:         id,
				Name:       n.Func + "()",
				Source:     libSource(n.File),
				Line:       n.Line + h.lineOff,
				CanRestart: false,
			})
		} else {
			pos := inst.GetSourceInfo().Start
			response.Body.StackFrames = append(response.Body.StackFrames, api.StackFrame{
				Id:         id,
				Name:       fr.Scope.Desc(),
				Source:     h.source,
				Line:       pos.Line + h.lineOff,
				Column:     h.colOff, // don't do columns yet... it's too weird
				CanRestart: true,

				InstructionPointerReference: asm.PcRef(pc),
			})
		}
	})
	response.Body.TotalFrames = len(response.Body.StackFrames)
	h.send(response)
}

func (h *Session) onScopesRequest(request *api.ScopesRequest) {
	if h.pausedSessionRequiredError(request) {
		return
	}
	response := &api.ScopesResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	targetFrameId := request.Arguments.FrameId
	h.sess.GetStackFrames(func(fr *asm.Frame, frameId int, inst asm.Inst, _ int) {
		if fr.Native != nil {
			return
		}
		si := fr.Scope.SourceInfo
		ids := getScopeIds(frameId)
		if fr.Scope.IsGlobal {
			response.Body.Scopes = append(response.Body.Scopes, api.Scope{
				Name:               "Globals",
				PresentationHint:   "globals",
				VariablesReference: ids.localId,
				NamedVariables:     len(fr.Locals),
				IndexedVariables:   0, // should also be len?
				Expensive:          false,
				Source:             h.source,
				Line:               si.Start.Line + h.lineOff,
				Column:             si.Start.Column + h.colOff,
				EndLine:            si.End.Line + h.lineOff,
				EndColumn:          si.End.Column + h.colOff,
			})
		} else if frameId == targetFrameId {
			if len(fr.Locals) > 0 {
				response.Body.Scopes = append(response.Body.Scopes, api.Scope{
					Name:               "Locals",
					PresentationHint:   "locals",
					VariablesReference: ids.localId,
					NamedVariables:     len(fr.Locals),
					IndexedVariables:   0, // should also be len?
					Expensive:          false,
					Source:             h.source,
					Line:               si.Start.Line + h.lineOff,
					Column:             si.Start.Column + h.colOff,
					EndLine:            si.End.Line + h.lineOff,
					EndColumn:          si.End.Column + h.colOff,
				})
			}
			if len(fr.Params) > 0 {
				response.Body.Scopes = append(response.Body.Scopes, api.Scope{
					Name:               "Params",
					PresentationHint:   "locals",
					VariablesReference: ids.paramId,
					NamedVariables:     len(fr.Params),
					IndexedVariables:   0, // should also be len?
					Expensive:          false,
					Source:             h.source,
					Line:               si.Start.Line + h.lineOff,
					Column:             si.Start.Column + h.colOff,
					EndLine:            si.End.Line + h.lineOff,
					EndColumn:          si.End.Column + h.colOff,
				})
			}
			if len(fr.Args) > 0 {
				response.Body.Scopes = append(response.Body.Scopes, api.Scope{
					Name:               "Arguments",
					PresentationHint:   "arguments",
					VariablesReference: ids.argsId,
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
		}
	})
	h.send(response)
}

type scopeIds struct {
	localId int
	paramId int
	argsId  int
}

func getScopeIds(frameId int) scopeIds {
	scopeId := frameId << 16
	paramId := scopeId + (1 << 14)
	argsId := scopeId + (2 << 14)
	return scopeIds{
		localId: scopeId,
		paramId: paramId,
		argsId:  argsId,
	}
}

func getFrameId(scopeId int) int {
	return scopeId >> 16
}

func getScopeId(varId int) int {
	return (varId >> 14) << 14
}
