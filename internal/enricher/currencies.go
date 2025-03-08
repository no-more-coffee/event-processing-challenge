package enricher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
)

func RunDescription(
	ctx context.Context,
	in <-chan casino.Event,
	out chan<- casino.Event,
) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("RunDescription Done")
			return nil
		case event := <-in:
			log.Println("in RunDescription received", event)
			AddDescription(&event)
			out <- event
		}
	}
}

func RunCurrencies(
	ctx context.Context,
	exchangeApi ExchangeApi,
	in <-chan casino.Event,
	out chan<- casino.Event,
) error {

	for {
		select {
		case <-ctx.Done():
			log.Println("RunCurrencies Done")
			return nil
		case event := <-in:
			// Blocks here to allow only single request
			log.Println("in RunCurrencies received", event)
			amountEur, err := ConvertCurrency(
				ctx,
				exchangeApi,
				float64(event.Amount),
				event.Currency,
			)
			if err != nil {
				return err
			}
			event.AmountEUR = amountEur

			out <- event
		}
	}
}

func ConvertCurrency(
	ctx context.Context,
	api ExchangeApi,
	value float64,
	currency string,
) (float64, error) {
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
	Quotes map[string]float64 `json:"quotes"`
}

type ExchangeApi interface {
	Fetch(context.Context) (Rates, error)
}

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

func AddDescription(event *casino.Event) {
	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("Player #%d ", event.PlayerID))
	if event.Player != nil && event.Player.Email != "" {
		builder.WriteString(fmt.Sprintf("(%s) ", event.Player.Email))
	}

	switch event.Type {
	case "game_start":
		builder.WriteString("started playing a game ")
	case "bet":
		if event.HasWon {
			builder.WriteString("sucessfully ")
		}
		builder.WriteString(fmt.Sprintf(
			"placed a bet of %d %s ",
			event.Amount,
			event.Currency,
		))
		if event.Currency != "EUR" {
			builder.WriteString(fmt.Sprintf("(%f EUR) ", event.AmountEUR))
		}
		builder.WriteString("on a game ")
	case "deposit":
		builder.WriteString(fmt.Sprintf(
			"made a deposit of %d %s ",
			event.Amount,
			event.Currency,
		))
	case "game_stop":
		builder.WriteString("stopped playing a game ")
	}

	game, ok := casino.Games[event.GameID]
	if ok && event.Type != "deposit" {
		builder.WriteString(fmt.Sprintf("\"%s\" ", game.Title))
	}

	formattedTime := event.CreatedAt.Format("on January 2nd, 2006 at 15:04 MST.")
	builder.WriteString(formattedTime)

	event.Description = builder.String()
}
