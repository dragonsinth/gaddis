package lib

import _ "embed"

//go:embed lib.go
var Source string

var External = map[string]any{
	"random": random,
	"sqrt":   sqrt,
	"pow":    pow,
	"abs":    abs,
	"cos":    cos,
	"round":  round,
	"sin":    sin,
	"tan":    tan,

	"toInteger": toInteger,
	"toReal":    toReal,

	"currencyFormat": currencyFormat,

	"length":          length,
	"append":          appendString,
	"toUpper":         toUpper,
	"toLower":         toLower,
	"substring":       substring,
	"contains":        contains,
	"stringToInteger": stringToInteger,
	"stringToReal":    stringToReal,
	"isInteger":       isInteger,
	"isReal":          isReal,
}
