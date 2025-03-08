package main

import (
	"context"
	"log"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/generator"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/publisher"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rdb := redis.NewClient(&redis.Options{
		Addr: ":6379",
	})

	eventCh := generator.Generate(ctx)

	for event := range eventCh {
		if err := publisher.Publish(ctx, rdb, "events", event); err != nil {
			panic(err)
		}
	}

	log.Println("finished")
}
