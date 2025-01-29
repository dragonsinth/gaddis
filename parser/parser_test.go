package parser

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

func TestParse(t *testing.T) {
	block, err := Parse([]byte(program))
	if block != nil {
		for _, stmt := range block.Statements {
			t.Log(stmt.String())
		}
	}
	if err != nil {
		t.Error(err)
	}
}
