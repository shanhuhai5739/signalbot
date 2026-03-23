package indicators

// GuppyResult holds the Guppy Multiple Moving Average (GMMA) values and signals.
//
// GMMA uses two EMA groups:
//   - Short-term (fast money / traders): 3, 5, 8, 10, 13, 21
//   - Long-term  (patient money / investors): 34, 55, 89, 144, 233, 377
//
// When the short-term group is entirely above the long-term group and both are
// expanding, it signals a strong uptrend. Compression (groups overlapping) signals
// a potential trend change or consolidation.
type GuppyResult struct {
	// Short-term EMA group (fast money / traders)
	EMA3  float64 `json:"ema3"`
	EMA5  float64 `json:"ema5"`
	EMA8  float64 `json:"ema8"`
	EMA10 float64 `json:"ema10"`
	EMA13 float64 `json:"ema13"`
	EMA21 float64 `json:"ema21"`

	// Long-term EMA group (patient money / investors)
	EMA34  float64 `json:"ema34"`
	EMA55  float64 `json:"ema55"`
	EMA89  float64 `json:"ema89"`
	EMA144 float64 `json:"ema144"`
	// EMA233 and EMA377 require 233+/377+ candles; 0 means insufficient data.
	EMA233 float64 `json:"ema233"`
	EMA377 float64 `json:"ema377"`

	ShortMin float64 `json:"short_min"` // lowest value in short-term group
	ShortMax float64 `json:"short_max"` // highest value in short-term group
	LongMin  float64 `json:"long_min"`  // lowest value in long-term group (non-zero EMAs)
	LongMax  float64 `json:"long_max"`  // highest value in long-term group (non-zero EMAs)

	// GapPct is the percentage gap between the two groups relative to current price.
	// Positive = short above long; negative = short below long; near 0 = compression.
	GapPct float64 `json:"gap_pct"`

	// Alignment describes the spatial relationship between the two groups.
	// "above_long"  – short-term group entirely above long-term (bullish expansion)
	// "below_long"  – short-term group entirely below long-term (bearish expansion)
	// "crossing"    – groups overlap (compression / transition)
	Alignment string `json:"alignment"`
	Signal    string `json:"signal"` // bullish | compression | bearish
}

// CalcGuppy computes the GMMA for the given closing prices.
// At least 89 data points are needed for the core long-term group;
// EMA233/EMA377 return 0 when data is insufficient.
func CalcGuppy(prices []float64) GuppyResult {
	r := GuppyResult{
		EMA3:   Round(LastEMA(prices, 3), 4),
		EMA5:   Round(LastEMA(prices, 5), 4),
		EMA8:   Round(LastEMA(prices, 8), 4),
		EMA10:  Round(LastEMA(prices, 10), 4),
		EMA13:  Round(LastEMA(prices, 13), 4),
		EMA21:  Round(LastEMA(prices, 21), 4),
		EMA34:  Round(LastEMA(prices, 34), 4),
		EMA55:  Round(LastEMA(prices, 55), 4),
		EMA89:  Round(LastEMA(prices, 89), 4),
		EMA144: Round(LastEMA(prices, 144), 4),
		EMA233: Round(LastEMA(prices, 233), 4),
		EMA377: Round(LastEMA(prices, 377), 4),
	}

	// Short-term group bounds
	shorts := []float64{r.EMA3, r.EMA5, r.EMA8, r.EMA10, r.EMA13, r.EMA21}
	r.ShortMin = SliceMin(shorts)
	r.ShortMax = SliceMax(shorts)

	// Long-term group: only include non-zero values (zero = insufficient data)
	longs := nonZero([]float64{r.EMA34, r.EMA55, r.EMA89, r.EMA144, r.EMA233, r.EMA377})
	if len(longs) == 0 {
		r.Alignment = "crossing"
		r.Signal = "compression"
		return r
	}
	r.LongMin = SliceMin(longs)
	r.LongMax = SliceMax(longs)

	// Gap between groups as % of current price
	current := prices[len(prices)-1]
	if current > 0 {
		gap := r.ShortMin - r.LongMax
		if gap < 0 {
			// short is below long; flip so gap describes the distance (negative means bearish)
			gap = -(r.LongMin - r.ShortMax)
			if r.LongMin > r.ShortMax {
				gap = r.LongMin - r.ShortMax
				gap = -gap
			}
		}
		r.GapPct = Round((r.ShortMin-r.LongMax)/current*100, 2)
	}

	// Alignment
	switch {
	case r.ShortMin > r.LongMax:
		r.Alignment = "above_long"
		r.Signal = "bullish"
	case r.ShortMax < r.LongMin:
		r.Alignment = "below_long"
		r.Signal = "bearish"
	default:
		r.Alignment = "crossing"
		r.Signal = "compression"
	}

	return r
}

// nonZero filters out zero values from a slice.
func nonZero(s []float64) []float64 {
	var out []float64
	for _, v := range s {
		if v != 0 {
			out = append(out, v)
		}
	}
	return out
}
