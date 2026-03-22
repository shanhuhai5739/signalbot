package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	BinanceBaseURL        string
	BinanceFuturesBaseURL string
	HTTPTimeout           time.Duration
	DefaultLimit          int
}

// Load reads configuration from environment variables with sensible defaults.
// No API key is required since Binance public market data endpoints are unauthenticated.
//
// Supported environment variables:
//   BINANCE_BASE_URL          Override Binance spot API base URL. Default: https://api.binance.com
//   BINANCE_FUTURES_BASE_URL  Override Binance USD-M futures API base URL. Default: https://fapi.binance.com
//   HTTP_TIMEOUT_SEC          HTTP request timeout in seconds. Default: 15
//   DEFAULT_LIMIT             Number of candles to fetch when --limit is not specified. Default: 200
func Load() *Config {
	baseURL := os.Getenv("BINANCE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.binance.com"
	}

	futuresBaseURL := os.Getenv("BINANCE_FUTURES_BASE_URL")
	if futuresBaseURL == "" {
		futuresBaseURL = "https://fapi.binance.com"
	}

	timeout := 15 * time.Second
	if v := os.Getenv("HTTP_TIMEOUT_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeout = time.Duration(n) * time.Second
		}
	}

	limit := 200
	if v := os.Getenv("DEFAULT_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	return &Config{
		BinanceBaseURL:        baseURL,
		BinanceFuturesBaseURL: futuresBaseURL,
		HTTPTimeout:           timeout,
		DefaultLimit:          limit,
	}
}
