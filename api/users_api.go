package api

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	UserID          string    `bun:"user_id,pk"`
	Username        string    `bun:"username,unique"`
	AccountCreated  time.Time `bun:"account_created,nullzero,notnull,default:current_timestamp"`
	AccountType     string    `bun:"account_type"`
}

type Friend struct {
	bun.BaseModel `bun:"table:friends,alias:f"`

	UserID1 string `bun:"user_id1,pk"`
	UserID2 string `bun:"user_id2,pk"`

	User1 *User `bun:"rel:belongs-to,join:user_id1=user_id"`
	User2 *User `bun:"rel:belongs-to,join:user_id2=user_id"`
}