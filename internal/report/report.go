package report

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

// Report is the top-level JSON output structure consumed by LLMs or downstream tools.
type Report struct {
	Asset      string     `json:"asset"`
	Timeframe  string     `json:"timeframe"`
	Timestamp  time.Time  `json:"timestamp"`
	Price      PriceInfo  `json:"price"`
	Indicators Indicators `json:"indicators"`
	Analysis   Analysis   `json:"analysis"`
}

// PriceInfo holds the latest OHLC snapshot and change percentage.
type PriceInfo struct {
	Current   float64 `json:"current"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	ChangePct float64 `json:"change_pct"` // % change vs previous candle close
}

// Indicators groups all computed technical indicator results.
type Indicators struct {
	RSI       RSIData       `json:"rsi"`
	MACD      MACDData      `json:"macd"`
	Bollinger BollingerData `json:"bollinger"`
	EMA       EMAData       `json:"ema"`
	ATR       ATRData       `json:"atr"`
	Volume    VolumeData    `json:"volume"`
}

// RSIData is the serialisable form of the RSI result.
type RSIData struct {
	Value  float64 `json:"value"`
	Signal string  `json:"signal"`
}

// MACDData is the serialisable form of the MACD result.
type MACDData struct {
	MACDLine   float64 `json:"macd_line"`
	SignalLine float64 `json:"signal_line"`
	Histogram  float64 `json:"histogram"`
	Trend      string  `json:"trend"`
	Cross      string  `json:"cross,omitempty"`
}

// BollingerData is the serialisable form of the Bollinger Bands result.
type BollingerData struct {
	Upper    float64 `json:"upper"`
	Middle   float64 `json:"middle"`
	Lower    float64 `json:"lower"`
	Width    float64 `json:"width"`
	PercentB float64 `json:"percent_b"`
	Position string  `json:"position"`
	Signal   string  `json:"signal"`
}

// EMAData is the serialisable form of the EMA result.
type EMAData struct {
	EMA9      float64 `json:"ema9"`
	EMA21     float64 `json:"ema21"`
	EMA50     float64 `json:"ema50"`
	EMA200    float64 `json:"ema200"`
	Alignment string  `json:"alignment"`
	Signal    string  `json:"signal"`
}

// ATRData is the serialisable form of the ATR result.
type ATRData struct {
	Value  float64 `json:"value"`
	Pct    float64 `json:"pct"`
	Regime string  `json:"regime"`
}

// VolumeData is the serialisable form of the volume result.
type VolumeData struct {
	Current float64 `json:"current"`
	SMA20   float64 `json:"sma20"`
	Ratio   float64 `json:"ratio"`
	OBV     float64 `json:"obv"`
	Signal  string  `json:"signal"`
}

// Analysis holds the aggregated signal, trend, and key price levels.
type Analysis struct {
	Trend      string    `json:"trend"`      // bullish | neutral | bearish
	Strength   string    `json:"strength"`   // strong | moderate | weak
	Signal     string    `json:"signal"`     // BUY | SELL | HOLD
	Confidence int       `json:"confidence"` // 0–100
	Score      int       `json:"score"`      // raw score for transparency
	KeyLevels  KeyLevels `json:"key_levels"`
}

// KeyLevels holds the most relevant support and resistance price levels.
type KeyLevels struct {
	Support    []float64 `json:"support"`
	Resistance []float64 `json:"resistance"`
}

// WriteJSON encodes the report as indented JSON to the given writer.
func (r *Report) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// Save writes the report as indented JSON to the specified file path.
func (r *Report) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return r.WriteJSON(f)
}

// ---------------------------------------------------------------------------
// Multi-timeframe report
// ---------------------------------------------------------------------------

// MultiReport holds analysis results across multiple timeframes for one asset.
type MultiReport struct {
	Asset      string               `json:"asset"`
	Timestamp  time.Time            `json:"timestamp"`
	Timeframes map[string]*Report   `json:"timeframes"` // key: interval string
	Summary    MultiSummary         `json:"summary"`
}

// MultiSummary is the cross-timeframe aggregated signal.
type MultiSummary struct {
	Alignment      string            `json:"alignment"`       // all_bullish | mostly_bullish | mixed | mostly_bearish | all_bearish
	BullishCount   int               `json:"bullish_count"`
	BearishCount   int               `json:"bearish_count"`
	NeutralCount   int               `json:"neutral_count"`
	DominantSignal string            `json:"dominant_signal"` // BUY | SELL | HOLD
	Confidence     int               `json:"confidence"`      // 0–100
	Signals        map[string]string `json:"signals"`         // interval → BUY/SELL/HOLD
	Trends         map[string]string `json:"trends"`          // interval → bullish/neutral/bearish
}

// WriteJSON encodes the multi report as indented JSON.
func (m *MultiReport) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(m)
}

// Save writes the multi report to the specified file path.
func (m *MultiReport) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return m.WriteJSON(f)
}
