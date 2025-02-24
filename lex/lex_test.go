package lex

import (
	_ "embed"
	"testing"
)

var (
	//go:embed lex_test.gad
	program string
)

func TestLex(t *testing.T) {
	lexer := New(program)
	for {
		r := lexer.Lex()
		t.Logf("%d:%d:%d\t%s\t%s\n", r.Pos.Pos, r.Pos.Line, r.Pos.Column, r.Token, r.Text)
		if r.Token == EOF {
			break
		} else if r.Token == ILLEGAL {
			t.Fatal(r.Error)
		}
	}
}
