package main_template

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

type syncWriter interface {
	io.Writer
	Sync() error
}

var stdin = bufio.NewReader(os.Stdin)
var stdout = syncWriter(os.Stdout)

func display(args ...any) {
	for _, arg := range args {
		switch typedArg := arg.(type) {
		case bool:
			if typedArg {
				arg = "True"
			} else {
				arg = "False"
			}
			// TODO: special formatting for floats maybe?
		}
		_, _ = fmt.Fprint(stdout, arg)
	}
	_, _ = fmt.Fprintln(stdout)
}

func readLine() string {
	_ = stdout.Sync() // ensure any prompts are flushed
	input, err := stdin.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	_ = stdout.Sync() // ensure user's newline is flushed to the terminal
	return strings.TrimSuffix(input, "\n")
}

func inputInteger() int64 {
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

func inputReal() float64 {
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

func inputString() string {
	_, _ = fmt.Fprint(stdout, "string> ")
	input := readLine()
	return input
}

func inputBoolean() bool {
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

func modInteger(a, b int64) int64 {
	return a % b
}

func expInteger(base, exp int64) int64 {
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

func modReal(a, b float64) float64 {
	return math.Mod(a, b)
}

func expReal(base, exp float64) float64 {
	return math.Pow(base, exp)
}
