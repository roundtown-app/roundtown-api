package api

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	UserID        uuid.UUID `bun:"user_id,pk,type:uuid"`
	FirebaseUID   string    `bun:"firebase_uid,notnull,unique"`
	Username      string    `bun:"username,notnull,unique"`
	AccountType   string    `bun:"account_type,notnull"`
	AccountCreated time.Time `bun:"account_created,nullzero,default:current_timestamp"`
}

type UserLocations struct {
	bun.BaseModel `bun:"table:user_locations,alias:ul"`

	UserID        uuid.UUID `bun:"user_id,pk,type:uuid"`
	Longitude float64 `bun:"longitude"`
	Latitude  float64 `bun:"latitude"`
	LastUpdated time.Time `bun:"last_updated,nullzero,default:current_timestamp"`
}

type Friend struct {
	bun.BaseModel `bun:"table:friends,alias:f"`

	UserID1   uuid.UUID `bun:"user_id1,pk"`
	UserID2   uuid.UUID `bun:"user_id2,pk"`
	CreatedAt time.Time `bun:"created_at,nullzero,default:current_timestamp"`
}