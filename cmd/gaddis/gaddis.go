package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/dragonsinth/gaddis"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/astprint"
	"github.com/dragonsinth/gaddis/goexec"
	"github.com/dragonsinth/gaddis/gogen"
	"github.com/dragonsinth/gaddis/gogen/builtins"
	"github.com/dragonsinth/gaddis/interp"
	"github.com/dragonsinth/gaddis/parse"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	fDebug = flag.Bool("d", false, "don't delete any generated files, leave for inspection")
	fTest  = flag.Bool("t", false, "legacy: run in test mode; capture")
	fNoRun = flag.Bool("no-run", false, "compile only, do not run")
	fJson  = flag.Bool("json", false, "emit errors as json")
	fGogen = flag.Bool("gogen", false, "run using go compile")
)

const help = `Usage: gaddis <command> [options] [arguments]

Available commands:

format: parse and format the input file
check:  parse and error check the input file
build:  parse, check, and build the input file
run:    everything, including format
test:   run in test mode
help:   print this help message
*.gad:  legacy: run the given file
`

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: gaddis <command> [options] [arguments]")
		os.Exit(1)
	}

	opts := runOpts{
		stopAfterBuild:    *fNoRun,
		isTest:            *fTest,
		leaveBuildOutputs: *fDebug,
		goGen:             *fGogen,
	}

	var err error
	switch args[0] {
	case "help":
		fmt.Print(help)
		// TODO: details subcommand help
		os.Exit(0)
	case "format":
		err = format(args[1:])
	case "check":
		err = check(args[1:])
	case "build":
		// always leave build outputs on build command
		opts.stopAfterBuild = true
		opts.leaveBuildOutputs = true
		err = run(args[1:], opts)
	case "test":
		opts.isTest = true
		err = run(args[1:], opts)
	case "run":
		err = run(args[1:], opts)
	default:
		err = run(args, opts)
	}

	type hasExitCode interface {
		ExitCode() int
	}

	var he hasExitCode
	if errors.As(err, &he) {
		log.Println(err)
		os.Exit(he.ExitCode())
	} else if err != nil {
		log.Fatal(err)
	}
}

type source struct {
	src      string
	filename string
	isStdin  bool
}

func (s *source) desc() string {
	if s.isStdin {
		return "stdin"
	} else {
		return s.filename
	}
}

func readSourceFromArgs(args []string) (*source, error) {
	switch len(args) {
	case 0:
		buf, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("reading stdin: %w", err)
		}
		return &source{
			src:      string(buf),
			filename: "",
			isStdin:  true,
		}, nil
	case 1:
		filename := args[0]
		buf, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", filename, err)
		}
		return &source{
			src:      string(buf),
			filename: filename,
			isStdin:  false,
		}, nil
	default:
		return nil, errors.New("expects 0 or 1 arguments: the source to parse")
	}
}

func format(args []string) error {
	src, err := readSourceFromArgs(args)
	if err != nil {
		return err
	}

	// parse and report lex/parse errors only
	prog, comments, errs := parse.Parse(src.src)
	reportErrors(errs, src.desc(), *fJson, os.Stderr)
	if len(errs) > 0 {
		os.Exit(1)
	}

	// dump formatted source
	outSrc := astprint.Print(prog, comments)
	if src.isStdin {
		_, _ = os.Stdout.WriteString(outSrc)
	} else if src.src != outSrc {
		if err = os.WriteFile(src.filename, []byte(outSrc), 0666); err != nil {
			return fmt.Errorf("writing to %s: %w", src.filename, err)
		}
	}

	return nil
}

func check(args []string) error {
	src, err := readSourceFromArgs(args)
	if err != nil {
		return err
	}

	// Only check errors; output to stdout
	_, _, errs := gaddis.Compile(src.src)
	reportErrors(errs, src.desc(), *fJson, os.Stdout)
	return nil
}

type runOpts struct {
	stopAfterBuild    bool
	isTest            bool
	leaveBuildOutputs bool
	goGen             bool
}

func run(args []string, opts runOpts) error {
	src, err := readSourceFromArgs(args)
	if err != nil {
		return err
	}

	prog, comments, errs := gaddis.Compile(src.src)
	reportErrors(errs, src.desc(), *fJson, os.Stdout)
	if len(errs) > 0 {
		os.Exit(1)
	}

	// auto format
	outSrc := astprint.Print(prog, comments)
	if !src.isStdin && src.src != outSrc {
		if err = os.WriteFile(src.filename, []byte(outSrc), 0666); err != nil {
			return fmt.Errorf("writing to %s: %w", src.filename, err)
		}
	}

	var cp *interp.Compilation
	var exeFile string
	ctx := context.Background()
	if !opts.goGen {
		cp = interp.Compile(prog)
		if opts.leaveBuildOutputs {
			asmFile := src.desc() + ".asm"
			var sb bytes.Buffer
			for i, inst := range cp.Code {
				si := inst.GetSourceInfo()
				line := si.Start.Line + 1
				text := strings.TrimSpace(src.src[si.Start.Pos:si.End.Pos])
				lhs := fmt.Sprintf("%3d: %s", i, inst)
				_, _ = fmt.Fprintf(&sb, "%-40s; %3d: %s\n", lhs, line, strings.SplitN(text, "\n", 2)[0])
			}
			if err := os.WriteFile(asmFile, sb.Bytes(), 0644); err != nil {
				return fmt.Errorf("writing to %s: %w", asmFile, err)
			}
		}

		if opts.stopAfterBuild {
			return nil
		}
	} else {
		goSrc := gogen.GoGenerate(prog, opts.isTest)

		var cancel context.CancelFunc
		ctx, cancel = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer cancel()

		br, err := goexec.Build(ctx, goSrc, src.desc())
		defer func() {
			if !opts.leaveBuildOutputs {
				if br.GoFile != "" {
					_ = os.Remove(br.GoFile)
				}
				if br.ExeFile != "" {
					_ = os.Remove(br.ExeFile)
				}
			}
		}()
		if err != nil {
			return fmt.Errorf("building %s: %w", src.desc(), err)
		}
		exeFile = br.ExeFile
	}

	if opts.stopAfterBuild {
		return nil
	}

	var terminalInput bool
	var input bytes.Buffer
	var stdin io.ReadCloser
	if opts.isTest {
		if inBytes, err := os.ReadFile(src.filename + ".in"); err == nil {
			// use input file as input
			terminalInput = false
			input.Write(inBytes)
			stdin = io.NopCloser(&input)
		}
	}
	if stdin == nil && src.isStdin {
		// there can be no input, we already consumed it all
		stdin = io.NopCloser(&input)
	}
	if stdin == nil {
		// capture input from terminal
		terminalInput = true
		r, w := io.Pipe()
		go func() {
			_, _ = io.Copy(w, os.Stdin)
			fmt.Println("closing")
			_ = w.Close()
		}()

		readCloser := struct {
			io.Reader
			io.Closer
		}{
			Reader: io.TeeReader(r, &input),
			Closer: w,
		}
		stdin = readCloser
	}

	// echo output to terminal if we need terminal input; or if we're not running test mode
	terminalOutput := terminalInput || !opts.isTest
	var output bytes.Buffer
	var errput bytes.Buffer

	var stdout io.Writer
	var stderr io.Writer
	if terminalOutput {
		// run on the terminal
		stdout = io.MultiWriter(os.Stdout, &output)
		stderr = io.MultiWriter(os.Stderr, &errput)
	} else {
		// run silent
		stdout = &output
		stderr = &errput
	}

	if cp != nil {
		var seed int64
		if !opts.isTest {
			seed = time.Now().UnixNano()
		}

		ec := &interp.ExecutionContext{
			Rng: rand.New(rand.NewSource(seed)),
			IoContext: builtins.IoContext{
				Stdin:  bufio.NewScanner(stdin),
				Stdout: gaddis.Synced(stdout),
				Stderr: gaddis.Synced(stderr),
			},
		}

		p := cp.NewProgram(ec)
		if err := p.Run(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			_, _ = fmt.Fprintln(os.Stderr, p.GetStackTrace(src.desc()))
			os.Exit(1)
		}
	} else {
		if err := goexec.Run(ctx, exeFile, stdin, stdout, stderr); err != nil {
			// If we were running silent and anything failed, spit the output to console
			if !terminalOutput {
				if output.Len() > 0 {
					_, _ = os.Stdout.Write(output.Bytes())
				}
				if errput.Len() > 0 {
					_, _ = os.Stderr.Write(errput.Bytes())
				}
			}
			return err
		}
	}

	// if we were running test mode and captured any input, dump it to an input file
	if opts.isTest && terminalInput && input.Len() > 0 {
		// drop the captured input into a
		_ = os.WriteFile(src.desc()+".in", input.Bytes(), 0644)
	}

	// if we were running test mode, either save or compare output
	if opts.isTest {
		gotOutput := output.Bytes()
		if expectOut, err := os.ReadFile(src.desc() + ".out"); err != nil {
			// dump the output we got, if any
			if len(gotOutput) > 0 {
				_ = os.WriteFile(src.desc()+".out", gotOutput, 0644)
			}
		} else if !bytes.Equal(output.Bytes(), expectOut) {
			// compare the output
			return fmt.Errorf("wrong output:\n%s", output.String())
		}
		fmt.Println("PASSED")
	}

	return nil
}

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
