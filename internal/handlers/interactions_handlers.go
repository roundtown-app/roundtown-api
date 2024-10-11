package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/roundtown-app/roundtown-api/api"
	"github.com/roundtown-app/roundtown-api/internal/middleware"
)

// SaveItemRequest represents the request body for saving an item
type SaveItemRequest struct {
	VenueID *uuid.UUID `json:"venue_id,omitempty"`
	EventID *uuid.UUID `json:"event_id,omitempty"`
}

// SavedItemResponse represents the response for saved items
type SavedItemResponse struct {
	Venue *api.Venue `json:"venue,omitempty"`
	Event *api.Event `json:"event,omitempty"`
}

// handleGetSavedItems retrieves all saved items for the authenticated user
func (h *Handlers) handleGetSavedItems(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

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
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

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
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

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

// SubscribedItemRequest represents the request body for subscribing to an item
type SubscribedItemRequest struct {
	VenueID *uuid.UUID `json:"venue_id,omitempty"`
	EventID *uuid.UUID `json:"event_id,omitempty"`
}

// SubscribedItemResponse represents the response for subscribed items
type SubscribedItemResponse struct {
	Venue *api.Venue `json:"venue,omitempty"`
	Event *api.Event `json:"event,omitempty"`
}

// handleGetSubscribedItems retrieves all subscribed items for the authenticated user
func (h *Handlers) handleGetSubscribedItems(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var subscribedItems []api.SubscribedItem
	err := h.DB.NewSelect().
		Model(&subscribedItems).
		Where("user_id = ?", userID).
		Relation("Event").
		Relation("Venue").
		Scan(r.Context())

	if err != nil {
		http.Error(w, "Failed to fetch subscribed items", http.StatusInternalServerError)
		return
	}

	response := make([]SubscribedItemResponse, len(subscribedItems))
	for _, item := range subscribedItems {
		var eventItem api.Event
		var venueItem api.Venue
		if item.EventID != uuid.Nil {
			err = h.DB.NewSelect().
				Model(&eventItem).
				Where("id = ?", item.EventID).
				Scan(r.Context())
			if err != nil {
				slog.Error("failed to fetch subscribed event: " + err.Error())
			}
			response = append(response, SubscribedItemResponse{
				Event: &eventItem,
			})
		} else {
			err = h.DB.NewSelect().
				Model(&venueItem).
				Where("id = ?", item.VenueID).
				Scan(r.Context())
			if err != nil {
				slog.Error("failed to fetch subscribed venue: " + err.Error())
			}
			response = append(response, SubscribedItemResponse{
				Venue: &venueItem,
			})
		}
	}

	json.NewEncoder(w).Encode(response)
}

// handleSubscribeItem subscribes to a new item (venue or event) for the authenticated user
func (h *Handlers) handleSubscribeItem(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req SubscribedItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.VenueID == nil && req.EventID == nil {
		http.Error(w, "Either venue_id or event_id must be provided", http.StatusBadRequest)
		return
	}

	// Check if already subscribed
	exists, err := h.DB.NewSelect().
		Model((*api.SubscribedItem)(nil)).
		Where("user_id = ?", userID).
		Where("venue_id = ? OR event_id = ?", req.VenueID, req.EventID).
		Exists(r.Context())

	if err != nil {
		http.Error(w, "Failed to check if subscription exists", http.StatusInternalServerError)
		return
	}

	if exists {
		http.Error(w, "Already subscribed to this item", http.StatusConflict)
		return
	}

	subscribedItem := &api.SubscribedItem{
		UserID:  userID,
		VenueID: *req.VenueID,
		EventID: *req.EventID,
	}

	_, err = h.DB.NewInsert().
		Model(subscribedItem).
		Exec(r.Context())

	if err != nil {
		http.Error(w, "Failed to subscribe to item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(subscribedItem)
}

// handleUnsubscribeItem removes a subscription for the authenticated user
func (h *Handlers) handleUnsubscribeItem(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req SubscribedItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.VenueID == nil && req.EventID == nil {
		http.Error(w, "Either venue_id or event_id must be provided", http.StatusBadRequest)
		return
	}

	result, err := h.DB.NewDelete().
		Model((*api.SubscribedItem)(nil)).
		Where("user_id = ?", userID).
		Where("venue_id = ? OR event_id = ?", req.VenueID, req.EventID).
		Exec(r.Context())

	if err != nil {
		http.Error(w, "Failed to unsubscribe from item", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to get rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// VisitedItemResponse represents the response for visited items
type VisitedItemResponse struct {
	Venue        *api.Venue `json:"venue,omitempty"`
	Event        *api.Event `json:"event,omitempty"`
	VisitCount   int        `json:"visit_count"`
	FirstVisited time.Time  `json:"first_visited"`
	LastVisited  time.Time  `json:"last_visited"`
}

// handleGetVisitedItems retrieves all visited items for the authenticated user
func (h *Handlers) handleGetVisitedItems(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var visitedItems []api.VisitedItem
	err := h.DB.NewSelect().
		Model(&visitedItems).
		Where("user_id = ?", userID).
		Relation("Event").
		Relation("Venue").
		Scan(r.Context())

	if err != nil {
		http.Error(w, "Failed to fetch visited items", http.StatusInternalServerError)
		return
	}

	response := make([]VisitedItemResponse, len(visitedItems))
	for _, item := range visitedItems {
		var eventItem api.Event
		var venueItem api.Venue
		if item.EventID != uuid.Nil {
			err = h.DB.NewSelect().
				Model(&eventItem).
				Where("id = ?", item.EventID).
				Scan(r.Context())
			if err != nil {
				slog.Error("failed to fetch visited event: " + err.Error())
			}
			response = append(response, VisitedItemResponse{
				Event:        &eventItem,
				VisitCount:   item.VisitCount,
				FirstVisited: item.FirstVisited,
				LastVisited:  item.LastVisited,
			})
		} else {
			err = h.DB.NewSelect().
				Model(&venueItem).
				Where("id = ?", item.VenueID).
				Scan(r.Context())
			if err != nil {
				slog.Error("failed to fetch visited venue: " + err.Error())
			}
			response = append(response, VisitedItemResponse{
				Venue:        &venueItem,
				VisitCount:   item.VisitCount,
				FirstVisited: item.FirstVisited,
				LastVisited:  item.LastVisited,
			})
		}
	}

	json.NewEncoder(w).Encode(response)
}

// handleVisitItem records a visit to an item for the authenticated user
func (h *Handlers) handleVisitItem(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req SubscribedItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.VenueID == nil && req.EventID == nil {
		http.Error(w, "Either venue_id or event_id must be provided", http.StatusBadRequest)
		return
	}

	now := time.Now()

	// Check if item has been visited before
	var existingVisit api.VisitedItem
	err := h.DB.NewSelect().
		Model(&existingVisit).
		Where("user_id = ?", userID).
		Where("venue_id = ? OR event_id = ?", req.VenueID, req.EventID).
		Scan(r.Context())

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Failed to check visit history", http.StatusInternalServerError)
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		// First visit
		visitedItem := &api.VisitedItem{
			UserID:       userID,
			VenueID:      *req.VenueID,
			EventID:      *req.EventID,
			VisitCount:   1,
			FirstVisited: now,
			LastVisited:  now,
		}

		_, err = h.DB.NewInsert().
			Model(visitedItem).
			Exec(r.Context())
	} else {
		// Update existing visit
		_, err = h.DB.NewUpdate().
			Model(&existingVisit).
			Set("visit_count = visit_count + 1").
			Set("last_visited = ?", now).
			Where("user_id = ?", userID).
			Where("venue_id = ? OR event_id = ?", req.VenueID, req.EventID).
			Exec(r.Context())
	}

	if err != nil {
		http.Error(w, "Failed to record visit", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SharedItemResponse represents the response for shared items
type SharedItemResponse struct {
	FromUser   uuid.UUID  `json:"from_user"`
	ToUser     uuid.UUID  `json:"to_user"`
	Venue      *api.Venue `json:"venue,omitempty"`
	Event      *api.Event `json:"event,omitempty"`
	TimeShared time.Time  `json:"time_shared"`
}

// handleGetSharedItems retrieves all items shared by or with the authenticated user
func (h *Handlers) handleGetSharedItems(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var sharedItems []api.SharedItem
	err := h.DB.NewSelect().
		Model(&sharedItems).
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Relation("Event").
		Relation("Venue").
		Scan(r.Context())

	if err != nil {
		http.Error(w, "Failed to fetch shared items", http.StatusInternalServerError)
		return
	}

	response := make([]SharedItemResponse, len(sharedItems))
	for _, item := range sharedItems {
		var eventItem api.Event
		var venueItem api.Venue
		if item.EventID != uuid.Nil {
			err = h.DB.NewSelect().
				Model(&eventItem).
				Where("id = ?", item.EventID).
				Scan(r.Context())
			if err != nil {
				slog.Error("failed to fetch shared event: " + err.Error())
			}
			response = append(response, SharedItemResponse{
				FromUser:   item.FromUserID,
				ToUser:     item.ToUserID,
				Event:      &eventItem,
				TimeShared: item.TimeShared,
			})
		} else {
			err = h.DB.NewSelect().
				Model(&venueItem).
				Where("id = ?", item.VenueID).
				Scan(r.Context())
			if err != nil {
				slog.Error("failed to fetch shared venue: " + err.Error())
			}
			response = append(response, SharedItemResponse{
				FromUser:   item.FromUserID,
				ToUser:     item.ToUserID,
				Venue:      &venueItem,
				TimeShared: item.TimeShared,
			})
		}
	}

	json.NewEncoder(w).Encode(response)
}

// ShareItemRequest represents the request body for sharing an item
type ShareItemRequest struct {
	ToUserID uuid.UUID  `json:"to_user_id"`
	VenueID  *uuid.UUID `json:"venue_id,omitempty"`
	EventID  *uuid.UUID `json:"event_id,omitempty"`
}

// handleShareItem shares an item with another user
func (h *Handlers) handleShareItem(w http.ResponseWriter, r *http.Request) {
	fromUserID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if fromUserID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req ShareItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.VenueID == nil && req.EventID == nil {
		http.Error(w, "Either venue_id or event_id must be provided", http.StatusBadRequest)
		return
	}

	if req.ToUserID == uuid.Nil {
		http.Error(w, "to_user_id is required", http.StatusBadRequest)
		return
	}

	// Check if item is already shared
	var existingShare api.SharedItem
	err := h.DB.NewSelect().
		Model(&existingShare).
		Where("from_user_id = ?", fromUserID).
		Where("to_user_id = ?", req.ToUserID).
		Where("venue_id = ? OR event_id = ?", req.VenueID, req.EventID).
		Scan(r.Context())

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Failed to check visit history", http.StatusInternalServerError)
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		// First share
		sharedItem := &api.SharedItem{
			FromUserID: fromUserID,
			ToUserID:   req.ToUserID,
			VenueID:    *req.VenueID,
			EventID:    *req.EventID,
			TimeShared: time.Now(),
		}

		_, err = h.DB.NewInsert().
			Model(sharedItem).
			Exec(r.Context())
	} else {
		// Update existing share
		_, err = h.DB.NewUpdate().
			Model(&existingShare).
			Set("last_visited = ?", existingShare.TimeShared).
			Exec(r.Context())
	}

	if err != nil {
		http.Error(w, "Failed to record visit", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
