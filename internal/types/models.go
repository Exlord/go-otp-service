package types

import (
	"time"
	"github.com/google/uuid"
)

type User struct {
	ID uuid.UUID `json:"id"`
	Phone string `json:"phone"`
	RegisteredAt time.Time `json:"registeredAt"`
}
