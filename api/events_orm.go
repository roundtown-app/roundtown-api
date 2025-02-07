package api

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Event struct {
	bun.BaseModel `bun:"table:events,alias:e"`

	ID            uuid.UUID `bun:"id,pk,type:uuid"`
	Title         string    `bun:"title,notnull"`
	Tag           string    `bun:"tag"`
	Description   string    `bun:"description"`
	Detailed      string    `bun:"detailed"`
	LocationID    uuid.UUID `bun:"location_id,notnull,unique"`
	ActivityCount int       `bun:"activity_count,nullzero"`
	Price         int       `bun:"price,nullzero"`
	Sponsor       int       `bun:"sponsor,nullzero"`
	IsDeal        bool      `bun:"is_deal,nullzero"`
	StartingTime  time.Time `bun:"starting_time,notnull"`
	EndingTime    time.Time `bun:"ending_time,notnull"`
	VenueID       uuid.UUID `bun:"venue_id,notnull"`
	CreatedAt     time.Time `bun:"created_at,nullzero,default:current_timestamp"`
	OwnerID       uuid.UUID `bun:"owner_id"`
	QRCode        string    `bun:"qr_code"`
}

type EventCategory struct {
	bun.BaseModel `bun:"table:event_categories,alias:ec"`

	EventID  uuid.UUID `bun:"event_id,pk"`
	Category string    `bun:"category,pk"`
}

type EventLocation struct {
	bun.BaseModel `bun:"table:event_locations,alias:el"`

	ID        uuid.UUID `bun:"id,pk"`
	Longitude float64   `bun:"longitude"`
	Latitude  float64   `bun:"latitude"`
	Address   string    `bun:"address"`
}

type EventRating struct {
	bun.BaseModel `bun:"table:event_ratings,alias:er"`

	EventID uuid.UUID `bun:"event_id,pk"`
	UserID  uuid.UUID `bun:"user_id,pk"`
	Rating  int       `bun:"rating"`
}

type EventAssets struct {
	bun.BaseModel `bun:"table:event_assets,alias:ea"`

	EventID uuid.UUID `bun:"event_id,pk"`
	AssetID string    `bun:"asset_id,pk"`
}

type EventPopulation struct {
	bun.BaseModel `bun:"table:event_population,alias:ep"`

	EventID     uuid.UUID `bun:"event_id,pk"`
	UserCount   int       `bun:"user_count"`
	LastUpdated time.Time `bun:"last_updated,nullzero,default:current_timestamp"`
}

type EventRecurrencePattern struct {
	bun.BaseModel `bun:"table:event_recurrence_patterns,alias:erp"`

	EventID     string    `bun:"event_id,pk,type:uuid"`
	Frequency   string    `bun:"frequency,notnull"`
	DaysOfWeek  []int     `bun:"days_of_week,array"`
	WeekOfMonth []int     `bun:"week_of_month,array"`
	StartDate   time.Time `bun:"start_date,notnull"`
	EndDate     time.Time `bun:"end_date"`
}

type EventException struct {
	bun.BaseModel `bun:"table:event_exceptions,alias:ee"`

	EventID               string    `bun:"event_id,pk,type:uuid"`
	ExceptionDate         time.Time `bun:"exception_date,pk"`
	IsCancelled           bool      `bun:"is_cancelled"`
	AlternateStartingTime time.Time `bun:"alternate_starting_time"`
	AlternateEndingTime   time.Time `bun:"alternate_ending_time"`
}
