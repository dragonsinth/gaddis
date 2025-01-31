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

Declare Boolean flag
Set flag = price == quantity OR price != quantity
Set flag = price <= quantity AND price < quantity
Set flag = price >= quantity AND price > quantity
Set flag = flag AND flag
Set flag = flag OR flag
Set flag = NOT flag
Input flag
If flag Then
  Display True
Else
  Display False
End If
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
