package playerdata

import (
	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"gorm.io/gorm"
)

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
