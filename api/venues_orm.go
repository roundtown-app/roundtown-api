package api

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Venue struct {
	bun.BaseModel `bun:"table:venues,alias:v"`

	ID            uuid.UUID `bun:"id,pk,type:uuid"`
	Title         string    `bun:"title,notnull"`
	Description   string    `bun:"description"`
	Detailed      string    `bun:"detailed"`
	LocationID    int       `bun:"location_id,notnull,unique"`
	ActivityCount int       `bun:"activity_count,nullzero"`
	Price         int       `bun:"price,nullzero"`
	Sponsor       int       `bun:"sponsor,nullzero"`
	Hidden        bool      `bun:"hidden,nullzero"`
	CreatedAt     time.Time `bun:"created_at,nullzero,default:current_timestamp"`
	OwnerID       uuid.UUID `bun:"owner_id"`
}

type VenueCategory struct {
	bun.BaseModel `bun:"table:venue_categories,alias:vc"`

	VenueID  uuid.UUID `bun:"venue_id,pk"`
	Category string    `bun:"category,pk"`
}

type VenueLocation struct {
	bun.BaseModel `bun:"table:venue_locations,alias:vl"`

	ID        int     `bun:"id,pk,autoincrement"`
	Longitude float64 `bun:"longitude"`
	Latitude  float64 `bun:"latitude"`
	Address   string  `bun:"address"`
}

type VenueRating struct {
	bun.BaseModel `bun:"table:venue_ratings,alias:vr"`

	VenueID uuid.UUID `bun:"venue_id,pk"`
	UserID  uuid.UUID `bun:"user_id,pk"`
	Rating  int       `bun:"rating"`
}

type VenueAssets struct {
	bun.BaseModel `bun:"table:venue_assets,alias:va"`

	VenueID uuid.UUID `bun:"venue_id,pk"`
	AssetID string    `bun:"asset_id,pk"`
}

type VenuePopulation struct {
	bun.BaseModel `bun:"table:venue_population,alias:vp"`

	VenueID     uuid.UUID `bun:"venue_id,pk"`
	UserCount   int       `bun:"user_count"`
	LastUpdated time.Time `bun:"last_updated,nullzero,default:current_timestamp"`
}
