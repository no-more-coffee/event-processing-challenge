package enricher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

func ConvertCurrency(
	ctx context.Context,
	api ExchangeApi,
	value float32,
	currency string,
) (float32, error) {
	if currency == "EUR" {
		return value, nil
	}

	rates, err := api.Fetch(ctx)
	if err != nil {
		return 0, err
	}

	conversion := fmt.Sprintf("EUR%s", currency)
	rate, ok := rates.Quotes[conversion]
	if !ok {
		return 0, fmt.Errorf("couldn't get rates for currency: %s", currency)
	}

	return value / rate, nil
}

type Rates struct {
	Quotes map[string]float32 `json:"quotes"`
}

type ExchangeApi interface {
	Fetch(context.Context) (Rates, error)
}

type ExchangeApiReal struct {
	BaseUrl    string
	ApiKey     string
	Currencies string
}

// Ensure implementation
var _ ExchangeApi = (*ExchangeApiReal)(nil)

func (e ExchangeApiReal) Fetch(ctx context.Context) (Rates, error) {
	var rates Rates

	u, err := url.Parse(e.BaseUrl)
	if err != nil {
		return rates, err
	}
	q := u.Query()
	q.Set("access_key", e.ApiKey)
	q.Set("source", "EUR")
	q.Set("currencies", e.Currencies)
	u.RawQuery = q.Encode()

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return rates, err
	}

	res, err := client.Do(req)
	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != http.StatusOK {
		return rates, fmt.Errorf("request failed with status: %s", res.Status)
	}

	if err := json.NewDecoder(res.Body).Decode(&rates); err != nil {
		return rates, err
	}

	log.Println(rates)
	return rates, err
}

type ExchangeApiMock struct{}

// Ensure implementation
var _ ExchangeApi = (*ExchangeApiMock)(nil)

func (e ExchangeApiMock) Fetch(ctx context.Context) (Rates, error) {
	time.Sleep(10 * time.Millisecond)
	return Rates{
		Quotes: map[string]float32{
			"EURBTC": 1.2553402e-05,
			"EURGBP": 0.839814,
			"EURNZD": 1.897485,
			"EURUSD": 1.083654,
		},
	}, nil
}
