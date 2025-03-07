package main

import (
	"encoding/json"
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"io"
	"os"
)

type Diagnostic struct {
	Range    Range  `json:"range"`
	Message  string `json:"message"`
	Severity int    `json:"severity"`
	Source   string `json:"source,omitempty"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

func reportErrors(errs []ast.Error, desc string, asJson bool, dst io.Writer) {
	if asJson {
		ret := make([]Diagnostic, 0, len(errs))
		for _, e := range errs {
			ret = append(ret, Diagnostic{
				Range: Range{
					Start: Position{Line: e.Start.Line, Character: e.Start.Column},
					End:   Position{Line: e.End.Line, Character: e.End.Column},
				},
				Message:  e.Desc,
				Severity: 0, // TODO: severities?
				Source:   "gaddis",
			})
		}
		buf, err := json.Marshal(ret)
		if err != nil {
			panic(err)
		}
		_, _ = os.Stdout.Write(buf)
	} else {
		for _, err := range ast.ErrorSort(errs) {
			_, _ = fmt.Fprintf(dst, "%s:%v\n", desc, err)
		}
	}
}
