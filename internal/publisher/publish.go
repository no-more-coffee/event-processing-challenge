package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/redis/go-redis/v9"
)

func RunPublish(
	ctx context.Context,
	publisher Publisher,
	in <-chan casino.Event,
) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("RunPublish Done")
			return nil
		case event := <-in:
			log.Println("in RunPublish received", event)
			if err := publisher.Publish(ctx, "events", event); err != nil {
				return err
			}
		}
	}
}

type Publisher interface {
	Publish(ctx context.Context, channel string, event casino.Event) error
}

type RedisPublisher struct {
	Db *redis.Client
}

// Ensure implementation
var _ Publisher = (*RedisPublisher)(nil)

type Event struct{ casino.Event }

func (e Event) MarshalBinary() ([]byte, error) {
	return json.Marshal(e)
}

func (p RedisPublisher) Publish(
	ctx context.Context,
	channel string,
	event casino.Event,
) error {
	if err := p.Db.Publish(ctx, channel, Event{event}).Err(); err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	return nil
}
