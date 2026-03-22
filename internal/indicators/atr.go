package indicators

import "math"

// ATRResult holds the ATR value, its percentage of current price, and a volatility regime label.
type ATRResult struct {
	Value  float64 `json:"value"`
	Pct    float64 `json:"pct"`    // ATR as % of current close price
	Regime string  `json:"regime"` // low_volatility | normal | high_volatility
}

// CalcATR computes the Average True Range using Wilder's smoothing method.
// Standard period: 14.
// Inputs are separate price slices extracted from []Candle to keep this package
// free of cross-package dependencies.
func CalcATR(highs, lows, closes []float64, period int) ATRResult {
	if period <= 0 {
		period = 14
	}
	n := len(closes)
	if n < period+1 || len(highs) != n || len(lows) != n {
		return ATRResult{Regime: "normal"}
	}

	// True Range for each bar (starting at index 1)
	trs := make([]float64, n-1)
	for i := 1; i < n; i++ {
		hl := highs[i] - lows[i]
		hpc := math.Abs(highs[i] - closes[i-1])
		lpc := math.Abs(lows[i] - closes[i-1])
		trs[i-1] = math.Max(hl, math.Max(hpc, lpc))
	}

	if len(trs) < period {
		return ATRResult{Regime: "normal"}
	}

	// Seed ATR with simple average of the first `period` true ranges
	atr := 0.0
	for i := 0; i < period; i++ {
		atr += trs[i]
	}
	atr /= float64(period)

	// Wilder's smoothing for remaining bars
	for i := period; i < len(trs); i++ {
		atr = (atr*float64(period-1) + trs[i]) / float64(period)
	}

	currentPrice := closes[n-1]
	pct := 0.0
	if currentPrice > 0 {
		pct = (atr / currentPrice) * 100
	}

	return ATRResult{
		Value:  Round(atr, 4),
		Pct:    Round(pct, 2),
		Regime: atrRegime(pct),
	}
}

// atrRegime classifies volatility based on ATR as a percentage of price.
// Thresholds are calibrated for crypto (BTC) and gold — adjust if needed.
func atrRegime(pct float64) string {
	switch {
	case pct < 1.0:
		return "low_volatility"
	case pct > 3.5:
		return "high_volatility"
	default:
		return "normal"
	}
}
