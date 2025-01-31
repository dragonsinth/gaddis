package lex

import "testing"

const program = `// Start
Constant Real TAX_RATE = 0.5
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

func TestLex(t *testing.T) {
	lexer := NewLexer([]byte(program))
	for {
		r := lexer.Lex()
		t.Logf("%d:%d\t%s\t%s\n", r.Pos.Line, r.Pos.Column, r.Token, r.Text)
		if r.Token == EOF {
			break
		} else if r.Token == ILLEGAL {
			t.Fatal(r.Error)
		}
	}
}
