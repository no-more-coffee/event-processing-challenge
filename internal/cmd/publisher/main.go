package main

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/enricher"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/enricher/currencies"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/generator"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/publisher"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	parentCtx := context.Background()
	ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer cancel()

	redisAddr, ok := os.LookupEnv("REDIS_ADDR")
	if !ok {
		panic("env var unset: REDIS_ADDR")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	rPub := publisher.RedisPublisher{
		Db: rdb,
	}

	dsn, ok := os.LookupEnv("PG_DSN")
	if !ok {
		panic("env var unset: PG_DSN")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	// pgDb :=enricher.PlayerDataMock{}
	pgDb := enricher.PlayerDataPg{
		Db: db,
	}

	apiKey, ok := os.LookupEnv("API_KEY")
	if !ok {
		panic("env var unset: API_KEY")
	}
	exchangeApi := currencies.ExchangeApiHttp{
		BaseUrl:    "http://api.exchangerate.host/live",
		ApiKey:     apiKey,
		Currencies: strings.Join(casino.Currencies, ","),
	}
	// exchangeApi := currencies.ExchangeApiMock{}
	cachedApi := currencies.CachedApi{
		ExchangeApi: exchangeApi,
		Timeout:     5 * time.Second,
		Db:          rdb,
	}

	var wg sync.WaitGroup
	curCh := make(chan casino.Event, 1000)
	pgCh := make(chan casino.Event, 1000)
	pubCh := make(chan casino.Event, 1000)
	descCh := make(chan casino.Event, 1000)

	wg.Add(1)
	go func(parent context.Context) {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(parent, 5*time.Second)
		defer cancel()

		eventCh := generator.Generate(ctx)

		for event := range eventCh {
			curCh <- event
		}
	}(parentCtx)

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := currencies.RunCurrencies(
			ctx,
			cachedApi,
			curCh,
			pgCh,
		); err != nil {
			panic(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := enricher.RunPlayerData(
			ctx,
			pgDb,
			pgCh,
			descCh,
		); err != nil {
			panic(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := enricher.RunDescription(
			ctx,
			descCh,
			pubCh,
		); err != nil {
			panic(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := publisher.RunPublish(
			ctx,
			rPub,
			pubCh,
		); err != nil {
			panic(err)
		}
	}()

	wg.Wait()
	log.Println("finished")
}
