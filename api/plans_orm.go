package api

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Plan struct {
	bun.BaseModel `bun:"table:plans,alias:p"`

	ID        uuid.UUID `bun:"id,pk,type:uuid"`
	Title     string    `bun:"title"`
	Date      time.Time `bun:"date"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	Private   bool      `bun:"private,nullzero"`
	CreatorID uuid.UUID `bun:"creator_id,notnull"`
}

type PlanItem struct {
	bun.BaseModel `bun:"table:plan_items,alias:pi"`

	ID      int       `bun:"id,pk,autoincrement"`
	PlanID  uuid.UUID `bun:"plan_id"`
	VenueID uuid.UUID `bun:"venue_id"`
	EventID uuid.UUID `bun:"event_id"`
}

type PlanUser struct {
	bun.BaseModel `bun:"table:plan_users,alias:pu"`

	PlanID uuid.UUID `bun:"plan_id,pk"`
	UserID uuid.UUID `bun:"user_id,pk"`
}
