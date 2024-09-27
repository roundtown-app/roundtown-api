package api

import (
	"github.com/uptrace/bun"
)

type Venue struct {
	bun.BaseModel `bun:"table:venues,alias:v"`

	ID                string    `bun:"id,pk"`
	Image             string    `bun:"image"`
	Title             string    `bun:"title"`
	Description       string    `bun:"description"`
	Detailed          string    `bun:"detailed"`
	Longitude         float64   `bun:"longitude,type:decimal(11,8)"`
	Latitude          float64   `bun:"latitude,type:decimal(11,8)"`
	ActivityCount     int       `bun:"activity_count"`
	Price             int       `bun:"price"`
	Sponsor           int       `bun:"sponsor"`
	IsHidden          bool      `bun:"hidden"`
}
