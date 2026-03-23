package indicators

import "math"

// Round rounds a float64 to the specified number of decimal places.
func Round(v float64, decimals int) float64 {
	p := math.Pow(10, float64(decimals))
	return math.Round(v*p) / p
}

// SMA calculates the Simple Moving Average of the last `period` values in prices.
// Returns 0 if there are not enough values.
func SMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}
	sum := 0.0
	for _, v := range prices[len(prices)-period:] {
		sum += v
	}
	return sum / float64(period)
}

// StdDev calculates the population standard deviation of the last `period` values.
func StdDev(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}
	mean := SMA(prices, period)
	recent := prices[len(prices)-period:]
	variance := 0.0
	for _, v := range recent {
		d := v - mean
		variance += d * d
	}
	return math.Sqrt(variance / float64(period))
}

// SliceMin returns the minimum value in a slice, or 0 if empty.
func SliceMin(s []float64) float64 {
	if len(s) == 0 {
		return 0
	}
	m := s[0]
	for _, v := range s[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

// SliceMax returns the maximum value in a slice, or 0 if empty.
func SliceMax(s []float64) float64 {
	if len(s) == 0 {
		return 0
	}
	m := s[0]
	for _, v := range s[1:] {
		if v > m {
			m = v
		}
	}
	return m
}
