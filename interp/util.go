package interp

func toFloat64(val any) float64 {
	switch v := val.(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	default:
		panic(v)
	}
}
