package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/enricher"
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
	// pgDb :=PlayerDataMock{}
	pgDb := PlayerDataPg{
		Db: db,
	}

	cachedApi := CachedApi{
		ExchangeApi: enricher.ExchangeApiMock{},
		timeout:     3 * time.Second,
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

		if err := enricher.RunCurrencies(
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

		if err := RunPlayerData(
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

type PlayerData interface {
	Find(playerId int) (casino.Player, error)
}

type CachedApi struct {
	enricher.ExchangeApi
	timeout time.Duration
	Db      *redis.Client
}

func (e CachedApi) Fetch(ctx context.Context) (enricher.Rates, error) {
	var rates enricher.Rates
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
	if err := e.Db.Expire(ctx, key, e.timeout).Err(); err != nil {
		return rates, err
	}

	return rates, nil
}

func RunPlayerData(
	ctx context.Context,
	playerData PlayerData,
	in <-chan casino.Event,
	out chan<- casino.Event,
) error {

	for {
		select {
		case <-ctx.Done():
			log.Println("RunPlayerData Done")
			return nil
		case event := <-in:
			log.Println("in RunPlayerData received", event)

			go func() {
				if err := AddPlayerData(
					ctx,
					playerData,
					&event,
				); err != nil {
					panic(err)
				}

				out <- event
			}()
		}
	}
}

func AddPlayerData(ctx context.Context, dbData PlayerData, event *casino.Event) error {
	player, err := dbData.Find(event.PlayerID)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println("Player not found:", event.PlayerID)
		return nil
	}

	if err != nil {
		return err
	}

	// Store DB data into Event field
	event.Player = &player
	return nil
}

type PlayerDataPg struct {
	Db *gorm.DB
}

// implements PlayerData
var _ PlayerData = (*PlayerDataPg)(nil)

func (p PlayerDataPg) Find(playerId int) (casino.Player, error) {
	var player casino.Player
	if tx := p.Db.Take(&player, playerId); tx.Error != nil {
		return player, tx.Error
	}

	return player, nil
}

type PlayerDataMock struct{}

// implements PlayerData
var _ PlayerData = (*PlayerDataMock)(nil)

func (p PlayerDataMock) Find(playerId int) (casino.Player, error) {
	if playerId > 9 && playerId < 15 {
		return casino.Player{
			ID:             playerId,
			Email:          "1@1.com",
			LastSignedInAt: time.Now(),
		}, nil
	}
	return casino.Player{}, gorm.ErrRecordNotFound
}
