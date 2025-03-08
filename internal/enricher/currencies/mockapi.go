package currencies

import (
	"context"
	"time"
)

type ExchangeApiMock struct{}

// Ensure implementation
var _ ExchangeApi = (*ExchangeApiMock)(nil)

func (e ExchangeApiMock) Fetch(ctx context.Context) (Rates, error) {
	time.Sleep(100 * time.Millisecond)
	return Rates{
		Quotes: map[string]float64{
			"EURBTC": 1.2553402e-05,
			"EURGBP": 0.839814,
			"EURNZD": 1.897485,
			"EURUSD": 1.083654,
		},
	}, nil
}
