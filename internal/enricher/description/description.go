package description

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
)

func RunDescription(
	ctx context.Context,
	in <-chan casino.Event,
	out chan<- casino.Event,
) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("RunDescription Done")
			return nil
		case event := <-in:
			log.Println("in RunDescription received", event)
			AddDescription(&event)
			out <- event
		}
	}
}

func AddDescription(event *casino.Event) {
	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("Player #%d ", event.PlayerID))
	if event.Player != nil && event.Player.Email != "" {
		builder.WriteString(fmt.Sprintf("(%s) ", event.Player.Email))
	}

	switch event.Type {
	case "game_start":
		builder.WriteString("started playing a game ")
	case "bet":
		if event.HasWon {
			builder.WriteString("sucessfully ")
		}
		builder.WriteString(fmt.Sprintf(
			"placed a bet of %d %s ",
			event.Amount,
			event.Currency,
		))
		if event.Currency != "EUR" {
			builder.WriteString(fmt.Sprintf("(%f EUR) ", event.AmountEUR))
		}
		builder.WriteString("on a game ")
	case "deposit":
		builder.WriteString(fmt.Sprintf(
			"made a deposit of %d %s ",
			event.Amount,
			event.Currency,
		))
	case "game_stop":
		builder.WriteString("stopped playing a game ")
	}

	game, ok := casino.Games[event.GameID]
	if ok && event.Type != "deposit" {
		builder.WriteString(fmt.Sprintf("\"%s\" ", game.Title))
	}

	formattedTime := event.CreatedAt.Format("on January 2nd, 2006 at 15:04 MST.")
	builder.WriteString(formattedTime)

	event.Description = builder.String()
}
