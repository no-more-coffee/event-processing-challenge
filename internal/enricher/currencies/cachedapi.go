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

const cacheKey = "cached"

func (e CachedApi) Fetch(ctx context.Context) (Rates, error) {
	var rates Rates

	cached, err := e.getCached(ctx)
	if err != nil {
		return rates, err
	}
	if cached != nil {
		log.Println("Using cached rates")
		return *cached, nil
	}

	log.Println("Fetching rates")
	rates, err = e.ExchangeApi.Fetch(ctx)
	if err != nil {
		return rates, err
	}

	log.Printf("Updating cache: %+v", rates.Quotes)
	if err := e.updateCache(ctx, rates); err != nil {
		return rates, err
	}

	return rates, nil
}

func (e CachedApi) getCached(ctx context.Context) (*Rates, error) {
	values, err := e.Db.HGetAll(ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}

	rates := Rates{
		Quotes: make(map[string]float64),
	}
	for k, v := range values {
		num, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, err
		}
		rates.Quotes[k] = num
	}

	return &rates, nil
}

func (e CachedApi) updateCache(ctx context.Context, rates Rates) error {
	values := make(map[string]any)
	for k, v := range rates.Quotes {
		values[k] = v
	}
	if err := e.Db.HSet(ctx, cacheKey, values).Err(); err != nil {
		return err
	}
	if err := e.Db.Expire(ctx, cacheKey, e.Timeout).Err(); err != nil {
		return err
	}

	return nil
}
