package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/enricher/currencies"
)

func main() {
	apiKey, ok := os.LookupEnv("API_KEY")
	if !ok {
		panic("env var unset: API_KEY")
	}

	log.Println(currencies.ConvertCurrency(
		context.TODO(),
		currencies.ExchangeApiHttp{
			BaseUrl:    "http://api.exchangerate.host/live",
			ApiKey:     apiKey,
			Currencies: strings.Join(casino.Currencies, ","),
		},
		1,
		"USD",
	))
}
