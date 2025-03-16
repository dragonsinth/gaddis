package lib

import (
	_ "embed"
	"path/filepath"
	"reflect"
)

type Func struct {
	Name    string
	FuncPtr reflect.Value
}

func CreateLibrary(io IoContext, rng RandContext) []Func {
	entries := getEntries()
	for i, v := range []any{
		io.Display,
		io.InputInteger,
		io.InputReal,
		io.InputString,
		io.InputCharacter,
		io.InputBoolean,
		rng.random,
	} {
		entries[i].funcPtr = v
	}

	ret := make([]Func, len(entries))
	for i, v := range entries {
		if v.funcPtr == nil {
			panic(v.name)
		}
		ret[i] = Func{
			Name:    v.name,
			FuncPtr: reflect.ValueOf(v.funcPtr),
		}
	}
	return ret
}

func GetLibrary() []Func {
	entries := getEntries()

	var ret []Func
	for _, v := range entries {
		if v.name == "random" {
			v.funcPtr = random
		} else if v.funcPtr == nil {
			continue
		}
		ret = append(ret, Func{
			Name:    v.name,
			FuncPtr: reflect.ValueOf(v.funcPtr),
		})
	}
	return ret
}

type entry struct {
	name    string
	funcPtr any
}

func getEntries() []entry {
	return []entry{
		{"Display", nil},
		{"InputInteger", nil},
		{"InputReal", nil},
		{"InputString", nil},
		{"InputCharacter", nil},
		{"InputBoolean", nil},

		{"random", nil},

		{"sqrt", sqrt},
		{"pow", pow},
		{"abs", abs},
		{"cos", cos},
		{"round", round},
		{"sin", sin},
		{"tan", tan},

		{"toInteger", toInteger},
		{"toReal", toReal},

		{"integerFormat", integerFormat},
		{"realFormat", realFormat},
		{"currencyFormat", currencyFormat},

		{"length", length},
		{"append", appendString},
		{"toUpper", toUpper},
		{"toLower", toLower},
		{"substring", substring},
		{"contains", contains},

		{"stringToInteger", stringToInteger},
		{"stringToReal", stringToReal},
		{"isInteger", isInteger},
		{"isReal", isReal},
	}
}

var indexMap = mapEntries()

func mapEntries() map[string]int {
	ret := map[string]int{}
	for i, v := range getEntries() {
		ret[v.name] = i
	}
	return ret
}

func IndexOf(name string) int {
	i, ok := indexMap[name]
	if !ok {
		panic(name)
	}
	return i
}

// BELOW: Used only by the gogen runtime.

//go:embed io.go
var IoSource string

//go:embed lib.go
var LibSource string

type LibSrc struct {
	Name string
	Src  string
	Id   int
}

var libSources = []LibSrc{
	{"io.go", IoSource, 1000},
	{"lib.go", LibSource, 2000},
}

func SrcByName(filename string) *LibSrc {
	base := filepath.Base(filename)
	for _, src := range libSources {
		if src.Name == base {
			return &src
		}
	}
	return nil
}

func SrcById(id int) *LibSrc {
	for _, src := range libSources {
		if src.Id == id {
			return &src
		}
	}
	return nil
}
