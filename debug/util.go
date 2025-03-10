package debug

import (
	"bytes"
	"github.com/dragonsinth/gaddis"
	"github.com/dragonsinth/gaddis/asm"
	"os"
)

func FindBreakpoints(filename string) map[int]bool {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil
	}
	src := string(buf)

	prog, _, errs := gaddis.Compile(src)
	if len(errs) > 0 {
		return nil
	}

	assembled := asm.Assemble(prog)

	found := map[int]bool{}
	for _, inst := range assembled.Code {
		line := inst.GetSourceInfo().Start.Line
		found[line] = true
	}
	return found
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

func (ds *Session) withOuterLock(f func()) {
	// force a yield, acquire the lock
	ds.yield.Store(true)
	ds.runMu.Lock()
	defer ds.runMu.Unlock()
	ds.yield.Store(false)
	f()
}
