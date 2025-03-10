package debug

import (
	"bytes"
	"github.com/dragonsinth/gaddis"
	"github.com/dragonsinth/gaddis/asm"
	"os"
	"slices"
)

func ValidateBreakpoints(filename string, bps []int) ([]int, bool) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, false
	}
	src := string(buf)

	prog, _, errs := gaddis.Compile(src)
	if len(errs) > 0 {
		return nil, false
	}

	assembled := asm.Assemble(prog)

	found := map[int]bool{}
	for _, inst := range assembled.Code {
		line := inst.GetSourceInfo().Start.Line
		found[line] = true
	}
	return slices.DeleteFunc(bps, func(line int) bool {
		return !found[line]
	}), true
}

type bufferedSyncWriter struct {
	out    func(string)
	buffer bytes.Buffer
}

func (b *bufferedSyncWriter) Write(p []byte) (int, error) {
	n := len(p)
	for pos := bytes.Index(p, []byte{'\n'}); pos >= 0; pos = bytes.Index(p, []byte{'\n'}) {
		first, rest := p[:pos+1], p[pos+1:]
		b.buffer.Write(first)
		_ = b.Sync()
		p = rest
	}
	b.buffer.Write(p)
	return n, nil
}

func (b *bufferedSyncWriter) Sync() error {
	if b.buffer.Len() > 0 {
		output := b.buffer.String()
		b.buffer.Reset()
		b.out(output)
	}
	return nil
}
