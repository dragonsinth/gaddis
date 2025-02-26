package lib

import "math"
import "strconv"
import "strings"
import "time"
import "math/rand"

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func random(lo int64, hi int64) int64 {
	return lo + rng.Int63n(hi-lo+1)
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

func currencyFormat(amount float64) string {
	var sb strings.Builder
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

func append(a, b string) string {
	return a + b
}

var (
	toUpper = strings.ToUpper
	toLower = strings.ToLower
)

func substring(s string, start int64, end int64) string {
	return s[start : end+1]
}

var contains = strings.Contains

func stringToInteger(s string) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return v
}

func stringToReal(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return v
}

func isInteger(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

func isReal(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
