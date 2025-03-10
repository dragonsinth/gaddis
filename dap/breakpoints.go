package dap

import (
	"github.com/dragonsinth/gaddis/debug"
	api "github.com/google/go-dap"
)

func (h *Session) onSetBreakpointsRequest(request *api.SetBreakpointsRequest) {
	source := &request.Arguments.Source
	path := source.Path

	response := &api.SetBreakpointsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	if h.noDebug {
		h.send(response)
		return
	}

	var found map[int]bool
	var isSourceSession bool
	if h.sess != nil && h.sess.File == path {
		isSourceSession = true
		found = h.sess.ValidLineBreaks
	} else if h.sess == nil {
		found = debug.FindBreakpoints(path)
	} else {
		// cannot accept the breakpoints
		found = nil
	}

	var bps []int
	for _, bp := range request.Arguments.Breakpoints {
		var msg string
		srcLine := bp.Line - h.lineOff
		if found[srcLine] {
			bps = append(bps, srcLine)
			var ref string
			if isSourceSession {
				ref = pcRef(h.sess.SourceToInst[srcLine])
			}
			response.Body.Breakpoints = append(response.Body.Breakpoints, api.Breakpoint{
				Verified:             true,
				Message:              msg,
				Source:               source,
				Line:                 bp.Line,
				Column:               h.colOff,
				InstructionReference: ref,
			})
			source = nil
		} else {
			response.Body.Breakpoints = append(response.Body.Breakpoints, api.Breakpoint{
				Verified: false,
				Message:  "failed",
			})
		}

	}

	if isSourceSession {
		h.sess.SetLineBreakpoints(bps)
	} else if h.sess == nil {
		h.bps[path] = bps
	}
	h.send(response)
}

func (h *Session) onSetExceptionBreakpointsRequest(request *api.SetExceptionBreakpointsRequest) {
	response := &api.SetExceptionBreakpointsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)
}

func (h *Session) onSetInstructionBreakpointsRequest(request *api.SetInstructionBreakpointsRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}

	response := &api.SetInstructionBreakpointsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)

	var pcs []int
	for _, bp := range request.Arguments.Breakpoints {
		origPc := refPc(bp.InstructionReference)
		pc := origPc + (bp.Offset / 4)
		if origPc < 0 || pc < 0 || pc >= h.sess.NInst {
			response.Body.Breakpoints = append(response.Body.Breakpoints, api.Breakpoint{
				Verified: false,
				Message:  "failed",
			})
		} else {
			pcs = append(pcs, pc)
			response.Body.Breakpoints = append(response.Body.Breakpoints, api.Breakpoint{
				Verified:             true,
				Line:                 h.sess.InstToSource[pc] + h.lineOff,
				Column:               h.colOff,
				InstructionReference: pcRef(pc),
			})
		}

	}
	h.sess.SetInstBreakpoints(pcs)
	h.send(response)
}

func (h *Session) onBreakpointLocationsRequest(request *api.BreakpointLocationsRequest) {
	path := request.Arguments.Source.Path
	response := &api.BreakpointLocationsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	startLine := request.Arguments.Line
	endLine := max(startLine, request.Arguments.EndLine)

	var found map[int]bool
	if h.sess != nil && h.sess.File == path {
		found = h.sess.ValidLineBreaks
	} else if h.sess == nil {
		found = debug.FindBreakpoints(path)
	} else {
		found = nil
	}

	for line := startLine; line <= endLine; line++ {
		if found[line-h.lineOff] {
			response.Body.Breakpoints = append(response.Body.Breakpoints, api.BreakpointLocation{
				Line:   line,
				Column: h.colOff,
			})
		}
	}

	h.send(response)
}

func (h *Session) onExceptionInfoRequest(request *api.ExceptionInfoRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	if request.Arguments.ThreadId != 1 {
		h.send(newErrorResponse(request.Seq, request.Command, "unknown threadId"))
		return
	}

	response := &api.ExceptionInfoResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	response.Body = api.ExceptionInfoResponseBody{
		ExceptionId: "",
		Description: "",
		BreakMode:   "always",
	}

	trace, exception := h.sess.GetCurrentException()
	if exception != nil {
		msg := exception.Error()
		response.Body.Description = msg
		response.Body.Details = &api.ExceptionDetails{
			Message:    msg,
			StackTrace: trace,
		}
	}
	h.send(response)
}
