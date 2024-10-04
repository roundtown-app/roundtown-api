package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/roundtown-app/roundtown-api/api"
)

// SaveItemRequest represents the request body for saving an item
type SaveItemRequest struct {
	VenueID *uuid.UUID `json:"venue_id,omitempty"`
	EventID *uuid.UUID `json:"event_id,omitempty"`
}

// SavedItemResponse represents the response for saved items
type SavedItemResponse struct {
	Venue   *api.Venue    `json:"venue,omitempty"`
	Event   *api.Event    `json:"event,omitempty"`
}

// handleGetSavedItems retrieves all saved items for the authenticated user
func (h *Handlers) handleGetSavedItems(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uuid.UUID)

	var savedItems []api.SavedItem
	err := h.DB.NewSelect().
		Model(&savedItems).
		Where("user_id = ?", userID).
		Relation("Event").
		Relation("Venue").
		Scan(r.Context())

	if err != nil {
		http.Error(w, "Failed to fetch saved items", http.StatusInternalServerError)
		return
	}

	response := make([]SavedItemResponse, len(savedItems))
	for _, savedItem := range savedItems {
		var eventItem api.Event
		var venueItem api.Venue
		if savedItem.EventID != uuid.Nil {
			err = h.DB.NewSelect().
				Model(&eventItem).
				Where("id = ?", savedItem.EventID).
				Scan(r.Context())
			if err != nil {
				slog.Error("failed to fetch saved event: " + err.Error())
			}
			response = append(response, SavedItemResponse{
				Event: &eventItem,
			})
		} else {
			err = h.DB.NewSelect().
				Model(&venueItem).
				Where("id = ?", savedItem.VenueID).
				Scan(r.Context())
			if err != nil {
				slog.Error("failed to fetch saved venue: " + err.Error())
			}
			response = append(response, SavedItemResponse{
				Venue: &venueItem,
			})
		}
	}

	json.NewEncoder(w).Encode(response)
}

// handleSaveItem saves a new item (venue or event) for the authenticated user
func (h *Handlers) handleSaveItem(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uuid.UUID)

	var req SaveItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.VenueID == nil && req.EventID == nil {
		http.Error(w, "Either venue_id or event_id must be provided", http.StatusBadRequest)
		return
	}

	// Check if item is already saved
	exists, err := h.DB.NewSelect().
		Model((*api.SavedItem)(nil)).
		Where("user_id = ?", userID).
		Where("venue_id = ? OR event_id = ?", req.VenueID, req.EventID).
		Exists(r.Context())

	if err != nil {
		http.Error(w, "Failed to check if item exists", http.StatusInternalServerError)
		return
	}

	if exists {
		http.Error(w, "Item already saved", http.StatusConflict)
		return
	}

	// Create new saved item
	savedItem := &api.SavedItem{
		UserID:  userID,
		VenueID: *req.VenueID,
		EventID: *req.EventID,
	}

	_, err = h.DB.NewInsert().
		Model(savedItem).
		Exec(r.Context())

	if err != nil {
		http.Error(w, "Failed to save item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(savedItem)
}

// handleUnsaveItem removes a saved item for the authenticated user
func (h *Handlers) handleUnsaveItem(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uuid.UUID)

	var req SaveItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.VenueID == nil && req.EventID == nil {
		http.Error(w, "Either venue_id or event_id must be provided", http.StatusBadRequest)
		return
	}

	result, err := h.DB.NewDelete().
		Model((*api.SavedItem)(nil)).
		Where("user_id = ?", userID).
		Where("venue_id = ? OR event_id = ?", req.VenueID, req.EventID).
		Exec(r.Context())

	if err != nil {
		http.Error(w, "Failed to unsave item", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to get rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}