package analysis

import (
	"context"
	"sync"
	"time"

	"signalbot/internal/data"
	"signalbot/internal/report"
)

// TimeframeConfig defines a single timeframe for multi-timeframe analysis.
type TimeframeConfig struct {
	Interval string
	Label    string // human-readable Chinese label
	Limit    int    // number of candles to fetch
}

// DefaultTimeframes is the standard set of timeframes for multi-timeframe analysis.
// 年线趋势通过月线（1M）的长期走势来判断，Binance 不提供年度 K 线。
var DefaultTimeframes = []TimeframeConfig{
	{"1h", "1小时线", 210},
	{"4h", "4小时线", 210},
	{"1d", "日线", 210},
	{"1w", "周线", 210},
	{"1M", "月线（年线趋势参考）", 60}, // 60个月=5年，足以判断年线方向
}

type tfResult struct {
	interval string
	rep      *report.Report
	err      error
}

// AnalyzeMulti concurrently fetches all timeframes for the given asset and
// returns a MultiReport with per-timeframe analysis and a cross-timeframe summary.
func AnalyzeMulti(ctx context.Context, asset string, provider *data.Provider) *report.MultiReport {
	ch := make(chan tfResult, len(DefaultTimeframes))
	var wg sync.WaitGroup

	for _, tf := range DefaultTimeframes {
		wg.Add(1)
		go func(cfg TimeframeConfig) {
			defer wg.Done()
			candles, err := provider.FetchKlines(ctx, asset, cfg.Interval, cfg.Limit)
			if err != nil {
				ch <- tfResult{cfg.Interval, nil, err}
				return
			}
			if len(candles) < 10 {
				ch <- tfResult{cfg.Interval, nil, nil}
				return
			}
			rep := Analyze(asset, cfg.Interval, candles)
			ch <- tfResult{cfg.Interval, rep, nil}
		}(tf)
	}

	wg.Wait()
	close(ch)

	// Collect results, preserving order
	tfMap := make(map[string]*report.Report, len(DefaultTimeframes))
	for r := range ch {
		if r.err == nil && r.rep != nil {
			tfMap[r.interval] = r.rep
		}
	}

	summary := buildSummary(tfMap)

	return &report.MultiReport{
		Asset:      asset,
		Timestamp:  time.Now().UTC(),
		Timeframes: tfMap,
		Summary:    summary,
	}
}

// buildSummary aggregates per-timeframe signals into a cross-timeframe summary.
func buildSummary(tfMap map[string]*report.Report) report.MultiSummary {
	signals := make(map[string]string, len(tfMap))
	trends := make(map[string]string, len(tfMap))

	var bullish, bearish, neutral int

	for interval, rep := range tfMap {
		signals[interval] = rep.Analysis.Signal
		trends[interval] = rep.Analysis.Trend

		switch rep.Analysis.Trend {
		case "bullish":
			bullish++
		case "bearish":
			bearish++
		default:
			neutral++
		}
	}

	total := bullish + bearish + neutral
	if total == 0 {
		return report.MultiSummary{
			Alignment:      "unknown",
			DominantSignal: "HOLD",
			Signals:        signals,
			Trends:         trends,
		}
	}

	alignment := classifyAlignment(bullish, bearish, neutral, total)
	dominantSignal, confidence := dominantSignalAndConfidence(bullish, bearish, neutral, total)

	return report.MultiSummary{
		Alignment:      alignment,
		BullishCount:   bullish,
		BearishCount:   bearish,
		NeutralCount:   neutral,
		DominantSignal: dominantSignal,
		Confidence:     confidence,
		Signals:        signals,
		Trends:         trends,
	}
}

// classifyAlignment maps bull/bear/neutral counts to a named alignment category.
func classifyAlignment(bullish, bearish, neutral, total int) string {
	bullPct := float64(bullish) / float64(total)
	bearPct := float64(bearish) / float64(total)

	switch {
	case bullPct >= 0.8:
		return "all_bullish"
	case bullPct >= 0.6:
		return "mostly_bullish"
	case bearPct >= 0.8:
		return "all_bearish"
	case bearPct >= 0.6:
		return "mostly_bearish"
	default:
		return "mixed"
	}
}

// dominantSignalAndConfidence returns the overall trading signal and a confidence score.
// Confidence reflects how strongly the timeframes agree.
func dominantSignalAndConfidence(bullish, bearish, neutral, total int) (signal string, confidence int) {
	if bullish > bearish && bullish > neutral {
		signal = "BUY"
		confidence = (bullish * 100) / total
	} else if bearish > bullish && bearish > neutral {
		signal = "SELL"
		confidence = (bearish * 100) / total
	} else {
		signal = "HOLD"
		// Confidence is low when mixed — reflect the dominant non-neutral count
		if bullish >= bearish {
			confidence = (bullish * 100) / total
		} else {
			confidence = (bearish * 100) / total
		}
	}
	return
}
