package indicators

// VPVRBin represents one price-level bucket in the volume profile.
type VPVRBin struct {
	PriceLow    float64 `json:"price_low"`
	PriceHigh   float64 `json:"price_high"`
	PriceMid    float64 `json:"price_mid"`
	Volume      float64 `json:"volume"`
	IsPOC       bool    `json:"is_poc"`       // Point of Control (highest-volume bin)
	IsValueArea bool    `json:"is_value_area"` // inside the 70% Value Area
}

// VPVRResult holds the Fixed Range Volume Profile analysis.
//
// Key concepts:
//   POC (Point of Control) – the price level with the most trading activity.
//   VAH (Value Area High)  – upper boundary of the 70% value area.
//   VAL (Value Area Low)   – lower boundary of the 70% value area.
//
// Price at or near the POC often acts as strong support/resistance.
// Breaking above VAH or below VAL with volume signals trend continuation.
type VPVRResult struct {
	POC    float64   `json:"poc"` // Point of Control price (mid of highest-volume bin)
	VAH    float64   `json:"vah"` // Value Area High
	VAL    float64   `json:"val"` // Value Area Low
	NumBins int      `json:"num_bins"`
	Bins   []VPVRBin `json:"bins"`
	// Signal: "above_vah" | "above_poc" | "at_poc" | "below_poc" | "below_val"
	Signal string `json:"signal"`
}

// CalcVPVR computes the Fixed Range Volume Profile over the provided candle data.
// Volume for each candle is distributed proportionally across the price bins it spans.
// numBins defaults to 24 when ≤ 0.
func CalcVPVR(highs, lows, closes, volumes []float64, numBins int) VPVRResult {
	if numBins <= 0 {
		numBins = 24
	}
	n := len(closes)
	if n < 2 {
		return VPVRResult{}
	}

	// Determine price range
	priceHigh := highs[0]
	priceLow := lows[0]
	for i := 1; i < n; i++ {
		if highs[i] > priceHigh {
			priceHigh = highs[i]
		}
		if lows[i] < priceLow {
			priceLow = lows[i]
		}
	}
	if priceHigh <= priceLow {
		return VPVRResult{}
	}

	binWidth := (priceHigh - priceLow) / float64(numBins)
	binVolumes := make([]float64, numBins)

	// Distribute each candle's volume across the bins it spans, proportionally.
	for i := 0; i < n; i++ {
		lo := lows[i]
		hi := highs[i]
		vol := volumes[i]
		candleRange := hi - lo

		if candleRange == 0 {
			// Degenerate candle: all volume to its single bin
			idx := binIndex(lo, priceLow, binWidth, numBins)
			binVolumes[idx] += vol
			continue
		}

		for b := 0; b < numBins; b++ {
			bLow := priceLow + float64(b)*binWidth
			bHigh := bLow + binWidth
			overlap := overlapLen(lo, hi, bLow, bHigh)
			if overlap > 0 {
				binVolumes[b] += vol * (overlap / candleRange)
			}
		}
	}

	// Build bin slice and find POC
	bins := make([]VPVRBin, numBins)
	maxVol := 0.0
	pocIdx := 0
	totalVol := 0.0
	for b := 0; b < numBins; b++ {
		lo := priceLow + float64(b)*binWidth
		hi := lo + binWidth
		bins[b] = VPVRBin{
			PriceLow:  Round(lo, 4),
			PriceHigh: Round(hi, 4),
			PriceMid:  Round((lo+hi)/2, 4),
			Volume:    Round(binVolumes[b], 4),
		}
		totalVol += binVolumes[b]
		if binVolumes[b] > maxVol {
			maxVol = binVolumes[b]
			pocIdx = b
		}
	}
	bins[pocIdx].IsPOC = true

	// Value Area: expand outward from POC until 70% of total volume is covered.
	targetVol := totalVol * 0.70
	vaVol := binVolumes[pocIdx]
	loIdx, hiIdx := pocIdx, pocIdx

	for vaVol < targetVol {
		canExpandLo := loIdx > 0
		canExpandHi := hiIdx < numBins-1
		if !canExpandLo && !canExpandHi {
			break
		}
		// Expand toward the side that adds more volume
		if canExpandLo && canExpandHi {
			if binVolumes[loIdx-1] >= binVolumes[hiIdx+1] {
				loIdx--
				vaVol += binVolumes[loIdx]
			} else {
				hiIdx++
				vaVol += binVolumes[hiIdx]
			}
		} else if canExpandLo {
			loIdx--
			vaVol += binVolumes[loIdx]
		} else {
			hiIdx++
			vaVol += binVolumes[hiIdx]
		}
	}

	for b := loIdx; b <= hiIdx; b++ {
		bins[b].IsValueArea = true
	}

	poc := bins[pocIdx].PriceMid
	vah := bins[hiIdx].PriceHigh
	val := bins[loIdx].PriceLow

	current := closes[n-1]

	return VPVRResult{
		POC:     Round(poc, 4),
		VAH:     Round(vah, 4),
		VAL:     Round(val, 4),
		NumBins: numBins,
		Bins:    bins,
		Signal:  vpvrSignal(current, poc, vah, val),
	}
}

// binIndex returns the bin index for a given price.
func binIndex(price, priceLow, binWidth float64, numBins int) int {
	idx := int((price - priceLow) / binWidth)
	if idx < 0 {
		idx = 0
	}
	if idx >= numBins {
		idx = numBins - 1
	}
	return idx
}

// overlapLen returns the length of the overlap between [a1,a2] and [b1,b2].
func overlapLen(a1, a2, b1, b2 float64) float64 {
	lo := a1
	if b1 > lo {
		lo = b1
	}
	hi := a2
	if b2 < hi {
		hi = b2
	}
	if hi <= lo {
		return 0
	}
	return hi - lo
}

func vpvrSignal(price, poc, vah, val float64) string {
	switch {
	case price > vah:
		return "above_vah"
	case price > poc:
		return "above_poc"
	case price >= val && price <= poc:
		atPocThreshold := poc * 0.005 // within 0.5% counts as "at POC"
		if price >= poc-atPocThreshold && price <= poc+atPocThreshold {
			return "at_poc"
		}
		return "below_poc"
	default:
		return "below_val"
	}
}
