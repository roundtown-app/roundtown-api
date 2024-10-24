package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/roundtown-app/roundtown-api/internal/middleware"
)

// Response represents the final response structure
type Response struct {
	Events []*EventDetails `json:"events"`
	Count  int             `json:"count"`
}

func (h *Handlers) forwardRequest(_ http.ResponseWriter, r *http.Request, url string) ([]string, error) {
	// Create new POST request with the original request's body
	req, err := http.NewRequestWithContext(r.Context(), "POST", url, r.Body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Copy relevant headers from original request
	req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("recommendation server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Decode the response as a string array
	var eventIDs []string
	if err := json.Unmarshal(body, &eventIDs); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return eventIDs, nil
}

func (h *Handlers) handleGetFeedRec(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Check if it's a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	url := h.RecommendationServerURL + "/feed-recs/" + userID.String()

	// Get recommendations
	eventIDs, err := h.forwardRequest(w, r, url)
	if err != nil {
		http.Error(w, "Error fetching recommendations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get event details for each recommended event
	var eventDetails []*EventDetails
	for _, eventIDStr := range eventIDs {
		eventID, err := uuid.Parse(eventIDStr)
		if err != nil {
			// Log the error but continue with other events
			continue
		}

		details, err := h.getEventDetails(ctx, eventID, userID)
		if err != nil {
			// Log the error but continue with other events
			continue
		}
		eventDetails = append(eventDetails, details)
	}

	// Prepare the response
	response := Response{
		Events: eventDetails,
		Count:  len(eventDetails),
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// func (h *Handlers) handleGetPlanRec(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)
// 	if userID == uuid.Nil {
// 		http.Error(w, "User not authenticated", http.StatusUnauthorized)
// 		return
// 	}

// 	url := h.RecommendationServerURL + "/plan-recs/" + userID.String()

// 	h.forwardRequest(w, url)
// }

// func (h *Handlers) handleGetEBRec(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)
// 	if userID == uuid.Nil {
// 		http.Error(w, "User not authenticated", http.StatusUnauthorized)
// 		return
// 	}

// 	url := h.RecommendationServerURL + "/event-based-recs/" + userID.String()

// 	h.forwardRequest(w, url)
// }
