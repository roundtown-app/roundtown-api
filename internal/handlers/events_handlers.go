package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/roundtown-app/roundtown-api/api"
)

// ----- GET helper struct -----

type EventDetails struct {
	Event        api.Event           `json:"event"`
	Categories   []api.EventCategory `json:"categories"`
	Location     api.EventLocation   `json:"location"`
	Ratings      []api.EventRating   `json:"ratings"`
	Assets       []api.EventAssets   `json:"assets"`
	Population   api.EventPopulation `json:"population"`
	IsSaved      bool                `json:"is_saved"`
	IsVisited    bool                `json:"is_visited"`
	IsSharedFrom bool                `json:"is_shared_from"`
	IsSharedTo   bool                `json:"is_shared_to"`
}

// ----- PUT helper structs -----

type EventUpdate struct {
	Title         *string         `json:"title"`
	Description   *string         `json:"description"`
	Detailed      *string         `json:"detailed"`
	Price         *int            `json:"price"`
	Sponsor       *int            `json:"sponsor"`
	IsDeal        *bool           `json:"is_deal"`
	StartingTime  *time.Time      `json:"starting_time"`
	EndingTime    *time.Time      `json:"ending_time"`
	Recurring     *int            `json:"recurring"`
	VenueID       *uuid.UUID      `json:"venue_id"`
	Categories    []string        `json:"categories"`
	Location      *EventLocationUpdate `json:"location"`
	Assets        []string        `json:"assets"`
}

type EventLocationUpdate struct {
	ID        int      `json:"id"`
	Longitude *float64 `json:"longitude"`
	Latitude  *float64 `json:"latitude"`
	Address   *string  `json:"address"`
}

// ----- POST helper struct -----

type EventInput struct {
	Event      api.Event         `json:"event"`
	Categories []string          `json:"categories"`
	Location   api.EventLocation `json:"location"`
	Assets     []api.EventAssets `json:"assets"`
}

// ----- GET helper functions -----

func (h *Handlers) getEventDetails(ctx context.Context, eventID uuid.UUID, userID uuid.UUID) (*EventDetails, error) {
	var details EventDetails

	// Fetch the main event
	err := h.DB.NewSelect().
		Model(&details.Event).
		Where("id = ?", eventID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event: %w", err)
	}

	// Fetch categories
	err = h.DB.NewSelect().
		Model(&details.Categories).
		Where("event_id = ?", eventID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event categories: %w", err)
	}

	// Fetch location
	err = h.DB.NewSelect().
		Model(&details.Location).
		Where("id = ?", details.Event.LocationID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event location: %w", err)
	}

	// Fetch ratings
	err = h.DB.NewSelect().
		Model(&details.Ratings).
		Where("event_id = ?", eventID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event ratings: %w", err)
	}

	// Fetch assets
	err = h.DB.NewSelect().
		Model(&details.Assets).
		Where("event_id = ?", eventID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event assets: %w", err)
	}

	// Fetch population
	err = h.DB.NewSelect().
		Model(&details.Population).
		Where("event_id = ?", eventID).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			// If no population record exists, create a default one
			details.Population = api.EventPopulation{
				EventID:     eventID,
				UserCount:   0,
				LastUpdated: time.Now(),
			}
		} else {
			return nil, fmt.Errorf("failed to fetch event population: %w", err)
		}
	}

	if userID != uuid.Nil {
		// Check if the event is saved by the user
		var savedItem api.SavedItem
		err = h.DB.NewSelect().
			Model(&savedItem).
			Where("user_id = ? AND event_id = ?", userID, eventID).
			Scan(ctx)
		details.IsSaved = err == nil

		// Check if the event is visited by the user
		var visitedItem api.VisitedItem
		err = h.DB.NewSelect().
			Model(&visitedItem).
			Where("user_id = ? AND event_id = ?", userID, eventID).
			Scan(ctx)
		details.IsVisited = err == nil

		// Check if the event is shared by the user
		var sharedItemFrom api.SharedItem
		err = h.DB.NewSelect().
			Model(&sharedItemFrom).
			Where("from_user_id = ? AND event_id = ?", userID, eventID).
			Scan(ctx)
		details.IsSharedFrom = err == nil

		// Check if the event is shared to the user
		var sharedItemTo api.SharedItem
		err = h.DB.NewSelect().
			Model(&sharedItemTo).
			Where("from_user_id = ? AND event_id = ?", userID, eventID).
			Scan(ctx)
		details.IsSharedFrom = err == nil
	}

	return &details, nil
}

// ----- PUT helper functions -----

func (h *Handlers) isEventOwner(ctx context.Context, eventID, userID uuid.UUID) (bool, error) {
	if ctx.Value("userRole").(string) == "admin" {
		return true, nil
	}

	var event api.Event
	err := h.DB.NewSelect().
		Model(&event).
		Where("id = ? AND owner_id = ?", eventID, userID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (h *Handlers) updateEvent(ctx context.Context, tx bun.Tx, eventID uuid.UUID, update *EventUpdate) error {
	// Update main event details
	_, err := tx.NewUpdate().
		Model(&api.Event{}).
		Where("id = ?", eventID).
		Set("title = COALESCE(?, title)", update.Title).
		Set("description = COALESCE(?, description)", update.Description).
		Set("detailed = COALESCE(?, detailed)", update.Detailed).
		Set("price = COALESCE(?, price)", update.Price).
		Set("sponsor = COALESCE(?, sponsor)", update.Sponsor).
		Set("is_deal = COALESCE(?, is_deal)", update.IsDeal).
		Set("starting_time = COALESCE(?, starting_time)", update.StartingTime).
		Set("ending_time = COALESCE(?, ending_time)", update.EndingTime).
		Set("recurring = COALESCE(?, recurring)", update.Recurring).
		Set("venue_id = COALESCE(?, venue_id)", update.VenueID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	// Update categories
	if update.Categories != nil {
		err = h.updateEventCategories(ctx, tx, eventID, update.Categories)
		if err != nil {
			return fmt.Errorf("failed to update event categories: %w", err)
		}
	}

	// Update location
	if update.Location != nil {
		err = h.updateEventLocation(ctx, tx, eventID, update.Location)
		if err != nil {
			return fmt.Errorf("failed to update event location: %w", err)
		}
	}

	// Update assets
	if update.Assets != nil {
		err = h.updateEventAssets(ctx, tx, eventID, update.Assets)
		if err != nil {
			return fmt.Errorf("failed to update event assets: %w", err)
		}
	}

	return nil
}

func (h *Handlers) updateEventLocation(ctx context.Context, tx bun.Tx, eventID uuid.UUID, location *EventLocationUpdate) error {
	_, err := tx.NewUpdate().
		Model(&api.EventLocation{}).
		Where("id = ?", location.ID).
		Set("longitude = COALESCE(?, longitude)", location.Longitude).
		Set("latitude = COALESCE(?, latitude)", location.Latitude).
		Set("address = COALESCE(?, address)", location.Address).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Update the location_id in the main event table
	_, err = tx.NewUpdate().
		Model(&api.Event{}).
		Where("id = ?", eventID).
		Set("location_id = ?", location.ID).
		Exec(ctx)

	return err
}

func (h *Handlers) updateEventCategories(ctx context.Context, tx bun.Tx, eventID uuid.UUID, newCategories []string) error {
	// Fetch existing categories
	var existingCategories []api.EventCategory
	err := tx.NewSelect().
		Model(&existingCategories).
		Where("event_id = ?", eventID).
		Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch existing categories: %w", err)
	}

	// Create maps for easier comparison
	existingMap := make(map[string]bool)
	for _, cat := range existingCategories {
		existingMap[cat.Category] = true
	}
	newMap := make(map[string]bool)
	for _, cat := range newCategories {
		newMap[cat] = true
	}

	// Find categories to add
	var categoriesToAdd []api.EventCategory
	for _, cat := range newCategories {
		if !existingMap[cat] {
			categoriesToAdd = append(categoriesToAdd, api.EventCategory{EventID: eventID, Category: cat})
		}
	}

	// Find categories to remove
	var categoriesToRemove []string
	for _, cat := range existingCategories {
		if !newMap[cat.Category] {
			categoriesToRemove = append(categoriesToRemove, cat.Category)
		}
	}

	// Add new categories
	if len(categoriesToAdd) > 0 {
		_, err := tx.NewInsert().
			Model(&categoriesToAdd).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to insert new categories: %w", err)
		}
	}

	// Remove old categories
	if len(categoriesToRemove) > 0 {
		_, err := tx.NewDelete().
			Model((*api.EventCategory)(nil)).
			Where("event_id = ? AND category IN (?)", eventID, bun.In(categoriesToRemove)).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete old categories: %w", err)
		}
	}

	return nil
}

func (h *Handlers) updateEventAssets(ctx context.Context, tx bun.Tx, eventID uuid.UUID, newAssets []string) error {
	// Fetch existing assets
	var existingAssets []api.EventAssets
	err := tx.NewSelect().
		Model(&existingAssets).
		Where("event_id = ?", eventID).
		Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch existing assets: %w", err)
	}

	// Create maps for easier comparison
	existingMap := make(map[string]bool)
	for _, asset := range existingAssets {
		existingMap[asset.AssetID] = true
	}
	newMap := make(map[string]bool)
	for _, assetID := range newAssets {
		newMap[assetID] = true
	}

	// Find assets to add
	var assetsToAdd []api.EventAssets
	for _, assetID := range newAssets {
		if !existingMap[assetID] {
			assetsToAdd = append(assetsToAdd, api.EventAssets{EventID: eventID, AssetID: assetID})
		}
	}

	// Find assets to remove
	var assetsToRemove []string
	for _, asset := range existingAssets {
		if !newMap[asset.AssetID] {
			assetsToRemove = append(assetsToRemove, asset.AssetID)
		}
	}

	// Add new assets
	if len(assetsToAdd) > 0 {
		_, err := tx.NewInsert().
			Model(&assetsToAdd).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to insert new assets: %w", err)
		}
	}

	// Remove old assets
	if len(assetsToRemove) > 0 {
		_, err := tx.NewDelete().
			Model((*api.EventAssets)(nil)).
			Where("event_id = ? AND asset_id IN (?)", eventID, bun.In(assetsToRemove)).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete old assets: %w", err)
		}
	}

	return nil
}

// ----- DELETE helper function -----

func (h *Handlers) DeleteEvent(ctx context.Context, tx bun.Tx, eventID uuid.UUID) error {
	// If the user owns the event, proceed with deletion
	result, err := tx.NewDelete().
		Model((*api.Event)(nil)).
		Where("id = ?", eventID).
		Exec(ctx)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no event found with the given ID")
	}

	return nil
}

// ----- Request handlers -----

func (h *Handlers) handleGetEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventIDStr := chi.URLParam(r, "eventID")

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		http.Error(w, "Invalid Event ID", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(ctx.Value("userID").(string))
	if err != nil {
		userID = uuid.Nil
	}

	details, err := h.getEventDetails(ctx, eventID, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching event details: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(details)
	if err != nil {
		slog.Error("Error encoding response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) handlePutEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventIDStr := chi.URLParam(r, "eventID")

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		http.Error(w, "Invalid Event ID", http.StatusBadRequest)
		return
	}

	// Get userID from context
	userID, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Check if the user is the owner of the event
	isOwner, err := h.isEventOwner(r.Context(), eventID, userID)
	if err != nil {
		http.Error(w, "Error checking event ownership", http.StatusInternalServerError)
		return
	}
	if !isOwner {
		http.Error(w, "User not authorized to update this event", http.StatusForbidden)
		return
	}

	// Parse the request body
	var update EventUpdate
	err = json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start a transaction
	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Update the event
	err = h.updateEvent(ctx, tx, eventID, &update)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating event: %v", err), http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Event updated successfully"})
}

func (h *Handlers) handleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventIDStr := chi.URLParam(r, "eventID")

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		http.Error(w, "Invalid Event ID", http.StatusBadRequest)
		return
	}

	// Get userID from context
	userID, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Check if the user is the owner of the event
	isOwner, err := h.isEventOwner(ctx, eventID, userID)
	if err != nil {
		http.Error(w, "Error checking event ownership", http.StatusInternalServerError)
		return
	}
	if !isOwner {
		http.Error(w, "User not authorized to update this event", http.StatusForbidden)
		return
	}

	// Start a transaction
	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Delete the event
	err = h.DeleteEvent(ctx, tx, eventID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting event: %v", err), http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Event deleted successfully"})
}

func (h *Handlers) handlePostEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var input EventInput

	// Parse JSON input
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	// Start a transaction
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Insert Event
	input.Event.ID = uuid.New()
	_, err = tx.NewInsert().Model(&input.Event).Exec(ctx)
	if err != nil {
		http.Error(w, "Failed to insert event", http.StatusInternalServerError)
		return
	}

	// Insert EventLocation
	_, err = tx.NewInsert().Model(&input.Location).Exec(ctx)
	if err != nil {
		http.Error(w, "Failed to insert event location", http.StatusInternalServerError)
		return
	}

	// Insert EventCategories
	for _, category := range input.Categories {
		eventCategory := api.EventCategory{
			EventID:  input.Event.ID,
			Category: category,
		}
		_, err = tx.NewInsert().Model(&eventCategory).Exec(ctx)
		if err != nil {
			http.Error(w, "Failed to insert event category", http.StatusInternalServerError)
			return
		}
	}

	// Insert EventAssets
	for _, asset := range input.Assets {
		asset.EventID = input.Event.ID
		_, err = tx.NewInsert().Model(&asset).Exec(ctx)
		if err != nil {
			http.Error(w, "Failed to insert event asset", http.StatusInternalServerError)
			return
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := map[string]interface{}{
		"message": "Event created successfully",
		"eventID": input.Event.ID,
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
