package builtins

import "bytes"
import "bufio"
import "fmt"
import "io"
import "log"
import "math"
import "os"
import "strconv"

type String = []byte

func Display(args ...any) {
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
		default:
			_, _ = fmt.Fprint(&sb, arg)
		}
	}
	sb.WriteByte('\n')
	_, _ = stdout.Write(sb.Bytes())
}

func InputInteger() int64 {
	for {
		_, _ = fmt.Fprint(stdout, "integer> ")
		input := readLine()
		v, err := strconv.ParseInt(string(input), 10, 64)
		if err == nil {
			return v
		}
		_, _ = fmt.Fprintln(stdout, "error, invalid integer, try again")
	}
}

func InputReal() float64 {
	for {
		_, _ = fmt.Fprint(stdout, "real> ")
		input := readLine()
		v, err := strconv.ParseFloat(string(input), 64)
		if err == nil {
			return v
		}
		_, _ = fmt.Fprintln(stdout, "error, invalid real, try again")
	}
}

func InputString() String {
	_, _ = fmt.Fprint(stdout, "string> ")
	input := readLine()
	return input
}

func InputBoolean() bool {
	for {
		_, _ = fmt.Fprint(stdout, "boolean> ")
		input := readLine()
		v, err := strconv.ParseBool(string(input))
		if err == nil {
			return v
		}
		_, _ = fmt.Fprintln(stdout, "error, invalid boolean, try again")
	}
}

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

func StepInteger(ref int64, stop int64, step int64) bool {
	if step < 0 {
		return ref >= stop
	} else {
		return ref <= stop
	}
}

func StepReal(ref float64, stop float64, step float64) bool {
	if step < 0 {
		return ref >= stop
	} else {
		return ref <= stop
	}
}

type syncWriter interface {
	io.Writer
	Sync() error
}

var stdin = bufio.NewScanner(os.Stdin)
var stdout = syncWriter(os.Stdout)

func readLine() String {
	_ = stdout.Sync() // ensure any prompts are flushed
	if !stdin.Scan() {
		log.Fatal(io.EOF)
	}
	input, err := stdin.Bytes(), stdin.Err()
	if err != nil {
		log.Fatal(err)
	}
	_ = stdout.Sync() // ensure user's newline is flushed to the terminal
	return input
}

// Tab keyword.
var Tab = String("\t")

type tabDisplay struct{}

// TabDisplay is "Magic" when passed directly to [Builtins.Display].
var TabDisplay = tabDisplay{}
