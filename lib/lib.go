package lib

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type RandContext struct {
	Rng *rand.Rand
}

func (ctx RandContext) random(lo int64, hi int64) int64 {
	return lo + ctx.Rng.Int63n(hi-lo+1)
}

var (
	sqrt  = math.Sqrt
	pow   = math.Pow
	abs   = math.Abs
	cos   = math.Cos
	round = math.Round
	sin   = math.Sin
	tan   = math.Tan
)

func toInteger(x float64) int64 {
	return int64(x)
}

func toReal(x int64) float64 {
	return float64(x)
}

// not specced!
func toString(arg any) string {
	if arg == nil {
		return "<nil>"
	}
	switch x := arg.(type) {
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		if x {
			return "True"
		} else {
			return "False"
		}
	case tabDisplay:
		return "\t"
	case []byte:
		panic("should not get")
	case string:
		return x
	case byte:
		return string([]byte{x})
	case []any:
		var sb strings.Builder
		sb.WriteByte('[')
		for i, e := range x {
			if i > 0 {
				sb.WriteByte(',')
			}
			sb.WriteString(toString(e))
		}
		sb.WriteByte(']')
		return sb.String()
	default:
		return fmt.Sprint(arg)
	}
}

func currencyFormat(amount float64) string {
	var sb bytes.Buffer
	cents := int64(math.Round(amount * 100))
	if cents < 0 {
		sb.WriteByte('-')
		cents = -cents
	}
	dollars := cents / 100
	pennies := byte(cents % 100)

	sb.WriteByte('$')
	if dollars == 0 {
		sb.WriteByte('0')
	} else {
		str := strconv.FormatInt(dollars, 10)
		first := true
		for str != "" {
			if first {
				first = false
			} else {
				sb.WriteByte(',')
			}
			count := len(str) % 3
			if count == 0 {
				count = 3
			}
			sb.WriteString(str[:count])
			str = str[count:]
		}
	}
	sb.WriteByte('.')
	sb.WriteByte('0' + pennies/10)
	sb.WriteByte('0' + pennies%10)
	return sb.String()
}

func length(s string) int64 {
	return int64(len(s))
}

func appendString(a, b string) string {
	return a + b
}

var (
	toUpper = strings.ToUpper
	toLower = strings.ToLower
)

func substring(s string, start int64, end int64) string {
	return s[start : end+1]
}

func insertString(s string, pos int64, add string) string {
	lhs := s[:pos]
	rhs := s[pos:]
	return lhs + add + rhs
}

func deleteString(s string, start int64, end int64) string {
	if end+1 < start {
		panic("delete: invalid range start(%d) should be less than or equal to end (%d)")
	}
	lhs := s[:start]
	rhs := s[end+1:]
	return lhs + rhs
}

var contains = strings.Contains

func stringToInteger(s string) int64 {
	v, err := strconv.ParseInt(string(s), 10, 64)
	if err != nil {
		panic(err)
	}
	return v
}

func stringToReal(s string) float64 {
	v, err := strconv.ParseFloat(string(s), 64)
	if err != nil {
		panic(err)
	}
	return v
}

func isInteger(s string) bool {
	_, err := strconv.ParseInt(string(s), 10, 64)
	return err == nil
}

func isReal(s string) bool {
	_, err := strconv.ParseFloat(string(s), 64)
	return err == nil
}

func isDigit(c byte) bool {
	return unicode.IsDigit(rune(c))
}

func isLetter(c byte) bool {
	return unicode.IsLetter(rune(c))
}

func isLower(c byte) bool {
	return unicode.IsLower(rune(c))
}

func isUpper(c byte) bool {
	return unicode.IsUpper(rune(c))
}

func isWhitespace(c byte) bool {
	return unicode.IsSpace(rune(c))
}

func stringWithCharUpdate(str string, idx int64, c byte) string {
	buf := []byte(str)
	buf[idx] = c
	return string(buf)
}

// BELOW: Used only by the gogen runtime.

func stringWithCharUpdateRef(s *string, idx int64, c byte) {
	*s = stringWithCharUpdate(*s, idx, c)
}

func insertStringRef(s *string, pos int64, add string) {
	*s = insertString(*s, pos, add)
}

func deleteStringRef(s *string, start int64, end int64) {
	*s = deleteString(*s, start, end)
}

var (
	randCtx = &RandContext{
		Rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	random = randCtx.random

	_ = stringWithCharUpdateRef
	_ = deleteStringRef
	_ = insertStringRef
)
