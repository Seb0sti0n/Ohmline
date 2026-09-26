package engine

import (
	"math"
	"sort"
)

func mean(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	s := 0.0
	for _, v := range x {
		s += v
	}
	return s / float64(len(x))
}

func std(x []float64) float64 {
	if len(x) < 2 {
		return 0
	}
	m := mean(x)
	s := 0.0
	for _, v := range x {
		s += (v - m) * (v - m)
	}
	return math.Sqrt(s / float64(len(x)-1))
}

func median(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	c := append([]float64(nil), x...)
	sort.Float64s(c)
	n := len(c)
	if n%2 == 1 {
		return c[n/2]
	}
	return (c[n/2-1] + c[n/2]) / 2
}

// scaledMAD is the median absolute deviation scaled to be comparable to a standard deviation.
func scaledMAD(x []float64) float64 {
	m := median(x)
	dev := make([]float64, len(x))
	for i, v := range x {
		dev[i] = math.Abs(v - m)
	}
	return 1.4826 * median(dev)
}

func round(v float64, decimals int) float64 {
	p := math.Pow(10, float64(decimals))
	return math.Round(v*p) / p
}

func clamp01(v float64) float64 { return math.Max(0, math.Min(1, v)) }
