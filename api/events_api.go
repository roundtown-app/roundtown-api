package api

import (
	"time"

	"github.com/uptrace/bun"
)

type Event struct {
	bun.BaseModel `bun:"table:events,alias:e"`

	ID           string    `bun:"id,pk"`
	Image        string    `bun:"image"`
	Title        string    `bun:"title"`
	Description  string    `bun:"description"`
	Detailed     string    `bun:"detailed"`
	Longitude    float64   `bun:"longitude,type:decimal(11,8)"`
	Latitude     float64   `bun:"latitude,type:decimal(11,8)"`
	Trending     int      `bun:"trending"`
	Price        int       `bun:"price"`
	Sponsor      int       `bun:"sponsor"`
	IsDeal       bool      `bun:"is_deal"`
	StartingTime time.Time `bun:"starting_time"`
	EndingTime   time.Time `bun:"ending_time"`
	VenueID      string    `bun:"venue_id"`
	Venue        *Venue    `bun:"rel:belongs-to,join:venue_id=id"`
}

type SavedItem struct {
	bun.BaseModel `bun:"table:saved_items,alias:si"`

	UserID   string `bun:"user_id,pk"`
	ItemID   string `bun:"item_id,pk"`
	ItemType string `bun:"item_type,notnull"` // "event" or "venue"
	Time time.Time `bun:"time"`

	User  *User  `bun:"rel:belongs-to,join:user_id=user_id"`
	Event *Event `bun:"rel:belongs-to,join:item_id=id"`
	Venue *Venue `bun:"rel:belongs-to,join:item_id=id"`
}

type VisitedItem struct {
	bun.BaseModel `bun:"table:visited_items,alias:vi"`

	UserID   string `bun:"user_id,pk"`
	ItemID   string `bun:"item_id,pk"`
	ItemType string `bun:"item_type,notnull"` // "event" or "venue"
	Time time.Time `bun:"time"`

	User  *User  `bun:"rel:belongs-to,join:user_id=user_id"`
	Event *Event `bun:"rel:belongs-to,join:item_id=id"`
	Venue *Venue `bun:"rel:belongs-to,join:item_id=id"`
}

type SharedItem struct {
	bun.BaseModel `bun:"table:shared_items,alias:shi"`

	FromUserID   string `bun:"user_id,pk"`
	ToUserID   string `bun:"user_id,pk"`
	ItemID   string `bun:"item_id,pk"`
	ItemType string `bun:"item_type,notnull"` // "event" or "venue"
	Time time.Time `bun:"time"`

	FromUser  *User  `bun:"rel:belongs-to,join:user_id=user_id"`
	ToUser  *User  `bun:"rel:belongs-to,join:user_id=user_id"`
	Event *Event `bun:"rel:belongs-to,join:item_id=id"`
	Venue *Venue `bun:"rel:belongs-to,join:item_id=id"`
}