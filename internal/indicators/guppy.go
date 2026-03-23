package indicators

// GuppyEMA is a single EMA value tagged with its period.
type GuppyEMA struct {
	Period int     `json:"period"`
	Value  float64 `json:"value"`
}

// GuppyResult holds GMMA values and signals for a given set of periods.
type GuppyResult struct {
	ShortEMAs []GuppyEMA `json:"short_emas"` // fast group (traders)
	LongEMAs  []GuppyEMA `json:"long_emas"`  // slow group (investors)

	ShortMin float64 `json:"short_min"`
	ShortMax float64 `json:"short_max"`
	LongMin  float64 `json:"long_min"`
	LongMax  float64 `json:"long_max"`

	// GapPct: (short_min - long_max) / price × 100
	// Positive = short above long; negative = short below long
	GapPct float64 `json:"gap_pct"`

	// Alignment: "above_long" | "crossing" | "below_long"
	Alignment string `json:"alignment"`
	// Signal: "bullish" | "compression" | "bearish"
	Signal string `json:"signal"`
}

// GuppyHistoryEntry is one historical GMMA snapshot.
// BarIndex 0 is the most recent candle, 1 is one bar ago, etc.
type GuppyHistoryEntry struct {
	BarIndex  int        `json:"bar_index"`
	Close     float64    `json:"close"`
	ShortEMAs []GuppyEMA `json:"short_emas"`
	LongEMAs  []GuppyEMA `json:"long_emas"`
	ShortMin  float64    `json:"short_min"`
	ShortMax  float64    `json:"short_max"`
	LongMin   float64    `json:"long_min"`
	LongMax   float64    `json:"long_max"`
	GapPct    float64    `json:"gap_pct"`
	Alignment string     `json:"alignment"`
	Signal    string     `json:"signal"`
}

// DefaultShortPeriods is the classic Guppy short-term (fast) group.
var DefaultShortPeriods = []int{3, 5, 8, 10, 13, 21}

// DefaultLongPeriods is the classic Guppy long-term (slow) group.
var DefaultLongPeriods = []int{34, 55, 89, 144, 233, 377}

// CalcGuppy computes GMMA with the classic Guppy periods.
func CalcGuppy(prices []float64) GuppyResult {
	return CalcGuppyWithPeriods(prices, DefaultShortPeriods, DefaultLongPeriods)
}

// CalcGuppyWithPeriods computes GMMA with fully custom short and long EMA periods.
func CalcGuppyWithPeriods(prices []float64, shortPeriods, longPeriods []int) GuppyResult {
	shortEMAs := calcEMAGroup(prices, shortPeriods)
	longEMAs := calcEMAGroup(prices, longPeriods)

	shortVals := emaVals(shortEMAs)
	longVals := nonZero(emaVals(longEMAs))

	r := GuppyResult{
		ShortEMAs: shortEMAs,
		LongEMAs:  longEMAs,
	}

	if len(shortVals) == 0 {
		r.Alignment = "crossing"
		r.Signal = "compression"
		return r
	}

	r.ShortMin = SliceMin(shortVals)
	r.ShortMax = SliceMax(shortVals)
	if len(longVals) > 0 {
		r.LongMin = SliceMin(longVals)
		r.LongMax = SliceMax(longVals)
	}

	current := prices[len(prices)-1]
	if current > 0 && r.LongMax > 0 {
		r.GapPct = Round((r.ShortMin-r.LongMax)/current*100, 2)
	}

	switch {
	case r.LongMax > 0 && r.ShortMin > r.LongMax:
		r.Alignment = "above_long"
		r.Signal = "bullish"
	case r.LongMin > 0 && r.ShortMax < r.LongMin:
		r.Alignment = "below_long"
		r.Signal = "bearish"
	default:
		r.Alignment = "crossing"
		r.Signal = "compression"
	}

	return r
}

// CalcGuppyHistory returns GMMA snapshots for the last n closing bars.
// The slice at index 0 represents the most recent candle; index 1 is one bar
// ago, and so on. This lets callers detect trend direction and gap expansion
// by comparing adjacent entries.
func CalcGuppyHistory(closes []float64, n int, shortPeriods, longPeriods []int) []GuppyHistoryEntry {
	if n <= 0 {
		n = 1
	}
	if n > len(closes) {
		n = len(closes)
	}
	entries := make([]GuppyHistoryEntry, n)
	for i := 0; i < n; i++ {
		end := len(closes) - i
		if end < 1 {
			break
		}
		g := CalcGuppyWithPeriods(closes[:end], shortPeriods, longPeriods)
		entries[i] = GuppyHistoryEntry{
			BarIndex:  i,
			Close:     Round(closes[end-1], 4),
			ShortEMAs: g.ShortEMAs,
			LongEMAs:  g.LongEMAs,
			ShortMin:  g.ShortMin,
			ShortMax:  g.ShortMax,
			LongMin:   g.LongMin,
			LongMax:   g.LongMax,
			GapPct:    g.GapPct,
			Alignment: g.Alignment,
			Signal:    g.Signal,
		}
	}
	return entries
}

// nonZero filters out zero values (used to skip EMA periods with insufficient data).
func nonZero(s []float64) []float64 {
	out := make([]float64, 0, len(s))
	for _, v := range s {
		if v != 0 {
			out = append(out, v)
		}
	}
	return out
}

// calcEMAGroup computes the last EMA value for each period in the group.
func calcEMAGroup(prices []float64, periods []int) []GuppyEMA {
	result := make([]GuppyEMA, len(periods))
	for i, p := range periods {
		result[i] = GuppyEMA{
			Period: p,
			Value:  Round(LastEMA(prices, p), 4),
		}
	}
	return result
}

// emaVals extracts the float values from a GuppyEMA slice.
func emaVals(emas []GuppyEMA) []float64 {
	vals := make([]float64, len(emas))
	for i, e := range emas {
		vals[i] = e.Value
	}
	return vals
}
