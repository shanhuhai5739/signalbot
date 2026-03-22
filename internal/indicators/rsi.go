package indicators

import "math"

// RSIResult holds the RSI value and a descriptive signal.
type RSIResult struct {
	Value  float64 `json:"value"`
	Signal string  `json:"signal"` // overbought | bullish | neutral | bearish | oversold
}

// CalcRSI computes RSI using Wilder's smoothing method (the standard).
// period is typically 14. Returns a neutral result if data is insufficient.
func CalcRSI(prices []float64, period int) RSIResult {
	if period <= 0 {
		period = 14
	}
	if len(prices) < period+1 {
		return RSIResult{Value: 50, Signal: "neutral"}
	}

	// Initial average gain/loss over the first `period` changes
	var avgGain, avgLoss float64
	for i := 1; i <= period; i++ {
		diff := prices[i] - prices[i-1]
		if diff > 0 {
			avgGain += diff
		} else {
			avgLoss += math.Abs(diff)
		}
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// Wilder's smoothing for the remaining prices
	for i := period + 1; i < len(prices); i++ {
		diff := prices[i] - prices[i-1]
		if diff > 0 {
			avgGain = (avgGain*float64(period-1) + diff) / float64(period)
			avgLoss = (avgLoss * float64(period-1)) / float64(period)
		} else {
			avgGain = (avgGain * float64(period-1)) / float64(period)
			avgLoss = (avgLoss*float64(period-1) + math.Abs(diff)) / float64(period)
		}
	}

	var rsi float64
	if avgLoss == 0 {
		rsi = 100
	} else {
		rs := avgGain / avgLoss
		rsi = 100 - (100 / (1 + rs))
	}

	value := Round(rsi, 2)
	return RSIResult{
		Value:  value,
		Signal: rsiSignal(value),
	}
}

func rsiSignal(rsi float64) string {
	switch {
	case rsi >= 70:
		return "overbought"
	case rsi >= 55:
		return "bullish"
	case rsi <= 30:
		return "oversold"
	case rsi <= 45:
		return "bearish"
	default:
		return "neutral"
	}
}
