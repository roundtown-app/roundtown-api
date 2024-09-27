package api

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type SurgeDeal struct {
	bun.BaseModel `bun:"table:surge_deals,alias:sd"`

	EventID   uuid.UUID     `bun:"event_id,pk,type:uuid"`
	MinUsers  int           `bun:"min_users,notnull"`
	Duration  time.Duration `bun:"duration,notnull"`
}

type ComboDeal struct {
	bun.BaseModel `bun:"table:combo_deals,alias:cd"`

	ID        uuid.UUID `bun:"id,pk,type:uuid"`
	Title     string    `bun:"title,notnull"`
	StartDate time.Time `bun:"start_date,notnull"`
	EndDate   time.Time `bun:"end_date,notnull"`
}

type ComboItem struct {
	bun.BaseModel `bun:"table:combo_items,alias:ci"`

	ID      int       `bun:"id,pk,autoincrement"`
	ComboID uuid.UUID `bun:"combo_id,notnull"`
	VenueID uuid.UUID `bun:"venue_id"`
	EventID uuid.UUID `bun:"event_id"`
}