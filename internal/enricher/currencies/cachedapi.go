package currencies

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type CachedApi struct {
	ExchangeApi
	Timeout time.Duration
	Db      *redis.Client
}

// Ensure implementation
var _ ExchangeApi = (*CachedApi)(nil)

func (e CachedApi) Fetch(ctx context.Context) (Rates, error) {
	var rates Rates
	key := "cached"

	// Try cache
	values, err := e.Db.HGetAll(ctx, key).Result()
	if err != nil {
		return rates, err
	}
	if len(values) > 0 {
		rates.Quotes = make(map[string]float64)
		for k, v := range values {
			num, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return rates, err
			}
			rates.Quotes[k] = float64(num)
		}
		log.Println("Using cached rates")
		return rates, nil
	}

	// Do request
	log.Println("Fetching rates")
	rates, err = e.ExchangeApi.Fetch(ctx)
	if err != nil {
		return rates, err
	}

	// Update cache
	values = make(map[string]string)
	for k, v := range rates.Quotes {
		values[k] = strconv.FormatFloat(float64(v), 'g', 2, 64)
	}
	if err := e.Db.HSet(ctx, key, values).Err(); err != nil {
		return rates, err
	}
	if err := e.Db.Expire(ctx, key, e.Timeout).Err(); err != nil {
		return rates, err
	}

	return rates, nil
}
