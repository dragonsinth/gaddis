package builtins

import (
	"testing"
)

func TestModInteger(t *testing.T) {
	assertEqual(t, 0, ModInteger(5, 1))
	assertEqual(t, 1, ModInteger(5, 2))
	assertEqual(t, 2, ModInteger(5, 3))
	// TODO: zero, negative behavior spec?
}

func TestModReal(t *testing.T) {
	assertEqual(t, 0.5, ModReal(5.5, 1))
	assertEqual(t, 1.5, ModReal(5.5, 2))
	assertEqual(t, 2.5, ModReal(5.5, 3))
	// TODO: zero, negative behavior spec?
}

func TestExpInteger(t *testing.T) {
	assertEqual(t, 5, ExpInteger(5, 1))
	assertEqual(t, 25, ExpInteger(5, 2))
	assertEqual(t, 125, ExpInteger(5, 3))
	// TODO: zero, negative behavior spec?
}

func TestExpReal(t *testing.T) {
	assertEqual(t, 0.5, ExpReal(0.5, 1))
	assertEqual(t, 0.25, ExpReal(0.5, 2))
	assertEqual(t, 0.125, ExpReal(0.5, 3))
	// TODO: zero, negative behavior spec?
}

func assertEqual[T comparable](t *testing.T, want T, got T) {
	if want != got {
		t.Errorf("want=%v, got=%v", want, got)
	}
}
