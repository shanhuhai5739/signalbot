package data

import (
	"context"

	"signalbot/config"
)

// Provider is a unified market data gateway backed solely by Binance.
// Spot assets (BTC, ETH, SOL, …) are fetched from the Binance spot API.
// Futures-only assets (XAUUSD) are transparently routed to the Binance USD-M futures API.
type Provider struct {
	binance *BinanceClient
}

// NewProvider creates a Provider backed by Binance.
func NewProvider(cfg *config.Config) *Provider {
	return &Provider{
		binance: NewBinanceClient(cfg),
	}
}

// FetchKlines retrieves OHLCV candles for the given asset and interval.
// Routing between spot and futures is handled automatically.
func (p *Provider) FetchKlines(ctx context.Context, asset, interval string, limit int) ([]Candle, error) {
	return p.binance.FetchKlines(ctx, asset, interval, limit)
}
