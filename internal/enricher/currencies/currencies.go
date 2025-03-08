package currencies

import (
	"context"
	"fmt"
	"log"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
)

type Rates struct {
	Quotes map[string]float64 `json:"quotes"`
	// We don't use the other fields
}

type ExchangeApi interface {
	Fetch(context.Context) (Rates, error)
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
