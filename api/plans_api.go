package api

import (
	"time"

	"github.com/uptrace/bun"
)

type Plan struct {
	bun.BaseModel `bun:"table:plans,alias:p"`

	ID        	string    `bun:"id,pk"`
	Title     	string    `bun:"title"`
	Date      	time.Time `bun:"date"`
	CreatedAt 	time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	IsPrivate   bool      `bun:"private"`

	Items []*PlanItem `bun:"m2m:plan_items,join:Plan=Item"`

	CreatorID string `bun:"creator_id"`
	Creator   *User  `bun:"rel:belongs-to,join:creator_id=id"`

	AddedUsers []*User `bun:"m2m:plan_users,join:Plan=User"`
}

// Junction table for Plan-Item (Event or Venue) many-to-many relationship
type PlanItem struct {
	bun.BaseModel `bun:"table:plan_items,alias:pi"`

	PlanID string `bun:"plan_id"`
	ItemID string `bun:"item_id"`
	Type   string `bun:"type"` // This will be either "event" or "venue"
}

type PlanUser struct {
	bun.BaseModel `bun:"table:plan_users,alias:pu"`

	PlanID string `bun:"plan_id"`
	UserID string `bun:"user_id"`
}

type ComboDeal struct {
	bun.BaseModel `bun:"table:combo_deals,alias:cd"`

	ID        string    `bun:"id,pk"`
	Title     string    `bun:"title,notnull"`
	StartDate time.Time `bun:"start_date,notnull"`
	EndDate   time.Time `bun:"end_date,notnull"`
}

type ComboItem struct {
	bun.BaseModel `bun:"table:combo_items,alias:ci"`

	ComboID string `bun:"combo_id"`
	ItemID string `bun:"item_id"`
	Type   string `bun:"type"` // This will be either "event" or "venue"
}
