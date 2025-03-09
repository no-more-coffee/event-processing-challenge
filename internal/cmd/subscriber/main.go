package main

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	redisAddr, ok := os.LookupEnv("REDIS_ADDR")
	if !ok {
		panic("env var unset: REDIS_ADDR")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	pubsub := rdb.Subscribe(ctx, "events")
	defer pubsub.Close()

	log.Println("Listening...")
	ch := pubsub.Channel()
	for msg := range ch {
		log.Println(msg.Payload)
	}
}
