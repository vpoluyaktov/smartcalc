package eval

import (
	"fmt"
	"math"
)

func callFn(name string, x float64) (float64, error) {
	switch name {
	case "sin":
		return math.Sin(x), nil
	case "cos":
		return math.Cos(x), nil
	case "tan":
		return math.Tan(x), nil
	case "asin":
		return math.Asin(x), nil
	case "acos":
		return math.Acos(x), nil
	case "atan":
		return math.Atan(x), nil
	case "sqrt":
		return math.Sqrt(x), nil
	case "abs":
		return math.Abs(x), nil
	case "ln":
		return math.Log(x), nil
	case "log":
		return math.Log10(x), nil
	default:
		return 0, fmt.Errorf("unknown function: %s", name)
	}
}
