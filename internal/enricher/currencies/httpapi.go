package currencies

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type ExchangeApiHttp struct {
	BaseUrl    string
	ApiKey     string
	Currencies string
}

// Ensure implementation
var _ ExchangeApi = (*ExchangeApiHttp)(nil)

func (e ExchangeApiHttp) Fetch(ctx context.Context) (Rates, error) {
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
