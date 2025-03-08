package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/redis/go-redis/v9"
)

type Event struct{ casino.Event }

func (e Event) MarshalBinary() ([]byte, error) {
	return json.Marshal(e)
}

func Publish(
	ctx context.Context,
	rdb *redis.Client,
	channel string,
	event casino.Event,
) error {
	if err := rdb.Publish(ctx, channel, Event{event}).Err(); err != nil {
		return fmt.Errorf("Publish: %w", err)
	}

	return nil
}
