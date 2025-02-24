package main_template

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type builtin struct{}

var Builtin = builtin{}

func (builtin) Display(args ...any) {
	var sb strings.Builder
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
		// TODO: special formatting for floats maybe?
		default:
			fmt.Fprint(&sb, arg)
		}
	}
	_, _ = fmt.Fprintln(stdout, sb.String())
}

func (builtin) InputInteger() int64 {
	for {
		_, _ = fmt.Fprint(stdout, "integer> ")
		input := readLine()
		v, err := strconv.ParseInt(input, 10, 64)
		if err == nil {
			return v
		}
		_, _ = fmt.Fprintln(stdout, "error, invalid integer, try again")
	}
}

func (builtin) InputReal() float64 {
	for {
		_, _ = fmt.Fprint(stdout, "real> ")
		input := readLine()
		v, err := strconv.ParseFloat(input, 64)
		if err == nil {
			return v
		}
		_, _ = fmt.Fprintln(stdout, "error, invalid real, try again")
	}
}

func (builtin) InputString() string {
	_, _ = fmt.Fprint(stdout, "string> ")
	input := readLine()
	return input
}

func (builtin) InputBoolean() bool {
	for {
		_, _ = fmt.Fprint(stdout, "boolean> ")
		input := readLine()
		v, err := strconv.ParseBool(input)
		if err == nil {
			return v
		}
		_, _ = fmt.Fprintln(stdout, "error, invalid boolean, try again")
	}
}

func (builtin) ModInteger(a, b int64) int64 {
	return a % b
}

func (builtin) ExpInteger(base, exp int64) int64 {
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

func (builtin) ModReal(a, b float64) float64 {
	return math.Mod(a, b)
}

func (builtin) ExpReal(base, exp float64) float64 {
	return math.Pow(base, exp)
}

func (builtin) StepInteger(ref int64, stop int64, step int64) bool {
	if step < 0 {
		return ref >= stop
	} else {
		return ref <= stop
	}
}

func (builtin) StepReal(ref float64, stop float64, step float64) bool {
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

func readLine() string {
	_ = stdout.Sync() // ensure any prompts are flushed
	if !stdin.Scan() {
		log.Fatal(io.EOF)
	}
	input, err := stdin.Text(), stdin.Err()
	if err != nil {
		log.Fatal(err)
	}
	_ = stdout.Sync() // ensure user's newline is flushed to the terminal
	return input
}

// Literal string Tab keyword.
const Tab = "\t"

type tabDisplay struct{}

// "Magic" Tab keyword when passed directly to [Builtins.Display].
var TabDisplay = tabDisplay{}

type lib struct{}

var Lib = lib{}

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func (lib) Random(lo int64, hi int64) int64 {
	return lo + random.Int63n(hi-lo+1)
}
