package main

import (
	"flag"
	"github.com/dragonsinth/gaddis/parser"
	"log"
	"os"
)

func main() {
	flag.Parse()

	var input []byte
	for _, fname := range flag.Args() {
		buf, err := os.ReadFile(fname)
		if err != nil {
			log.Fatalf("reading %s: %v\n", fname, err)
		}
		input = append(input, buf...)
	}

	block, err := parser.Parse(input)
	if err != nil {
		log.Fatal(err)
	}
	for _, stmt := range block.Statements {
		log.Println(stmt.String())
	}
}
