package lib

import (
	"testing"
)

func TestCurrencyFormat(t *testing.T) {
	for _, tc := range []struct {
		in   float64
		want string
	}{
		{0, "$0.00"},
		{0.99, "$0.99"},
		{1, "$1.00"},
		{10, "$10.00"},
		{12345.67891, "$12,345.68"},
		{123456.7891, "$123,456.79"},
		{1234567.891, "$1,234,567.89"},
		{-9876.54, "-$9,876.54"},
	} {
		got := currencyFormat(tc.in)
		if got != tc.want {
			t.Errorf("currencyFormat(%f) = %s, want %s", tc.in, got, tc.want)
		}
	}
}

func TestSubstring(t *testing.T) {
	for _, tc := range []struct {
		in         string
		start, end int64
		want       string
	}{
		{"", 0, -1, ""},
		{"a", 0, 0, "a"},
		{"ab", 0, 1, "ab"},
		{"abc", 1, 1, "b"},
		{"abcd", 1, 2, "bc"},
		{"abcd", 1, 3, "bcd"},
		{"abcd", 2, 3, "cd"},
	} {
		got := substring(tc.in, tc.start, tc.end)
		if got != tc.want {
			t.Errorf("substring(%s) = %s, want %s", tc.in, got, tc.want)
		}
	}
}
