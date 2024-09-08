package tools

import (
	"time"
)

type mock_database struct{}

var mockLoginDetails = map[string]LoginDetails{
	"123ABC": {
		AuthToken: "123ABC",
		Username:  "alex",
	},
	"456DEF": {
		AuthToken: "456DEF",
		Username:  "jason",
	},
	"789GHI": {
		AuthToken: "789GHI",
		Username:  "marie",
	},
}

var mockEventDetails = map[string]EventDetails{
	"1": {
		Name: "Happy Hour",
		ID:   "1",
	},
	"2": {
		Name: "Dance",
		ID:   "2",
	},
}

func (d *mock_database) GetUserLoginDetails(token string) *LoginDetails {
	// Simulate DB call
	time.Sleep(time.Second * 1)

	var clientData = LoginDetails{}
	clientData, ok := mockLoginDetails[token]
	if !ok {
		return nil
	}

	return &clientData
}

func (d *mock_database) GetEvents(id string) *EventDetails {
	// Simulate DB call
	time.Sleep(time.Second * 1)

	var clientData = EventDetails{}
	clientData, ok := mockEventDetails[id]
	if !ok {
		return nil
	}

	return &clientData
}

func (d *mock_database) SetupDatabase() error {
	return nil
}
