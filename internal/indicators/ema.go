package indicators

// EMAResult holds EMA values for four key periods and a trend alignment signal.
type EMAResult struct {
	EMA9      float64 `json:"ema9"`
	EMA21     float64 `json:"ema21"`
	EMA50     float64 `json:"ema50"`
	EMA200    float64 `json:"ema200"`
	Alignment string  `json:"alignment"` // strongly_bullish | bullish | neutral | bearish | strongly_bearish
	Signal    string  `json:"signal"`    // bullish | neutral | bearish
}

// EMA computes the full Exponential Moving Average series for the given period.
// The first (period-1) values are zero (insufficient data).
// Uses standard EMA seeding: SMA of the first `period` prices as the initial value.
func EMA(prices []float64, period int) []float64 {
	result := make([]float64, len(prices))
	if len(prices) < period {
		return result
	}

	k := 2.0 / float64(period+1)

	// Seed with SMA of the first `period` candles
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	result[period-1] = sum / float64(period)

	for i := period; i < len(prices); i++ {
		result[i] = prices[i]*k + result[i-1]*(1-k)
	}
	return result
}

// LastEMA returns only the most recent EMA value for the given period.
func LastEMA(prices []float64, period int) float64 {
	series := EMA(prices, period)
	if len(series) == 0 {
		return 0
	}
	return series[len(series)-1]
}

// CalcEMA computes EMA9, EMA21, EMA50, EMA200 and determines trend alignment.
func CalcEMA(prices []float64) EMAResult {
	r := EMAResult{
		EMA9:   Round(LastEMA(prices, 9), 4),
		EMA21:  Round(LastEMA(prices, 21), 4),
		EMA50:  Round(LastEMA(prices, 50), 4),
		EMA200: Round(LastEMA(prices, 200), 4),
	}

	switch {
	case r.EMA9 > r.EMA21 && r.EMA21 > r.EMA50 && r.EMA50 > r.EMA200:
		r.Alignment = "strongly_bullish"
		r.Signal = "bullish"
	case r.EMA9 > r.EMA21 && r.EMA21 > r.EMA50:
		r.Alignment = "bullish"
		r.Signal = "bullish"
	case r.EMA9 < r.EMA21 && r.EMA21 < r.EMA50 && r.EMA50 < r.EMA200:
		r.Alignment = "strongly_bearish"
		r.Signal = "bearish"
	case r.EMA9 < r.EMA21 && r.EMA21 < r.EMA50:
		r.Alignment = "bearish"
		r.Signal = "bearish"
	default:
		r.Alignment = "neutral"
		r.Signal = "neutral"
	}

	return r
}
