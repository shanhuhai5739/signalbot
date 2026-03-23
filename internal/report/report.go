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
	RSI         RSIData              `json:"rsi"`
	MACD        MACDData             `json:"macd"`
	Bollinger   BollingerData        `json:"bollinger"`
	EMA         EMAData              `json:"ema"`
	ATR         ATRData              `json:"atr"`
	Volume      VolumeData           `json:"volume"`
	Guppy       GuppyData            `json:"guppy"`
	GuppyHistory []GuppyHistoryEntry `json:"guppy_history,omitempty"`
	Fibonacci   FibonacciData        `json:"fibonacci"`
	VWAP        VWAPData             `json:"vwap"`
	VPVR        VPVRData             `json:"vpvr"`
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

// ---------------------------------------------------------------------------
// New indicator data types
// ---------------------------------------------------------------------------

// GuppyEMAItem is a single EMA in the Guppy groups.
type GuppyEMAItem struct {
	Period int     `json:"period"`
	Value  float64 `json:"value"`
}

// GuppyData is the serialisable form of the GMMA result.
// ShortEMAs and LongEMAs use arrays so custom period configurations are
// represented without any fixed field names.
type GuppyData struct {
	ShortEMAs []GuppyEMAItem `json:"short_emas"` // fast group (traders)
	LongEMAs  []GuppyEMAItem `json:"long_emas"`  // slow group (investors)
	ShortMin  float64        `json:"short_min"`
	ShortMax  float64        `json:"short_max"`
	LongMin   float64        `json:"long_min"`
	LongMax   float64        `json:"long_max"`
	GapPct    float64        `json:"gap_pct"`
	Alignment string         `json:"alignment"`
	Signal    string         `json:"signal"`
}

// GuppyHistoryEntry is one historical GMMA snapshot (bar_index 0 = most recent).
type GuppyHistoryEntry struct {
	BarIndex  int            `json:"bar_index"`
	Close     float64        `json:"close"`
	ShortEMAs []GuppyEMAItem `json:"short_emas"`
	LongEMAs  []GuppyEMAItem `json:"long_emas"`
	ShortMin  float64        `json:"short_min"`
	ShortMax  float64        `json:"short_max"`
	LongMin   float64        `json:"long_min"`
	LongMax   float64        `json:"long_max"`
	GapPct    float64        `json:"gap_pct"`
	Alignment string         `json:"alignment"`
	Signal    string         `json:"signal"`
}

// FibLevel is a single Fibonacci retracement level.
type FibLevel struct {
	Label   string  `json:"label"`
	Ratio   float64 `json:"ratio"`
	Price   float64 `json:"price"`
	IsAbove bool    `json:"is_above"`
}

// FibonacciData is the serialisable form of the Fibonacci retracement result.
type FibonacciData struct {
	SwingHigh    float64    `json:"swing_high"`
	SwingLow     float64    `json:"swing_low"`
	Range        float64    `json:"range"`
	Levels       []FibLevel `json:"levels"`
	NearestLevel string     `json:"nearest_level"`
	DistancePct  float64    `json:"distance_pct"`
	Signal       string     `json:"signal"`
	Direction    string     `json:"direction"`
}

// VWAPData is the serialisable form of the Anchored VWAP result.
type VWAPData struct {
	Value        float64 `json:"value"`
	UpperBand1   float64 `json:"upper_band1"`
	LowerBand1   float64 `json:"lower_band1"`
	UpperBand2   float64 `json:"upper_band2"`
	LowerBand2   float64 `json:"lower_band2"`
	StdDev       float64 `json:"std_dev"`
	DeviationPct float64 `json:"deviation_pct"`
	Position     string  `json:"position"`
	Signal       string  `json:"signal"`
}

// VPVRBin is a single price-level bucket in the volume profile.
type VPVRBin struct {
	PriceLow    float64 `json:"price_low"`
	PriceHigh   float64 `json:"price_high"`
	PriceMid    float64 `json:"price_mid"`
	Volume      float64 `json:"volume"`
	IsPOC       bool    `json:"is_poc"`
	IsValueArea bool    `json:"is_value_area"`
}

// VPVRData is the serialisable form of the Fixed Range Volume Profile result.
type VPVRData struct {
	POC     float64   `json:"poc"`
	VAH     float64   `json:"vah"`
	VAL     float64   `json:"val"`
	NumBins int       `json:"num_bins"`
	Bins    []VPVRBin `json:"bins"`
	Signal  string    `json:"signal"`
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
