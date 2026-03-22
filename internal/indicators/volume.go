package indicators

// VolumeResult holds volume statistics and On-Balance Volume (OBV).
type VolumeResult struct {
	Current float64 `json:"current"`
	SMA20   float64 `json:"sma20"`
	Ratio   float64 `json:"ratio"`  // current / SMA20
	OBV     float64 `json:"obv"`    // On-Balance Volume (cumulative)
	Signal  string  `json:"signal"` // high_volume | above_average | normal | below_average | low_volume
}

// CalcVolume computes volume statistics and OBV.
// volumes and closes must be the same length and contain at least 21 elements.
func CalcVolume(volumes, closes []float64) VolumeResult {
	const smaPeriod = 20
	n := len(volumes)
	if n < smaPeriod+1 || len(closes) != n {
		return VolumeResult{Signal: "normal"}
	}

	sma := SMA(volumes, smaPeriod)
	current := volumes[n-1]

	ratio := 0.0
	if sma > 0 {
		ratio = current / sma
	}

	// OBV: add volume on up-closes, subtract on down-closes
	obv := 0.0
	for i := 1; i < n; i++ {
		switch {
		case closes[i] > closes[i-1]:
			obv += volumes[i]
		case closes[i] < closes[i-1]:
			obv -= volumes[i]
		}
	}

	return VolumeResult{
		Current: Round(current, 2),
		SMA20:   Round(sma, 2),
		Ratio:   Round(ratio, 2),
		OBV:     Round(obv, 0),
		Signal:  volumeSignal(ratio),
	}
}

func volumeSignal(ratio float64) string {
	switch {
	case ratio >= 2.0:
		return "high_volume"
	case ratio >= 1.3:
		return "above_average"
	case ratio <= 0.5:
		return "low_volume"
	case ratio <= 0.7:
		return "below_average"
	default:
		return "normal"
	}
}
