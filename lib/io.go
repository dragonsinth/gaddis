package lib

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type IoProvider interface {
	Input() (string, error)
	Output(string)
	Dir() string
}

type ioContext struct {
	provider IoProvider
}

func (ctx ioContext) Display(args ...any) {
	var sb bytes.Buffer
	tabCount := 0
	for _, arg := range args {
		switch typedArg := arg.(type) {
		case tabDisplay:
			tabCount++
			for sb.Len() < 8*tabCount {
				sb.WriteByte(' ')
			}
		default:
			sb.WriteString(toString(typedArg))
		}
	}
	sb.WriteRune('\n')
	ctx.provider.Output(sb.String())
}

func (ctx ioContext) InputInteger() int64 {
	for {
		ctx.provider.Output("integer> ")
		input := ctx.readLine()
		v, err := strconv.ParseInt(input, 10, 64)
		if err == nil {
			return v
		}
		ctx.provider.Output("error, invalid integer, try again\n")
	}
}

func (ctx ioContext) InputReal() float64 {
	for {
		ctx.provider.Output("real> ")
		input := ctx.readLine()
		v, err := strconv.ParseFloat(input, 64)
		if err == nil {
			return v
		}
		ctx.provider.Output("error, invalid real, try again\n")
	}
}

func (ctx ioContext) InputString() string {
	ctx.provider.Output("string> ")
	input := ctx.readLine()
	return input
}

func (ctx ioContext) InputCharacter() byte {
	for {
		ctx.provider.Output("character> ")
		input := ctx.readLine()
		if len(input) == 1 {
			return input[0]
		}
		ctx.provider.Output("error, input exactly 1 character, try again\n")
	}
}

func (ctx ioContext) InputBoolean() bool {
	for {
		ctx.provider.Output("boolean> ")
		input := ctx.readLine()
		v, err := strconv.ParseBool(input)
		if err == nil {
			return v
		}
		ctx.provider.Output("error, invalid boolean, try again\n")
	}
}

func (ctx ioContext) readLine() string {
	in, err := ctx.provider.Input()
	if err != nil {
		panic(err)
	}
	return in
}

func (ctx ioContext) OpenOutputFile(file OutputFile, name string) OutputFile {
	if file.File != nil {
		panic("file already open")
	}
	filename := filepath.Join(ctx.provider.Dir(), string(name))
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	return OutputFile{File: f}
}

func (ctx ioContext) OpenAppendFile(file OutputFile, name string) OutputFile {
	if file.File != nil {
		panic("file already open")
	}
	filename := filepath.Join(ctx.provider.Dir(), string(name))
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	return OutputFile{File: f, IsAppend: true}
}

func (ctx ioContext) OpenInputFile(file InputFile, name string) InputFile {
	if file.File != nil {
		panic("file already open")
	}
	filename := filepath.Join(ctx.provider.Dir(), string(name))
	f, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	return InputFile{File: f, Reader: bufio.NewReader(f)}
}

func (ctx ioContext) DeleteFile(name string) {
	filename := filepath.Join(ctx.provider.Dir(), string(name))
	err := os.Remove(filename)
	if err != nil {
		panic(err)
	}
}

func (ctx ioContext) RenameFile(oldName string, newName string) {
	oldFilename := filepath.Join(ctx.provider.Dir(), string(oldName))
	newFilename := filepath.Join(ctx.provider.Dir(), string(newName))
	err := os.Rename(oldFilename, newFilename)
	if err != nil {
		panic(err)
	}
}

func CloseOutputFile(file OutputFile) {
	if file.File == nil {
		panic("file not open")
	}
	err := file.File.Close()
	file.File = nil
	if err != nil {
		panic(err)
	}
}

func CloseInputFile(file InputFile) {
	if file.File == nil {
		panic("file not open")
	}
	err := file.File.Close()
	file.File = nil
	file.Reader = nil
	if err != nil {
		panic(err)
	}
}

func WriteFile(of OutputFile, args ...any) {
	if of.File == nil {
		panic("file not open")
	}
	file := of.File
	for _, arg := range args {
		var err error
		switch typedArg := arg.(type) {
		case bool:
			if typedArg {
				_, err = file.WriteString("True")
			} else {
				_, err = file.WriteString("False")
			}
		case string:
			_, err = file.WriteString(strconv.Quote(string(typedArg)))
		case byte:
			_, err = file.WriteString(strconv.QuoteRune(rune(typedArg)))
		case int64:
			_, err = file.WriteString(strconv.FormatInt(typedArg, 10))
		case float64:
			_, err = file.WriteString(strconv.FormatFloat(typedArg, 'g', -1, 64))
		default:
			panic(typedArg)
		}
		if err != nil {
			panic(err)
		}
		_, err = file.WriteString("\n")
		if err != nil {
			panic(err)
		}
	}
}

func ReadInteger(file InputFile) int64 {
	if file.File == nil {
		panic("file not open")
	}
	input := scanLine(file)
	v, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		panic(err)
	}
	return v
}

func ReadReal(file InputFile) float64 {
	if file.File == nil {
		panic("file not open")
	}
	input := scanLine(file)
	v, err := strconv.ParseFloat(input, 64)
	if err != nil {
		panic(err)
	}
	return v
}

func ReadString(file InputFile) string {
	input := scanLine(file)
	v, err := strconv.Unquote(input)
	if err != nil {
		panic(err)
	}
	return v
}

func ReadCharacter(file InputFile) byte {
	input := scanLine(file)
	v, err := strconv.Unquote(input)
	if err != nil {
		panic(err)
	}
	if len(v) != 0 {
		panic("invalid character")
	}
	return v[0]
}

func ReadBoolean(file InputFile) bool {
	input := scanLine(file)
	v, err := strconv.ParseBool(input)
	if err != nil {
		panic(err)
	}
	return v
}

func scanLine(file InputFile) string {
	if file.File == nil {
		panic("file not open")
	}
	v, err := file.Reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(v)
}

func eof(file InputFile) bool {
	if file.File == nil {
		panic("file not open")
	}
	_, err := file.Reader.Peek(1)
	return err == io.EOF
}

type tabDisplay struct{}

// TabDisplay is "Magic" when passed directly to [Builtins.Display].
var TabDisplay = tabDisplay{}

type OutputFile struct {
	File     *os.File
	IsAppend bool
}

func (of *OutputFile) String() string {
	if of == nil {
		return "<nil>"
	}
	if of.IsAppend {
		return "<OutputFile AppendMode>"
	} else {
		return "<OutputFile>"
	}
}

type AppendFile = OutputFile

type InputFile struct {
	File   *os.File
	Reader *bufio.Reader
}

func (of *InputFile) String() string {
	if of == nil {
		return "<nil>"
	}
	return "<InputFile>"
}

// BELOW: Used only by the gogen runtime.

type defaultIo struct {
	in *bufio.Scanner
}

func (dio defaultIo) Input() (string, error) {
	if !dio.in.Scan() {
		return "", io.EOF
	}
	input, err := dio.in.Text(), dio.in.Err()
	if err != nil {
		return "", err
	}
	return input, nil
}

func (dio defaultIo) Output(text string) {
	_, _ = os.Stdout.WriteString(text)
	_ = os.Stdout.Sync()
}

func (dio defaultIo) Dir() string {
	return "."
}

var (
	ioCtx = ioContext{provider: defaultIo{in: bufio.NewScanner(os.Stdin)}}

	Display        = ioCtx.Display
	InputInteger   = ioCtx.InputInteger
	InputReal      = ioCtx.InputReal
	InputString    = ioCtx.InputString
	InputCharacter = ioCtx.InputCharacter
	InputBoolean   = ioCtx.InputBoolean

	OpenOutputFile = ioCtx.OpenOutputFile
	OpenAppendFile = ioCtx.OpenAppendFile
	OpenInputFile  = ioCtx.OpenInputFile
	DeleteFile     = ioCtx.DeleteFile
	RenameFile     = ioCtx.RenameFile
)
