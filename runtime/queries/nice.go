package queries

import (
	"math"
)

var (
	e10 = math.Sqrt(50)
	e5  = math.Sqrt(10)
	e2  = math.Sqrt(2)
)

func calculateE(stepError float64) float64 {
	if stepError >= e10 {
		return 10
	} else if stepError >= e5 {
		return 5
	} else if stepError >= e2 {
		return 2
	}
	return 1
}

func tickIncrement(start, stop, count float64) float64 {
	step := (stop - start) / math.Max(0, count)
	power := math.Floor(math.Log(step) / math.Log(10))
	stepError := step / math.Pow(10, power)

	e := calculateE(stepError)
	if power >= 0 {
		return e * math.Pow(10, power)
	} else {
		return -math.Pow(10, -power) / e
	}
}

func NiceAndStep(start, stop, count float64) []float64 {
	var prestep float64
	for {
		step := tickIncrement(start, stop, count)
		if step == prestep || step == 0 || math.IsInf(step, 0) || math.IsNaN(step) {
			return []float64{start, stop, prestep}
		} else if step > 0 {
			start = math.Floor(start/step) * step
			stop = math.Ceil(stop/step) * step
		} else if step < 0 {
			start = math.Ceil(start*step) / step
			stop = math.Floor(stop*step) / step
		}
		prestep = step
	}
}
