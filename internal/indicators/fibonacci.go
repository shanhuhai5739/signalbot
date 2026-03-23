package indicators

import "math"

// FibLevel represents a single Fibonacci retracement level.
type FibLevel struct {
	Label    string  `json:"label"`    // e.g. "61.8%"
	Ratio    float64 `json:"ratio"`    // 0.618
	Price    float64 `json:"price"`    // actual price at this level
	IsAbove  bool    `json:"is_above"` // true when current price is at or above this level
}

// FibonacciResult holds the retracement analysis between the significant
// swing high and swing low within the lookback window.
type FibonacciResult struct {
	SwingHigh    float64    `json:"swing_high"`    // highest high in lookback
	SwingLow     float64    `json:"swing_low"`     // lowest low in lookback
	Range        float64    `json:"range"`         // SwingHigh - SwingLow
	Levels       []FibLevel `json:"levels"`        // 0% → 100% levels
	NearestLevel string     `json:"nearest_level"` // e.g. "61.8%"
	DistancePct  float64    `json:"distance_pct"`  // % distance from nearest level
	// Signal: "at_support" | "at_resistance" | "between_levels"
	Signal    string `json:"signal"`
	// Direction: "upper_half" (price > midpoint, still bullish territory)
	//            "lower_half" (price < midpoint, bearish territory)
	Direction string `json:"direction"`
}

// fibRatios defines the standard Fibonacci retracement levels.
var fibRatios = []struct {
	label string
	ratio float64
}{
	{"0.0%", 0.0},
	{"23.6%", 0.236},
	{"38.2%", 0.382},
	{"50.0%", 0.500},
	{"61.8%", 0.618},
	{"78.6%", 0.786},
	{"100.0%", 1.0},
}

// CalcFibonacci identifies the swing high/low within the last `lookback` candles
// and computes retracement levels. Levels are measured from the swing high downward,
// which matches a standard bearish-retracement setup; use the IsAbove field to
// determine whether each level acts as support or resistance for the current price.
func CalcFibonacci(highs, lows, closes []float64, lookback int) FibonacciResult {
	n := len(closes)
	if n == 0 {
		return FibonacciResult{}
	}
	if lookback <= 0 || lookback > n {
		lookback = n
	}
	start := n - lookback

	// Find swing high and low within the lookback window
	swingHigh := highs[start]
	swingLow := lows[start]
	for i := start + 1; i < n; i++ {
		if highs[i] > swingHigh {
			swingHigh = highs[i]
		}
		if lows[i] < swingLow {
			swingLow = lows[i]
		}
	}

	priceRange := swingHigh - swingLow
	if priceRange == 0 {
		return FibonacciResult{
			SwingHigh: Round(swingHigh, 4),
			SwingLow:  Round(swingLow, 4),
		}
	}

	current := closes[n-1]

	// Build level slice (measured top-down from swing high)
	levels := make([]FibLevel, len(fibRatios))
	for i, fr := range fibRatios {
		price := swingHigh - fr.ratio*priceRange
		levels[i] = FibLevel{
			Label:   fr.label,
			Ratio:   fr.ratio,
			Price:   Round(price, 4),
			IsAbove: current >= price,
		}
	}

	// Nearest level by absolute price distance
	nearestIdx := 0
	minDist := math.Abs(current - levels[0].Price)
	for i, l := range levels {
		d := math.Abs(current - l.Price)
		if d < minDist {
			minDist = d
			nearestIdx = i
		}
	}

	distancePct := Round((minDist/current)*100, 2)

	signal := "between_levels"
	if distancePct < 1.5 {
		// Within 1.5% of a fib level — treat as "at" that level
		if current >= levels[nearestIdx].Price {
			signal = "at_resistance"
		} else {
			signal = "at_support"
		}
	}

	direction := "lower_half"
	if current > (swingHigh+swingLow)/2 {
		direction = "upper_half"
	}

	return FibonacciResult{
		SwingHigh:    Round(swingHigh, 4),
		SwingLow:     Round(swingLow, 4),
		Range:        Round(priceRange, 4),
		Levels:       levels,
		NearestLevel: levels[nearestIdx].Label,
		DistancePct:  distancePct,
		Signal:       signal,
		Direction:    direction,
	}
}
