package indicators

import "math"

// VWAPResult holds the Anchored VWAP and its standard-deviation bands.
//
// The VWAP is anchored to the first candle of the provided slice. The engine
// controls the anchor window (default: last 50 candles) so the result reflects
// a meaningful short-to-medium-term volume-weighted average price.
//
// Standard deviation bands:
//   Band1 = VWAP ± 1σ  (normal fluctuation zone)
//   Band2 = VWAP ± 2σ  (extended / mean-reversion zone)
type VWAPResult struct {
	Value      float64 `json:"value"`
	UpperBand1 float64 `json:"upper_band1"` // VWAP + 1σ
	LowerBand1 float64 `json:"lower_band1"` // VWAP − 1σ
	UpperBand2 float64 `json:"upper_band2"` // VWAP + 2σ
	LowerBand2 float64 `json:"lower_band2"` // VWAP − 2σ
	StdDev     float64 `json:"std_dev"`
	// DeviationPct: how far the current price deviates from VWAP (%)
	DeviationPct float64 `json:"deviation_pct"`
	// Position: "above_band2" | "above_band1" | "above_vwap"
	//           "below_vwap"  | "below_band1" | "below_band2"
	Position string `json:"position"`
	// Signal: "overbought" | "bullish" | "neutral" | "bearish" | "oversold"
	Signal string `json:"signal"`
}

// CalcVWAP computes the anchored VWAP from the first candle of the provided slices.
// Typical price = (High + Low + Close) / 3.
// Volume-weighted variance is used to derive the standard deviation bands.
func CalcVWAP(highs, lows, closes, volumes []float64) VWAPResult {
	n := len(closes)
	if n < 2 {
		return VWAPResult{}
	}

	// Accumulate cumulative TPV (Typical Price × Volume) and volume
	cumTPV := 0.0
	cumVol := 0.0
	for i := 0; i < n; i++ {
		tp := (highs[i] + lows[i] + closes[i]) / 3.0
		cumTPV += tp * volumes[i]
		cumVol += volumes[i]
	}
	if cumVol == 0 {
		return VWAPResult{}
	}

	vwap := cumTPV / cumVol

	// Volume-weighted variance for standard deviation bands
	cumVar := 0.0
	for i := 0; i < n; i++ {
		tp := (highs[i] + lows[i] + closes[i]) / 3.0
		d := tp - vwap
		cumVar += volumes[i] * d * d
	}
	sd := math.Sqrt(cumVar / cumVol)

	upper1 := vwap + sd
	lower1 := vwap - sd
	upper2 := vwap + 2*sd
	lower2 := vwap - 2*sd

	current := closes[n-1]
	devPct := 0.0
	if vwap != 0 {
		devPct = (current - vwap) / vwap * 100
	}

	pos, sig := vwapClassify(current, vwap, upper1, lower1, upper2, lower2)

	return VWAPResult{
		Value:        Round(vwap, 4),
		UpperBand1:   Round(upper1, 4),
		LowerBand1:   Round(lower1, 4),
		UpperBand2:   Round(upper2, 4),
		LowerBand2:   Round(lower2, 4),
		StdDev:       Round(sd, 4),
		DeviationPct: Round(devPct, 2),
		Position:     pos,
		Signal:       sig,
	}
}

func vwapClassify(price, vwap, upper1, lower1, upper2, lower2 float64) (position, signal string) {
	switch {
	case price > upper2:
		return "above_band2", "overbought"
	case price > upper1:
		return "above_band1", "bullish"
	case price >= vwap:
		return "above_vwap", "bullish"
	case price < lower2:
		return "below_band2", "oversold"
	case price < lower1:
		return "below_band1", "bearish"
	default:
		return "below_vwap", "bearish"
	}
}
