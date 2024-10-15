package api

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type SavedItem struct {
	bun.BaseModel `bun:"table:saved_items,alias:si"`

	ID      uuid.UUID `bun:"id,pk"`
	UserID  uuid.UUID `bun:"user_id,notnull"`
	VenueID uuid.UUID `bun:"venue_id"`
	EventID uuid.UUID `bun:"event_id"`
}

type SubscribedItem struct {
	bun.BaseModel `bun:"table:subscribed_items,alias:sui"`

	ID      uuid.UUID `bun:"id,pk"`
	UserID  uuid.UUID `bun:"user_id,notnull"`
	VenueID uuid.UUID `bun:"venue_id"`
	EventID uuid.UUID `bun:"event_id"`
}

type VisitedItem struct {
	bun.BaseModel `bun:"table:visited_items,alias:vi"`

	ID           uuid.UUID `bun:"id,pk"`
	UserID       uuid.UUID `bun:"user_id,notnull"`
	VenueID      uuid.UUID `bun:"venue_id"`
	EventID      uuid.UUID `bun:"event_id"`
	VisitCount   int       `bun:"visit_count,nullzero"`
	FirstVisited time.Time `bun:"first_visited,nullzero,default:current_timestamp"`
	LastVisited  time.Time `bun:"last_visited,nullzero,default:current_timestamp"`
}

type SharedItem struct {
	bun.BaseModel `bun:"table:shared_items,alias:shi"`

	FromUserID uuid.UUID `bun:"from_user_id,pk"`
	ToUserID   uuid.UUID `bun:"to_user_id,pk"`
	VenueID    uuid.UUID `bun:"venue_id"`
	EventID    uuid.UUID `bun:"event_id"`
	TimeShared time.Time `bun:"time_shared,nullzero,default:current_timestamp"`
}
