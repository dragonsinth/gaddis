package dap

import (
	"github.com/dragonsinth/gaddis/debug"
	api "github.com/google/go-dap"
)

func (h *Session) tryStartSession(args launchArgs, request *api.Request) bool {
	source, err := debug.LoadSource(args.Program)
	if err != nil {
		h.send(newErrorResponse(request.Seq, request.Command, err.Error()))
		return false
	}
	if len(source.Errors) > 0 {
		srcPtr := dapSource(*source)
		for _, err := range source.Errors {
			h.send(&api.OutputEvent{
				Event: *newEvent("output"),
				Body: api.OutputEventBody{
					Category: "stderr",
					Output:   err.Desc + "\n",
					Source:   srcPtr,
					Line:     err.Start.Line + h.lineOff,
					Column:   err.Start.Column + h.colOff,
				},
			})
		}
		h.send(newErrorResponse(request.Seq, request.Command, "compile errors"))
		return false
	}

	source.Name = args.Name
	h.sourceByPath[source.Path] = source
	h.sourceBySum[source.Sum] = source
	h.source = dapSource(*source)
	h.stopOnEntry = args.StopOnEntry
	h.noDebug = args.NoDebug

	opts := debug.Opts{
		Stdin:   tryReadInput(args.Program),
		Stdout:  h.stdout,
		WorkDir: args.WorkDir,
	}
	host := eventHost{
		send:    h.send,
		source:  h.source,
		lineOff: h.lineOff,
		colOff:  h.colOff,
	}

	h.sess = debug.New(*source, host, opts)
	return true
}
