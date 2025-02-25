package lib

import "time"
import "math/rand"

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func random(lo int64, hi int64) int64 {
	return lo + rng.Int63n(hi-lo+1)
}
