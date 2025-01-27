package ast

import "fmt"

func stringArray[T any](a []T) []string {
	var ret []string
	for _, v := range a {
		ret = append(ret, fmt.Sprintf("%v", v))
	}
	return ret
}
