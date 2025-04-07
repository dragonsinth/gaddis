package asm

import (
	"runtime"
	"strings"
)

var (
	GitSha     string
	GoMod      = "github.com/dragonsinth/gaddis" // can be overridden
	ModuleBase = computeModuleBase()
)

func computeModuleBase() string {
	_, file, _, _ := runtime.Caller(0)
	if strings.HasSuffix(file, "/asm/util.go") {
		return strings.TrimSuffix(file, "/asm/util.go")
	} else if strings.HasSuffix(file, "\\asm\\util.go") {
		return strings.TrimSuffix(file, "\\asm\\util.go")
	} else {
		panic("cannot compute base path: " + file)
	}
}

func isLibFile(path string) bool {
	path = strings.TrimPrefix(path, ModuleBase)
	path = strings.Replace(path, "\\", "/", -1)
	return strings.HasPrefix(path, "/lib/")
}

func toFloat64(val any) float64 {
	switch v := val.(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	default:
		panic(v)
	}
}

type baseInst struct {
}

func (baseInst) Sym() string {
	return ""
}
