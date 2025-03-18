package debug

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/dragonsinth/gaddis"
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/ast"
	"os"
)

type Source struct {
	Name string
	Path string
	Src  string
	Sum  string

	Errors  []ast.Error
	Program *ast.Program

	// will be nil if the source is invalid
	Assembled   *asm.Assembly
	Breakpoints *Breakpoints
}

func LoadSource(filename string) (*Source, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading: %s", filename)
	}
	src := string(buf)

	sh := sha256.New()
	sh.Write(buf)
	sum := hex.EncodeToString(sh.Sum(nil))

	prog, _, errs := gaddis.Compile(src)
	ret := &Source{
		Path: filename,
		Src:  src,
		Sum:  sum,

		Errors:  errs,
		Program: prog,
	}
	if len(errs) == 0 {
		ret.Assembled = asm.Assemble(prog)
		ret.Breakpoints = NewBreakpoints(ret.Assembled.Code)
	}

	return ret, nil
}

func (ds *Session) runInVm(f func(fromIoWait bool)) {
	if ds.commandsClosed.Load() {
		return
	}

	// force a yield, run the given command on the internal thread
	ds.yield.Store(true)

	done := make(chan struct{})
	cmd := func(fromIoWait bool) {
		defer close(done)
		ds.yield.Store(false)
		f(fromIoWait)
	}

	select {
	case <-ds.done:
	case ds.commands <- cmd:
	}

	select {
	case <-ds.done:
	case <-done:
	}
}

type sentinelIoExit struct{}

type sentinelIoInterrupt struct{}
