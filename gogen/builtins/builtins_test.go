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

func TestStepInteger(t *testing.T) {
	var i int64
	assertEqual(t, true, StepInteger(&i, 2, 1))
	assertEqual(t, 1, i)
	assertEqual(t, true, StepInteger(&i, 2, 1))
	assertEqual(t, 2, i)
	assertEqual(t, false, StepInteger(&i, 2, 1))
	assertEqual(t, 3, i)

	assertEqual(t, true, StepInteger(&i, 1, -1))
	assertEqual(t, 2, i)
	assertEqual(t, true, StepInteger(&i, 1, -1))
	assertEqual(t, 1, i)
	assertEqual(t, false, StepInteger(&i, 1, -1))
	assertEqual(t, 0, i)
}

func TestStepReal(t *testing.T) {
	var i float64
	assertEqual(t, true, StepReal(&i, 2, 1))
	assertEqual(t, 1, i)
	assertEqual(t, true, StepReal(&i, 2, 1))
	assertEqual(t, 2, i)
	assertEqual(t, false, StepReal(&i, 2, 1))
	assertEqual(t, 3, i)

	assertEqual(t, true, StepReal(&i, 1, -1))
	assertEqual(t, 2, i)
	assertEqual(t, true, StepReal(&i, 1, -1))
	assertEqual(t, 1, i)
	assertEqual(t, false, StepReal(&i, 1, -1))
	assertEqual(t, 0, i)
}

func assertEqual[T comparable](t *testing.T, want T, got T) {
	if want != got {
		t.Errorf("want=%v, got=%v", want, got)
	}
}
