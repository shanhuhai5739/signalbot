package indicators

// MACDResult holds MACD line, signal line, histogram and trend classification.
type MACDResult struct {
	MACDLine   float64 `json:"macd_line"`
	SignalLine float64 `json:"signal_line"`
	Histogram  float64 `json:"histogram"`
	Trend      string  `json:"trend"`  // bullish | bearish | neutral
	Cross      string  `json:"cross"`  // golden_cross | death_cross | "" (empty = no cross)
}

// CalcMACD computes the MACD indicator.
// Standard parameters: fast=12, slow=26, signal=9.
// Returns a zero-value result if there is insufficient price data.
func CalcMACD(prices []float64, fast, slow, signal int) MACDResult {
	if fast <= 0 {
		fast = 12
	}
	if slow <= 0 {
		slow = 26
	}
	if signal <= 0 {
		signal = 9
	}
	if len(prices) < slow+signal {
		return MACDResult{Trend: "neutral"}
	}

	fastSeries := EMA(prices, fast)
	slowSeries := EMA(prices, slow)

	// MACD line is only valid from index (slow-1) onward
	macdLine := make([]float64, len(prices)-slow+1)
	for i := slow - 1; i < len(prices); i++ {
		macdLine[i-(slow-1)] = fastSeries[i] - slowSeries[i]
	}

	signalSeries := EMA(macdLine, signal)
	if len(signalSeries) == 0 {
		return MACDResult{Trend: "neutral"}
	}

	lastMACD := macdLine[len(macdLine)-1]
	lastSignal := signalSeries[len(signalSeries)-1]
	histogram := lastMACD - lastSignal

	// Detect bullish/bearish crossover by comparing the last two histograms
	cross := ""
	if len(signalSeries) >= 2 {
		prevMACD := macdLine[len(macdLine)-2]
		prevSignal := signalSeries[len(signalSeries)-2]
		prevHistogram := prevMACD - prevSignal
		if prevHistogram < 0 && histogram > 0 {
			cross = "golden_cross"
		} else if prevHistogram > 0 && histogram < 0 {
			cross = "death_cross"
		}
	}

	trend := "neutral"
	switch {
	case histogram > 0:
		trend = "bullish"
	case histogram < 0:
		trend = "bearish"
	}

	return MACDResult{
		MACDLine:   Round(lastMACD, 4),
		SignalLine: Round(lastSignal, 4),
		Histogram:  Round(histogram, 4),
		Trend:      trend,
		Cross:      cross,
	}
}
