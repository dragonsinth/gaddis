package builtins

import (
	"math"
	"reflect"
)

// Easier codegen.

type Integer = int64
type Real = float64
type String = string
type Character = byte
type Boolean = bool

func ModInteger(a, b Integer) Integer {
	return a % b
}

func ExpInteger(base, exp Integer) Integer {
	if exp < 0 {
		return 0 // Or handle negative exponents as needed (e.g., return 1 / intPow(base, -exp))
	}
	if exp == 0 {
		return 1
	}

	result := Integer(1)
	for {
		if exp&1 == 1 { // Check if the least significant bit of exp is 1
			result *= base
		}
		exp >>= 1 // Right shift exp (equivalent to dividing by 2)
		if exp == 0 {
			break
		}
		base *= base // Square the base
	}
	return result
}

func ModReal(a, b Real) Real {
	return math.Mod(a, b)
}

func ExpReal(base, exp Real) Real {
	return math.Pow(base, exp)
}

func StepInteger(ref *Integer, stop, step Integer) Boolean {
	*ref += step
	if step < 0 {
		return *ref >= stop
	} else {
		return *ref <= stop
	}
}

func StepReal(ref *Real, stop, step Real) Boolean {
	*ref += step
	if step < 0 {
		return *ref >= stop
	} else {
		return *ref <= stop
	}
}

func Clone[T any](slice []T) []T {
	ret := make([]T, len(slice))
	elemType := reflect.TypeOf(slice).Elem()
	if elemType.Kind() == reflect.Slice {
		for i := range slice {
			ret[i] = reflectClone(reflect.ValueOf(slice[i])).Interface().(T)
		}
	} else {
		copy(ret, slice)
	}
	return ret
}

func reflectClone(slice reflect.Value) reflect.Value {
	ret := reflect.MakeSlice(slice.Type(), slice.Len(), slice.Cap())
	elemType := slice.Type().Elem()
	if elemType.Kind() == reflect.Slice {
		for i := 0; i < slice.Len(); i++ {
			ret.Index(i).Set(reflectClone(slice.Index(i)))
		}
	} else {
		reflect.Copy(ret, slice)
	}
	return ret
}
