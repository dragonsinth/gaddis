package dap

import (
	"fmt"
	"github.com/dragonsinth/gaddis"
	"github.com/dragonsinth/gaddis/debug"
	api "github.com/google/go-dap"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

	host := eventHost{
		sendFunc: h.send,
		source:   h.source,
		lineOff:  h.lineOff,
		colOff:   h.colOff,
	}

	ioProvider := gaddis.IoAdapter{
		In: func() (string, error) {
			return "", io.EOF
		},
		Out: h.stdout,
	}
	if args.TestMode {
		outfile := args.Program + ".out"
		expectOutput, err := os.ReadFile(outfile)
		if err != nil {
			h.send(newErrorResponse(request.Seq, request.Command, "testing requires and output file; please create "+outfile))
		}

		host.isTest = true
		host.wantOutput = string(expectOutput)
		testInput := tryReadInput(args.Program)
		host.remainingInput = gaddis.SplitInput(string(testInput))
		ioProvider.In = host.remainingInput
		ioProvider.Out = func(line string) {
			host.capturedOutput.WriteString(line)
			h.stdout(line)
		}
	} else {
		if h.terminal == nil && h.canTerminal {
			title := "Gaddis Debug " + args.Name
			if args.NoDebug {
				title = "Gaddis Run " + title
			}

			if term, err := StartTerminal(); err != nil {
				log.Println("error: failed to start terminal conn:", err)
			} else {
				h.terminal = term
				terminalArgs := []string{os.Args[0], "-port", strconv.Itoa(term.Port), "terminal"}
				h.send(&api.RunInTerminalRequest{
					Request: *newRequest("runInTerminal"),
					Arguments: api.RunInTerminalRequestArguments{
						Kind:  "integrated",
						Title: title,
						Cwd:   "",
						Args:  terminalArgs,
						Env:   nil,
					},
				})
			}
		}
		if h.terminal != nil {
			// TODO: actually implement io/wait signaling.
			ioProvider.In = gaddis.StreamInput(h.terminal)
			ioProvider.Out = func(line string) {
				h.stdout(line)
				h.terminal.Output(line)
			}
			banner := strings.Repeat("-", len(args.Name))
			h.terminal.Output(fmt.Sprintf("\x1b[H\x1b[J/%s\\\n|%s|\n\\%s/\n", banner, args.Name, banner))
		}
	}

	source.Name = filepath.Base(source.Path)
	h.sourceByPath[source.Path] = source
	h.sourceBySum[source.Sum] = source
	h.source = dapSource(*source)
	h.launchArgs = args

	opts := debug.Opts{
		IoProvider:  ioProvider,
		NoDebug:     args.NoDebug,
		StopOnEntry: args.StopOnEntry,
		LineBreaks:  h.bpsBySum[source.Sum],
		InstBreaks:  h.instBps,
	}
	h.sess = debug.New(*source, &host, opts)
	h.runId++
	return true
}
