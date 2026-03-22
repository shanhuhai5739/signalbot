package data

import "time"

// Candle represents a single OHLCV candlestick.
type Candle struct {
	OpenTime  time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	CloseTime time.Time
}

// ExtractCloses returns the closing prices from a slice of candles.
func ExtractCloses(candles []Candle) []float64 {
	out := make([]float64, len(candles))
	for i, c := range candles {
		out[i] = c.Close
	}
	return out
}

// ExtractHighs returns the high prices from a slice of candles.
func ExtractHighs(candles []Candle) []float64 {
	out := make([]float64, len(candles))
	for i, c := range candles {
		out[i] = c.High
	}
	return out
}

// ExtractLows returns the low prices from a slice of candles.
func ExtractLows(candles []Candle) []float64 {
	out := make([]float64, len(candles))
	for i, c := range candles {
		out[i] = c.Low
	}
	return out
}

// ExtractOpens returns the open prices from a slice of candles.
func ExtractOpens(candles []Candle) []float64 {
	out := make([]float64, len(candles))
	for i, c := range candles {
		out[i] = c.Open
	}
	return out
}

// ExtractVolumes returns the volumes from a slice of candles.
func ExtractVolumes(candles []Candle) []float64 {
	out := make([]float64, len(candles))
	for i, c := range candles {
		out[i] = c.Volume
	}
	return out
}
