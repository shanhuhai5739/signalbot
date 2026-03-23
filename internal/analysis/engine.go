package analysis

import (
	"sort"

	"signalbot/internal/data"
	"signalbot/internal/indicators"
	"signalbot/internal/report"
)

// minCandles is the minimum number of candles required to compute all indicators reliably.
// EMA200 requires 200+ bars, plus buffer for Wilder smoothing convergence.
const minCandles = 210

// Analyze runs all technical indicators on the provided candle data and returns
// a fully populated Report ready for JSON serialisation.
func Analyze(asset, timeframe string, candles []data.Candle) *report.Report {
	closes := data.ExtractCloses(candles)
	highs := data.ExtractHighs(candles)
	lows := data.ExtractLows(candles)
	opens := data.ExtractOpens(candles)
	volumes := data.ExtractVolumes(candles)

	last := candles[len(candles)-1]
	current := closes[len(closes)-1]
	prevClose := closes[len(closes)-2]
	changePct := 0.0
	if prevClose != 0 {
		changePct = (current - prevClose) / prevClose * 100
	}

	// --- Compute all indicators ---
	emaResult := indicators.CalcEMA(closes)
	rsiResult := indicators.CalcRSI(closes, 14)
	macdResult := indicators.CalcMACD(closes, 12, 26, 9)
	bbResult := indicators.CalcBollinger(closes, 20, 2.0)
	atrResult := indicators.CalcATR(highs, lows, closes, 14)
	volResult := indicators.CalcVolume(volumes, closes)

	// --- New indicators ---
	guppyResult := indicators.CalcGuppy(closes)
	fibResult := indicators.CalcFibonacci(highs, lows, closes, 100)

	// Anchored VWAP: anchor to the last 50 candles (or all if fewer)
	vwapWindow := 50
	if len(closes) < vwapWindow {
		vwapWindow = len(closes)
	}
	vwapResult := indicators.CalcVWAP(
		highs[len(highs)-vwapWindow:],
		lows[len(lows)-vwapWindow:],
		closes[len(closes)-vwapWindow:],
		volumes[len(volumes)-vwapWindow:],
	)

	// VPVR: fixed range volume profile over the last 100 candles, 24 bins
	vpvrWindow := 100
	if len(closes) < vpvrWindow {
		vpvrWindow = len(closes)
	}
	vpvrResult := indicators.CalcVPVR(
		highs[len(highs)-vpvrWindow:],
		lows[len(lows)-vpvrWindow:],
		closes[len(closes)-vpvrWindow:],
		volumes[len(volumes)-vpvrWindow:],
		24,
	)

	// --- Score-based signal aggregation ---
	score := calcScore(emaResult, rsiResult, macdResult, bbResult)
	trend, strength, signal := classifySignal(score)
	confidence := calcConfidence(score)

	// --- Key levels ---
	supports, resistances := calcKeyLevels(closes, highs, lows, bbResult, emaResult)

	return &report.Report{
		Asset:     asset,
		Timeframe: timeframe,
		Timestamp: last.CloseTime,
		Price: report.PriceInfo{
			Current:   indicators.Round(current, 4),
			Open:      indicators.Round(opens[len(opens)-1], 4),
			High:      indicators.Round(highs[len(highs)-1], 4),
			Low:       indicators.Round(lows[len(lows)-1], 4),
			ChangePct: indicators.Round(changePct, 2),
		},
		Indicators: report.Indicators{
			RSI: report.RSIData{
				Value:  rsiResult.Value,
				Signal: rsiResult.Signal,
			},
			MACD: report.MACDData{
				MACDLine:   macdResult.MACDLine,
				SignalLine: macdResult.SignalLine,
				Histogram:  macdResult.Histogram,
				Trend:      macdResult.Trend,
				Cross:      macdResult.Cross,
			},
			Bollinger: report.BollingerData{
				Upper:    bbResult.Upper,
				Middle:   bbResult.Middle,
				Lower:    bbResult.Lower,
				Width:    bbResult.Width,
				PercentB: bbResult.PercentB,
				Position: bbResult.Position,
				Signal:   bbResult.Signal,
			},
			EMA: report.EMAData{
				EMA9:      emaResult.EMA9,
				EMA21:     emaResult.EMA21,
				EMA50:     emaResult.EMA50,
				EMA200:    emaResult.EMA200,
				Alignment: emaResult.Alignment,
				Signal:    emaResult.Signal,
			},
			ATR: report.ATRData{
				Value:  atrResult.Value,
				Pct:    atrResult.Pct,
				Regime: atrResult.Regime,
			},
			Volume: report.VolumeData{
				Current: volResult.Current,
				SMA20:   volResult.SMA20,
				Ratio:   volResult.Ratio,
				OBV:     volResult.OBV,
				Signal:  volResult.Signal,
			},
			Guppy: report.GuppyData{
				EMA3:      guppyResult.EMA3,
				EMA5:      guppyResult.EMA5,
				EMA8:      guppyResult.EMA8,
				EMA10:     guppyResult.EMA10,
				EMA13:     guppyResult.EMA13,
				EMA21:     guppyResult.EMA21,
				EMA34:     guppyResult.EMA34,
				EMA55:     guppyResult.EMA55,
				EMA89:     guppyResult.EMA89,
				EMA144:    guppyResult.EMA144,
				EMA233:    guppyResult.EMA233,
				EMA377:    guppyResult.EMA377,
				ShortMin:  guppyResult.ShortMin,
				ShortMax:  guppyResult.ShortMax,
				LongMin:   guppyResult.LongMin,
				LongMax:   guppyResult.LongMax,
				GapPct:    guppyResult.GapPct,
				Alignment: guppyResult.Alignment,
				Signal:    guppyResult.Signal,
			},
			Fibonacci: report.FibonacciData{
				SwingHigh:    fibResult.SwingHigh,
				SwingLow:     fibResult.SwingLow,
				Range:        fibResult.Range,
				Levels:       mapFibLevels(fibResult.Levels),
				NearestLevel: fibResult.NearestLevel,
				DistancePct:  fibResult.DistancePct,
				Signal:       fibResult.Signal,
				Direction:    fibResult.Direction,
			},
			VWAP: report.VWAPData{
				Value:        vwapResult.Value,
				UpperBand1:   vwapResult.UpperBand1,
				LowerBand1:   vwapResult.LowerBand1,
				UpperBand2:   vwapResult.UpperBand2,
				LowerBand2:   vwapResult.LowerBand2,
				StdDev:       vwapResult.StdDev,
				DeviationPct: vwapResult.DeviationPct,
				Position:     vwapResult.Position,
				Signal:       vwapResult.Signal,
			},
			VPVR: report.VPVRData{
				POC:     vpvrResult.POC,
				VAH:     vpvrResult.VAH,
				VAL:     vpvrResult.VAL,
				NumBins: vpvrResult.NumBins,
				Bins:    mapVPVRBins(vpvrResult.Bins),
				Signal:  vpvrResult.Signal,
			},
		},
		Analysis: report.Analysis{
			Trend:      trend,
			Strength:   strength,
			Signal:     signal,
			Confidence: confidence,
			Score:      score,
			KeyLevels: report.KeyLevels{
				Support:    supports,
				Resistance: resistances,
			},
		},
	}
}

// calcScore aggregates individual indicator signals into a single integer score.
//
// Scoring logic (range approximately -8 to +8):
//   EMA:  strongly_bullish=+2, bullish=+1, bearish=-1, strongly_bearish=-2
//   RSI:  bullish=+1, oversold=+1(bounce), overbought=-1(caution), bearish=-1
//   MACD: bullish=+1, bearish=-1; golden_cross=+2 bonus, death_cross=-2 bonus
//   BB:   bullish=+1, bearish=-1, overbought=-1, oversold=+1(bounce)
func calcScore(
	ema indicators.EMAResult,
	rsi indicators.RSIResult,
	macd indicators.MACDResult,
	bb indicators.BollingerResult,
) int {
	score := 0

	// EMA alignment
	switch ema.Alignment {
	case "strongly_bullish":
		score += 2
	case "bullish":
		score += 1
	case "bearish":
		score -= 1
	case "strongly_bearish":
		score -= 2
	}

	// RSI
	switch rsi.Signal {
	case "bullish":
		score += 1
	case "oversold":
		score += 1 // potential bounce
	case "bearish":
		score -= 1
	case "overbought":
		score -= 1 // caution, not a hard sell
	}

	// MACD trend + cross bonus
	switch macd.Trend {
	case "bullish":
		score += 1
	case "bearish":
		score -= 1
	}
	switch macd.Cross {
	case "golden_cross":
		score += 2
	case "death_cross":
		score -= 2
	}

	// Bollinger Bands
	switch bb.Signal {
	case "bullish":
		score += 1
	case "oversold":
		score += 1
	case "bearish":
		score -= 1
	case "overbought":
		score -= 1
	}

	return score
}

// classifySignal maps a numeric score to trend, strength, and trading signal strings.
func classifySignal(score int) (trend, strength, signal string) {
	switch {
	case score >= 5:
		return "bullish", "strong", "BUY"
	case score >= 2:
		return "bullish", "moderate", "BUY"
	case score == 1:
		return "bullish", "weak", "HOLD"
	case score == 0:
		return "neutral", "none", "HOLD"
	case score == -1:
		return "bearish", "weak", "HOLD"
	case score >= -4:
		return "bearish", "moderate", "SELL"
	default:
		return "bearish", "strong", "SELL"
	}
}

// calcConfidence converts the score into a 0–100 confidence percentage.
// Maximum possible score magnitude is 8 (EMA±2 + MACD cross±2 + MACD±1 + RSI±1 + BB±1 = 7..8).
func calcConfidence(score int) int {
	const maxScore = 8
	abs := score
	if abs < 0 {
		abs = -abs
	}
	if abs > maxScore {
		abs = maxScore
	}
	return (abs * 100) / maxScore
}

// calcKeyLevels identifies support and resistance levels from swing pivots,
// EMA200, and Bollinger Bands. Returns up to 3 of each, sorted by proximity to current price.
func calcKeyLevels(
	closes, highs, lows []float64,
	bb indicators.BollingerResult,
	ema indicators.EMAResult,
) (supports, resistances []float64) {
	current := closes[len(closes)-1]

	// Collect pivot lows and highs from the last 100 candles (or all if fewer)
	lookback := 100
	if len(closes) < lookback {
		lookback = len(closes)
	}
	const pivot = 5 // bars on each side

	var rawSupports, rawResistances []float64

	start := len(closes) - lookback
	for i := start + pivot; i < len(closes)-pivot; i++ {
		// Swing low
		isSwingLow := true
		for j := i - pivot; j <= i+pivot; j++ {
			if j != i && lows[j] <= lows[i] {
				isSwingLow = false
				break
			}
		}
		if isSwingLow {
			rawSupports = append(rawSupports, lows[i])
		}

		// Swing high
		isSwingHigh := true
		for j := i - pivot; j <= i+pivot; j++ {
			if j != i && highs[j] >= highs[i] {
				isSwingHigh = false
				break
			}
		}
		if isSwingHigh {
			rawResistances = append(rawResistances, highs[i])
		}
	}

	// Add EMA200 and BB bands as additional structural levels
	if ema.EMA200 > 0 {
		if ema.EMA200 < current {
			rawSupports = append(rawSupports, ema.EMA200)
		} else {
			rawResistances = append(rawResistances, ema.EMA200)
		}
	}
	if bb.Lower > 0 {
		rawSupports = append(rawSupports, bb.Lower)
	}
	if bb.Upper > 0 {
		rawResistances = append(rawResistances, bb.Upper)
	}

	supports = filterAndSort(rawSupports, current, false, 3)
	resistances = filterAndSort(rawResistances, current, true, 3)
	return
}

// mapFibLevels converts indicator FibLevel slice to report FibLevel slice.
func mapFibLevels(src []indicators.FibLevel) []report.FibLevel {
	out := make([]report.FibLevel, len(src))
	for i, l := range src {
		out[i] = report.FibLevel{
			Label:   l.Label,
			Ratio:   l.Ratio,
			Price:   l.Price,
			IsAbove: l.IsAbove,
		}
	}
	return out
}

// mapVPVRBins converts indicator VPVRBin slice to report VPVRBin slice.
func mapVPVRBins(src []indicators.VPVRBin) []report.VPVRBin {
	out := make([]report.VPVRBin, len(src))
	for i, b := range src {
		out[i] = report.VPVRBin{
			PriceLow:    b.PriceLow,
			PriceHigh:   b.PriceHigh,
			PriceMid:    b.PriceMid,
			Volume:      b.Volume,
			IsPOC:       b.IsPOC,
			IsValueArea: b.IsValueArea,
		}
	}
	return out
}

// filterAndSort deduplicates levels within 0.3% of each other, keeps only those
// below/above current price (depending on `above`), and returns the `limit` closest ones.
func filterAndSort(levels []float64, current float64, above bool, limit int) []float64 {
	var filtered []float64
	for _, l := range levels {
		if above && l <= current {
			continue
		}
		if !above && l >= current {
			continue
		}
		// Dedup: skip if we already have a level within 0.3% of this one
		duplicate := false
		for _, existing := range filtered {
			if existing == 0 {
				continue
			}
			diff := (l - existing) / existing
			if diff < 0 {
				diff = -diff
			}
			if diff < 0.003 {
				duplicate = true
				break
			}
		}
		if !duplicate {
			filtered = append(filtered, l)
		}
	}

	// Sort by proximity to current price
	sort.Slice(filtered, func(i, j int) bool {
		di := filtered[i] - current
		dj := filtered[j] - current
		if di < 0 {
			di = -di
		}
		if dj < 0 {
			dj = -dj
		}
		return di < dj
	})

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	// Round levels to 2 decimal places for readability
	for i, v := range filtered {
		filtered[i] = indicators.Round(v, 2)
	}
	return filtered
}
