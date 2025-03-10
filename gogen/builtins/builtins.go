package builtins

import "bytes"
import "bufio"
import "fmt"
import "io"
import "math"
import "os"
import "strconv"

type String = []byte

type SyncWriter interface {
	io.Writer
	Sync() error
}

type IoContext struct {
	Stdin   *bufio.Scanner
	Stdout  SyncWriter
	WorkDir string
}

func (ctx IoContext) Display(args ...any) {
	var sb bytes.Buffer
	tabCount := 0
	for _, arg := range args {
		switch typedArg := arg.(type) {
		case bool:
			if typedArg {
				sb.WriteString("True")
			} else {
				sb.WriteString("False")
			}
		case tabDisplay:
			tabCount++
			for sb.Len() < 8*tabCount {
				sb.WriteByte(' ')
			}
		case string:
			panic(typedArg) // should be impossible
		case String:
			sb.Write(typedArg)
		// TODO: special formatting for floats maybe?
		case byte:
			sb.WriteByte(typedArg)
		default:
			_, _ = fmt.Fprint(&sb, arg)
		}
	}
	sb.WriteByte('\n')
	_, _ = ctx.Stdout.Write(sb.Bytes())
	_ = ctx.Stdout.Sync()
}

func (ctx IoContext) InputInteger() int64 {
	for {
		_, _ = fmt.Fprint(ctx.Stdout, "integer> ")
		input := ctx.readLine()
		v, err := strconv.ParseInt(string(input), 10, 64)
		if err == nil {
			return v
		}
		_, _ = fmt.Fprintln(ctx.Stdout, "error, invalid integer, try again")
	}
}

func (ctx IoContext) InputReal() float64 {
	for {
		_, _ = fmt.Fprint(ctx.Stdout, "real> ")
		input := ctx.readLine()
		v, err := strconv.ParseFloat(string(input), 64)
		if err == nil {
			return v
		}
		_, _ = fmt.Fprintln(ctx.Stdout, "error, invalid real, try again")
	}
}

func (ctx IoContext) InputString() String {
	_, _ = fmt.Fprint(ctx.Stdout, "string> ")
	input := ctx.readLine()
	return input
}

func (ctx IoContext) InputCharacter() byte {
	for {
		_, _ = fmt.Fprint(ctx.Stdout, "character> ")
		input := ctx.readLine()
		if len(input) == 1 {
			return input[0]
		}
		_, _ = fmt.Fprintln(ctx.Stdout, "error, input exactly 1 character, try again")
	}
}

func (ctx IoContext) InputBoolean() bool {
	for {
		_, _ = fmt.Fprint(ctx.Stdout, "boolean> ")
		input := ctx.readLine()
		v, err := strconv.ParseBool(string(input))
		if err == nil {
			return v
		}
		_, _ = fmt.Fprintln(ctx.Stdout, "error, invalid boolean, try again")
	}
}

func (ctx IoContext) readLine() String {
	_ = ctx.Stdout.Sync() // ensure any prompts are flushed
	if !ctx.Stdin.Scan() {
		panic(io.EOF)
	}
	input, err := ctx.Stdin.Bytes(), ctx.Stdin.Err()
	if err != nil {
		panic(err)
	}
	_ = ctx.Stdout.Sync() // ensure user's newline is flushed to the terminal
	return input
}

var defaultCtx = IoContext{
	Stdin:   bufio.NewScanner(os.Stdin),
	Stdout:  os.Stdout,
	WorkDir: ".",
}

var (
	Display        = defaultCtx.Display
	InputInteger   = defaultCtx.InputInteger
	InputReal      = defaultCtx.InputReal
	InputString    = defaultCtx.InputString
	InputCharacter = defaultCtx.InputCharacter
	InputBoolean   = defaultCtx.InputBoolean
)

func ModInteger(a, b int64) int64 {
	return a % b
}

func ExpInteger(base, exp int64) int64 {
	if exp < 0 {
		return 0 // Or handle negative exponents as needed (e.g., return 1 / intPow(base, -exp))
	}
	if exp == 0 {
		return 1
	}

	result := int64(1)
	for {
		if exp&1 == 1 { // Check if the least significant bit of exp is 1
			result *= base
		}
		exp >>= 1 // Right shift exp (equivalent to dividing by 2)
		if exp == 0 {
			break
		}
		base *= base // Square the base
	}
	return result
}

func ModReal(a, b float64) float64 {
	return math.Mod(a, b)
}

func ExpReal(base, exp float64) float64 {
	return math.Pow(base, exp)
}

func ForInteger(ref *int64, start, stop, step int64) bool {
	*ref = start
	if step < 0 {
		return *ref >= stop
	} else {
		return *ref <= stop
	}
}

func StepInteger(ref *int64, stop, step int64) bool {
	*ref += step
	if step < 0 {
		return *ref >= stop
	} else {
		return *ref <= stop
	}
}

func ForReal(ref *float64, start, stop, step float64) bool {
	*ref = start
	if step < 0 {
		return *ref >= stop
	} else {
		return *ref <= stop
	}
}

func StepReal(ref *float64, stop, step float64) bool {
	*ref += step
	if step < 0 {
		return *ref >= stop
	} else {
		return *ref <= stop
	}
}

type tabDisplay struct{}

// TabDisplay is "Magic" when passed directly to [Builtins.Display].
var TabDisplay = tabDisplay{}
