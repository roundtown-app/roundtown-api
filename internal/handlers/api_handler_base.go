package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"firebase.google.com/go/auth"
	"github.com/google/uuid"
	"github.com/roundtown-app/roundtown-api/api"
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

// handleQRCodeScan records a visit to an item for the given user and event
func (h *Handlers) handleQRCodeScan(w http.ResponseWriter, r *http.Request) {
	// Parse the URL
	u, err := url.Parse(r.URL.String())
	if err != nil {
		slog.Error("handleQRCodeScan - failed to parse URL: " + err.Error())
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Extract query parameters
	params, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		slog.Error("handleQRCodeScan - failed to parse query: " + err.Error())
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	// Retrieve specific parameters
	userID, err := uuid.Parse(params.Get("user"))
	if err != nil {
		slog.Error("handleQRCodeScan - failed to parse user ID: " + err.Error())
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	eventID, err := uuid.Parse(params.Get("event"))
	if err != nil {
		slog.Error("handleQRCodeScan - failed to parse event ID: " + err.Error())
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	now := time.Now()

	new_visited_item_id, err := uuid.NewV7()
	if err != nil {
		slog.Error("handleQRCodeScan - failed to generate new UUID: " + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	visitedItem := &api.VisitedItem{
		ID:          new_visited_item_id,
		UserID:      userID,
		VenueID:     uuid.Nil,
		EventID:     eventID,
		VisitedTime: now,
		QRCode:      true,
	}

	_, err = h.DB.NewInsert().
		Model(visitedItem).
		Exec(r.Context())
	if err != nil {
		slog.Error("handleQRCodeScan - failed to record scan: " + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Roundtown</title>
	</head>
	<body>
		<h1>Success</h1>
		<p>The QR code scan has been registered.</p>
	</body>
	</html>
	`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}
