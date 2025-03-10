package dap

import (
	"github.com/dragonsinth/gaddis/debug"
	api "github.com/google/go-dap"
)

func (h *Session) onSetBreakpointsRequest(request *api.SetBreakpointsRequest) {
	source := request.Arguments.Source

	response := &api.SetBreakpointsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	if h.noDebug {
		h.send(response)
		return
	}

	var found map[int]bool
	if h.sess != nil && h.sess.File == source.Path {
		found = h.sess.ValidLineBreaks
	} else if h.sess == nil {
		found = debug.FindBreakpoints(source.Path)
	} else {
		// cannot accept the breakpoints
		found = nil
	}

	var bps []int
	for _, bp := range request.Arguments.Breakpoints {
		var msg string
		srcLine := bp.Line - h.lineOff
		verified := found[srcLine]
		if verified {
			bps = append(bps, srcLine)
		} else {
			msg = "failed"
		}
		response.Body.Breakpoints = append(response.Body.Breakpoints, api.Breakpoint{
			Verified: verified,
			Message:  msg,
			Source:   &source,
			Line:     bp.Line,
			Column:   h.colOff,
		})
	}

	if h.sess != nil && h.sess.File == source.Path {
		h.sess.SetBreakpoints(bps)
	} else if h.sess == nil {
		h.bps[source.Path] = bps
	}
	h.send(response)
}

func (h *Session) onSetExceptionBreakpointsRequest(request *api.SetExceptionBreakpointsRequest) {
	response := &api.SetExceptionBreakpointsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)
}

func (h *Session) onBreakpointLocationsRequest(request *api.BreakpointLocationsRequest) {
	// TODO(scottb: debug this more
	path := request.Arguments.Source.Path
	response := &api.BreakpointLocationsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	startLine := request.Arguments.Line
	endLine := max(startLine, request.Arguments.EndLine)
	if h.sess != nil && h.sess.File == path {
		// filter down only to lines that are actually executable
		for line := startLine; line <= endLine; line++ {
			if h.sess.MapSourceLine(line-h.lineOff) >= 0 {
				response.Body.Breakpoints = append(response.Body.Breakpoints, api.BreakpointLocation{
					Line:   line,
					Column: h.colOff,
				})
			}
		}
	} else {
		found := debug.FindBreakpoints(request.Arguments.Source.Path)
		for line := startLine; line <= endLine; line++ {
			if found[line-h.lineOff] {
				response.Body.Breakpoints = append(response.Body.Breakpoints, api.BreakpointLocation{
					Line:   line,
					Column: h.colOff,
				})
			}
		}
	}
	h.send(response)
}
