package playerdata_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/enricher/playerdata"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestAddPlayerData(t *testing.T) {
	testCases := []struct {
		name     string
		event    casino.Event
		expected *casino.Player
	}{
		{
			"empty",
			casino.Event{},
			nil,
		},
		{
			"not found",
			casino.Event{
				ID:       1,
				PlayerID: 1,
				GameID:   1,
				Type:     "test type",
				Amount:   1,
				Currency: "USD",
			},
			nil,
		},
		{
			"found",
			casino.Event{
				ID:       10,
				PlayerID: 10,
				GameID:   100,
				Type:     "bet",
				Amount:   1,
				Currency: "USD",
				HasWon:   true,
			},
			&casino.Player{
				ID:             10,
				Email:          "1@1.com",
				LastSignedInAt: time.Date(2023, time.April, 15, 12, 30, 0, 0, time.UTC),
			},
		},
	}

	mockDb := PlayerDataMock{}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf(tc.name), func(t *testing.T) {
			ctx := context.TODO()
			playerdata.AddPlayerData(ctx, mockDb, &tc.event)
			if tc.expected == nil {
				assert.Nil(t, tc.event.Player)
			} else {
				if assert.NotNil(t, tc.event.Player) {
					assert.Equal(t, *tc.expected, *tc.event.Player)
				}
			}
		})
	}
}

type PlayerDataMock struct{}

// implements PlayerData
var _ playerdata.PlayerData = (*PlayerDataMock)(nil)

func (p PlayerDataMock) Find(playerId int) (casino.Player, error) {
	if playerId > 9 && playerId < 15 {
		return casino.Player{
			ID:             playerId,
			Email:          "1@1.com",
			LastSignedInAt: time.Date(2023, time.April, 15, 12, 30, 0, 0, time.UTC),
		}, nil
	}
	return casino.Player{}, gorm.ErrRecordNotFound
}
