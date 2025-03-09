package playerdata

import (
	"context"
	"errors"
	"log"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"gorm.io/gorm"
)

type PlayerData interface {
	Find(playerId int) (casino.Player, error)
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
