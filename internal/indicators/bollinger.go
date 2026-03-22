package indicators

// BollingerResult holds Bollinger Band values and price position classification.
type BollingerResult struct {
	Upper    float64 `json:"upper"`
	Middle   float64 `json:"middle"` // SMA(20)
	Lower    float64 `json:"lower"`
	Width    float64 `json:"width"`     // upper - lower (absolute bandwidth)
	PercentB float64 `json:"percent_b"` // (price - lower) / width, 0–1 range
	Position string  `json:"position"`  // above_upper | upper_zone | middle | lower_zone | below_lower
	Signal   string  `json:"signal"`    // bullish | neutral | bearish | overbought | oversold
}

// CalcBollinger computes Bollinger Bands with the given period and standard deviation multiplier.
// Standard parameters: period=20, multiplier=2.0.
func CalcBollinger(prices []float64, period int, multiplier float64) BollingerResult {
	if period <= 0 {
		period = 20
	}
	if multiplier <= 0 {
		multiplier = 2.0
	}
	if len(prices) < period {
		return BollingerResult{}
	}

	middle := SMA(prices, period)
	stdDev := StdDev(prices, period)

	upper := middle + multiplier*stdDev
	lower := middle - multiplier*stdDev
	width := upper - lower
	current := prices[len(prices)-1]

	var percentB float64
	if width > 0 {
		percentB = (current - lower) / width
	}

	position, signal := classifyBBPosition(current, upper, lower, percentB)

	return BollingerResult{
		Upper:    Round(upper, 4),
		Middle:   Round(middle, 4),
		Lower:    Round(lower, 4),
		Width:    Round(width, 4),
		PercentB: Round(percentB, 4),
		Position: position,
		Signal:   signal,
	}
}

// classifyBBPosition categorises price relative to the Bollinger Bands.
// %B > 1.0 means price is above the upper band; < 0 means below lower.
func classifyBBPosition(price, upper, lower, percentB float64) (position, signal string) {
	switch {
	case price > upper:
		return "above_upper", "overbought"
	case percentB >= 0.8:
		return "upper_zone", "bullish"
	case percentB <= 0.2 && price > lower:
		return "lower_zone", "bearish"
	case price < lower:
		return "below_lower", "oversold"
	default:
		return "middle", "neutral"
	}
}
