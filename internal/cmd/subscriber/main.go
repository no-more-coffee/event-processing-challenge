package main

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: ":6379",
	})

	pubsub := rdb.Subscribe(ctx, "events")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		log.Println(msg.Payload)
	}
}
