package lib

import _ "embed"

//go:embed lib.go
var Source string

var External = map[string]any{
	"random": random,
}
