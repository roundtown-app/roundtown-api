package handlers

import (
	"net/http"
	"time"

	"firebase.google.com/go/auth"
	"github.com/roundtown-app/roundtown-api/db"
)

type Handlers struct {
	DB                      *db.DB
	RecommendationServerURL string
	AuthClient              *auth.Client
}

func NewHandler(db *db.DB, recServerURL string, client *auth.Client) *Handlers {
	return &Handlers{
		DB:                      db,
		RecommendationServerURL: recServerURL,
		AuthClient:              client,
	}
}

func (h *Handlers) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	// Get current time
	currentTime := time.Now()

	// Format the response
	response := currentTime.Format(time.RFC3339)

	// Set the response header and write the response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}
