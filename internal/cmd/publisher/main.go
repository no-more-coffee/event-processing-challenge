package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/enricher"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/generator"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/publisher"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var ctx = context.Background()

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	redisAddr, ok := os.LookupEnv("REDIS_ADDR")
	if !ok {
		panic("env var unset: REDIS_ADDR")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	eventCh := generator.Generate(ctx)

	dsn, ok := os.LookupEnv("PG_DSN")
	if !ok {
		panic("env var unset: PG_DSN")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	for event := range eventCh {
		if err := AddPlayerData(
			ctx,
			// PlayerDataMock{},
			PlayerDataPg{
				Db: db,
			},
			&event,
		); err != nil {
			panic(err)
		}

		amountEur, err := enricher.ConvertCurrency(
			ctx,
			enricher.ExchangeApiMock{},
			float32(event.Amount),
			event.Currency,
		)
		if err != nil {
			panic(err)
		}
		event.AmountEUR = amountEur

		if err := publisher.Publish(ctx, rdb, "events", event); err != nil {
			panic(err)
		}
	}

	log.Println("finished")
}

type PlayerData interface {
	Find(playerId int) (casino.Player, error)
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
