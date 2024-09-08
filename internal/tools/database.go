package tools

import (
	"log/slog"
)

// Database collections
type LoginDetails struct {
	AuthToken string
	Username  string
}
type EventDetails struct {
	Name string
	ID   string
}

type DatabaseInterface interface {
	GetUserLoginDetails(token string) *LoginDetails
	GetEvents(id string) *EventDetails
	SetupDatabase() error
}

func NewDatabase() (*DatabaseInterface, error) {

	var database DatabaseInterface = &mock_database{}

	var err error = database.SetupDatabase()
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return &database, nil
}
