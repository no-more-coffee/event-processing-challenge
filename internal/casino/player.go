package casino

import "time"

type Player struct {
	ID             int       `json:"-"`
	Email          string    `json:"email"`
	LastSignedInAt time.Time `json:"last_signed_in_at"`
}

func (p Player) IsZero() bool {
	return p.Email == "" || p.LastSignedInAt.IsZero()
}
