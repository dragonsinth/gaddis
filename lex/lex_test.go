package lex

import (
	"testing"
)

const program = `Constant Real TAX_RATE = 0.5
Declare Integer price, quantity, subtotal
Display "Input price:"
Input price
Display "Input quantity:"
Input quantity
Set subtotal = price * quantity
Display "Subtotal:", subtotal
Display "Total:", subtotal + subtotal * TAX_RATE
`

func TestLex(t *testing.T) {
	lexer := NewLexer([]byte(program))
	for {
		r := lexer.Lex()
		t.Logf("%d:%d\t%s\t%s\n", r.Pos.Line, r.Pos.Column, r.Token, r.Text)
		if r.Token == EOF {
			break
		} else if r.Token == ILLEGAL {
			t.Error(r.Error)
		}
	}
}
