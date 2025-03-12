package lib

import (
	"bytes"
	"math"
	"math/rand"
	"strconv"
	"time"
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

func currencyFormat(amount float64) []byte {
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
	return sb.Bytes()
}

func length(s []byte) int64 {
	return int64(len(s))
}

func appendString(a, b []byte) []byte {
	// make a copy
	ret := make([]byte, 0, len(a)+len(b))
	ret = append(ret, a...)
	ret = append(ret, b...)
	return ret
}

var (
	toUpper = bytes.ToUpper
	toLower = bytes.ToLower
)

func substring(s []byte, start int64, end int64) []byte {
	return s[start : end+1]
}

var contains = bytes.Contains

func stringToInteger(s []byte) int64 {
	v, err := strconv.ParseInt(string(s), 10, 64)
	if err != nil {
		panic(err)
	}
	return v
}

func stringToReal(s []byte) float64 {
	v, err := strconv.ParseFloat(string(s), 64)
	if err != nil {
		panic(err)
	}
	return v
}

func isInteger(s []byte) bool {
	_, err := strconv.ParseInt(string(s), 10, 64)
	return err == nil
}

func isReal(s []byte) bool {
	_, err := strconv.ParseFloat(string(s), 64)
	return err == nil
}

// BELOW: Used only by the gogen runtime.

var (
	randCtx = &RandContext{
		Rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	random = randCtx.random
)
