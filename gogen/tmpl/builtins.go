package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

var stdin = bufio.NewReader(os.Stdin)

func display(args ...any) {
	for _, arg := range args {
		fmt.Print(arg)
	}
	fmt.Println()
}

func readLine() string {
	_ = os.Stdout.Sync() // ensure any prompts are flushed
	input, err := stdin.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	_ = os.Stdout.Sync() // ensure user's newline is flushed to the terminal
	return strings.TrimSuffix(input, "\n")
}

func inputInteger() int64 {
	for {
		fmt.Print("integer> ")
		input := readLine()
		v, err := strconv.ParseInt(input, 10, 64)
		if err == nil {
			return v
		}
		fmt.Println("error, invalid integer, try again")
	}
}

func inputReal() float64 {
	for {
		fmt.Print("real> ")
		input := readLine()
		v, err := strconv.ParseFloat(input, 64)
		if err == nil {
			return v
		}
		fmt.Println("error, invalid real, try again")
	}
}

func inputString() string {
	fmt.Print("string> ")
	input := readLine()
	return input
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

func main() {
}
