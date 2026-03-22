package data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"signalbot/config"
)

// symbolMap maps friendly asset names to Binance trading pair symbols.
// Add more pairs here as new assets are supported.
var symbolMap = map[string]string{
	"BTC":    "BTCUSDT",
	"XAUUSD": "XAUUSDT",
	"ETH":    "ETHUSDT",
	"SOL":    "SOLUSDT",
	"BNB":    "BNBUSDT",
}

// BinanceClient fetches public market data from the Binance REST API.
// No authentication is required for klines data.
type BinanceClient struct {
	cfg    *config.Config
	client *http.Client
}

// NewBinanceClient creates a new client using the provided configuration.
func NewBinanceClient(cfg *config.Config) *BinanceClient {
	return &BinanceClient{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.HTTPTimeout},
	}
}

// FetchKlines retrieves OHLCV candlestick data for the given asset and interval.
// asset can be a friendly name ("BTC", "XAUUSD") or a raw Binance symbol ("BTCUSDT").
// interval follows Binance conventions: 1m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d, 3d, 1w, 1M.
// limit controls how many candles to return (max 1500).
func (b *BinanceClient) FetchKlines(ctx context.Context, asset, interval string, limit int) ([]Candle, error) {
	symbol := resolveSymbol(asset)

	url := fmt.Sprintf(
		"%s/api/v3/klines?symbol=%s&interval=%s&limit=%d",
		b.cfg.BinanceBaseURL, symbol, interval, limit,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "signalbot/1.0")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch klines (%s %s): %w", symbol, interval, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance returned HTTP %d for %s %s", resp.StatusCode, symbol, interval)
	}

	// Each element is a JSON array: [openTime, open, high, low, close, volume, closeTime, ...]
	var raw [][]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode klines response: %w", err)
	}

	candles := make([]Candle, 0, len(raw))
	for _, row := range raw {
		c, err := parseKlineRow(row)
		if err != nil {
			continue
		}
		candles = append(candles, c)
	}

	if len(candles) == 0 {
		return nil, fmt.Errorf("no candles returned for %s %s — symbol may not exist on Binance", symbol, interval)
	}

	return candles, nil
}

// resolveSymbol maps a user-facing asset name to a Binance trading symbol.
func resolveSymbol(asset string) string {
	upper := strings.ToUpper(asset)
	if sym, ok := symbolMap[upper]; ok {
		return sym
	}
	return upper
}

// parseKlineRow parses a single Binance kline array into a Candle.
// Binance kline format (indices):
//   0: open time (ms), 1: open, 2: high, 3: low, 4: close, 5: volume,
//   6: close time (ms), 7: quote asset volume, 8: number of trades, ...
func parseKlineRow(row []json.RawMessage) (Candle, error) {
	if len(row) < 7 {
		return Candle{}, fmt.Errorf("unexpected kline row length %d", len(row))
	}

	var openTimeMs, closeTimeMs int64
	if err := json.Unmarshal(row[0], &openTimeMs); err != nil {
		return Candle{}, fmt.Errorf("parse open time: %w", err)
	}
	if err := json.Unmarshal(row[6], &closeTimeMs); err != nil {
		return Candle{}, fmt.Errorf("parse close time: %w", err)
	}

	open, err := parseStringFloat(row[1])
	if err != nil {
		return Candle{}, fmt.Errorf("parse open: %w", err)
	}
	high, err := parseStringFloat(row[2])
	if err != nil {
		return Candle{}, fmt.Errorf("parse high: %w", err)
	}
	low, err := parseStringFloat(row[3])
	if err != nil {
		return Candle{}, fmt.Errorf("parse low: %w", err)
	}
	close_, err := parseStringFloat(row[4])
	if err != nil {
		return Candle{}, fmt.Errorf("parse close: %w", err)
	}
	volume, err := parseStringFloat(row[5])
	if err != nil {
		return Candle{}, fmt.Errorf("parse volume: %w", err)
	}

	return Candle{
		OpenTime:  time.UnixMilli(openTimeMs).UTC(),
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close_,
		Volume:    volume,
		CloseTime: time.UnixMilli(closeTimeMs).UTC(),
	}, nil
}

// parseStringFloat unmarshals a JSON-quoted float string (e.g. "87234.50") into float64.
func parseStringFloat(raw json.RawMessage) (float64, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return 0, err
	}
	return strconv.ParseFloat(s, 64)
}
