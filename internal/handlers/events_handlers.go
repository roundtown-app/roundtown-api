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
	"github.com/roundtown-app/roundtown-api/internal/middleware"
)

// ----- GET helper struct -----

type EventDetails struct {
	Event             api.Event                  `json:"event"`
	Venue             api.Venue                  `json:"venue"`
	Categories        []api.EventCategory        `json:"categories"`
	Location          api.EventLocation          `json:"location"`
	Assets            []api.EventAssets          `json:"assets"`
	Population        api.EventPopulation        `json:"population"`
	IsSaved           bool                       `json:"is_saved"`
	IsVisited         bool                       `json:"is_visited"`
	IsSharedFrom      bool                       `json:"is_shared_from"`
	IsSharedTo        bool                       `json:"is_shared_to"`
	RecurrencePattern api.EventRecurrencePattern `json:"recurrence_pattern,omitempty"`
	Exceptions        []api.EventException       `json:"exceptions,omitempty"`
}

// ----- PUT helper structs -----

type EventUpdate struct {
	Title             *string                     `json:"title"`
	Description       *string                     `json:"description"`
	Detailed          *string                     `json:"detailed"`
	Tag               *string                     `json:"tag"`
	Price             *int                        `json:"price"`
	Sponsor           *int                        `json:"sponsor"`
	IsDeal            *bool                       `json:"is_deal"`
	StartingTime      *time.Time                  `json:"starting_time"`
	EndingTime        *time.Time                  `json:"ending_time"`
	VenueID           *uuid.UUID                  `json:"venue_id"`
	Categories        []string                    `json:"categories"`
	Location          *EventLocationUpdate        `json:"location"`
	Assets            []string                    `json:"assets"`
	RecurrencePattern *api.EventRecurrencePattern `json:"recurrence_pattern"`
	Exceptions        []api.EventException        `json:"exceptions"`
}

type EventLocationUpdate struct {
	ID        int      `json:"id"`
	Longitude *float64 `json:"longitude"`
	Latitude  *float64 `json:"latitude"`
	Address   *string  `json:"address"`
}

// ----- Search helper struct -----

type EventSearchParams struct {
	Title                  *string    `json:"title"`
	Description            *string    `json:"description"`
	LocationLongitude      *float64   `json:"location_longitude"`
	LocationLatitude       *float64   `json:"location_latitude"`
	LocationAddress        *string    `json:"location_address"`
	Price                  *int       `json:"price"`
	IsDeal                 *bool      `json:"is_deal"`
	StartingTime           *time.Time `json:"starting_time"`
	EndingTime             *time.Time `json:"ending_time"`
	VenueID                *uuid.UUID `json:"venue_id"`
	VenueTitle             *string    `json:"venue_title"`
	OwnerID                *uuid.UUID `json:"owner_id"`
	Category               *string    `json:"category"`
	PopulationUserCountMin *int       `json:"population_user_count_min"`
	PopulationUserCountMax *int       `json:"population_user_count_max"`
}

// ----- POST helper struct -----

type EventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Detailed    string `json:"detailed"`
	Tag         string `json:"tag"`
	Location    struct {
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
		Address   string  `json:"address"`
	} `json:"location"`
	Categories        []string `json:"categories"`
	Price             int      `json:"price"`
	Sponsor           int      `json:"sponsor"`
	IsDeal            bool     `json:"is_deal"`
	StartingTime      string   `json:"starting_time"`
	EndingTime        string   `json:"ending_time"`
	VenueID           string   `json:"venue_id"`
	OwnerID           string   `json:"owner_id"`
	Assets            []string `json:"assets"`
	RecurrencePattern struct {
		Frequency   string `json:"frequency"`
		DaysOfWeek  []int  `json:"days_of_week,omitempty"`
		WeekOfMonth []int  `json:"week_of_month,omitempty"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
	} `json:"recurrence_pattern,omitempty"`
	Exceptions []struct {
		ExceptionDate      string `json:"exception_date"`
		IsCancelled        bool   `json:"is_cancelled"`
		AlternateStartTime string `json:"alternate_start_time,omitempty"`
		AlternateEndTime   string `json:"alternate_end_time,omitempty"`
	} `json:"exceptions,omitempty"`
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

	// Fetch venue details
	err = h.DB.NewSelect().
		Model(&details.Venue).
		Where("id = ?", details.Event.VenueID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch venue: %w", err)
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

	// Fetch recurrence pattern
	err = h.DB.NewSelect().
		Model(&details.RecurrencePattern).
		Where("event_id = ?", eventID).
		Scan(ctx)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch event recurrence pattern: %w", err)
	}

	// Fetch exceptions
	err = h.DB.NewSelect().
		Model(&details.Exceptions).
		Where("event_id = ?", eventID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event exceptions: %w", err)
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
	if ctx.Value(middleware.AccountTypeKey).(string) == "admin" {
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
		Set("tag = COALESCE(?, tag)", update.Tag).
		Set("price = COALESCE(?, price)", update.Price).
		Set("sponsor = COALESCE(?, sponsor)", update.Sponsor).
		Set("is_deal = COALESCE(?, is_deal)", update.IsDeal).
		Set("starting_time = COALESCE(?, starting_time)", update.StartingTime).
		Set("ending_time = COALESCE(?, ending_time)", update.EndingTime).
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

	// Update recurrence pattern
	if update.RecurrencePattern != nil {
		err = h.updateEventRecurrencePattern(ctx, tx, eventID, update.RecurrencePattern)
		if err != nil {
			return fmt.Errorf("failed to update event recurrence pattern: %w", err)
		}
	}

	// Update exceptions
	if update.Exceptions != nil {
		err = h.updateEventExceptions(ctx, tx, eventID, update.Exceptions)
		if err != nil {
			return fmt.Errorf("failed to update event exceptions: %w", err)
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

func (h *Handlers) updateEventRecurrencePattern(ctx context.Context, tx bun.Tx, eventID uuid.UUID, pattern *api.EventRecurrencePattern) error {
	if pattern == nil {
		_, err := tx.NewDelete().
			Model((*api.EventRecurrencePattern)(nil)).
			Where("event_id = ?", eventID).
			Exec(ctx)
		return err
	}

	pattern.EventID = eventID.String()
	_, err := tx.NewInsert().
		Model(pattern).
		On("CONFLICT (event_id) DO UPDATE").
		Set("frequency = EXCLUDED.frequency").
		Set("days_of_week = EXCLUDED.days_of_week").
		Set("week_of_month = EXCLUDED.week_of_month").
		Set("start_date = EXCLUDED.start_date").
		Set("end_date = EXCLUDED.end_date").
		Exec(ctx)
	return err
}

func (h *Handlers) updateEventExceptions(ctx context.Context, tx bun.Tx, eventID uuid.UUID, exceptions []api.EventException) error {
	if len(exceptions) == 0 {
		_, err := tx.NewDelete().
			Model((*api.EventException)(nil)).
			Where("event_id = ?", eventID).
			Exec(ctx)
		return err
	}

	for i := range exceptions {
		exceptions[i].EventID = eventID.String()
	}

	_, err := tx.NewInsert().
		Model(&exceptions).
		On("CONFLICT (event_id, exception_date) DO UPDATE").
		Set("is_cancelled = EXCLUDED.is_cancelled").
		Set("alternate_starting_time = EXCLUDED.alternate_starting_time").
		Set("alternate_ending_time = EXCLUDED.alternate_ending_time").
		Exec(ctx)
	return err
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

	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)

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

func (h *Handlers) handleBulkGetEvent(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)

    var request struct {
        Events []string `json:"events"`
    }

    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if len(request.Events) == 0 {
        http.Error(w, "No event IDs provided", http.StatusBadRequest)
        return
    }

    eventIDs := make([]uuid.UUID, 0, len(request.Events))
    for _, idStr := range request.Events {
        id, err := uuid.Parse(idStr)
        if err != nil {
            http.Error(w, fmt.Sprintf("Invalid event ID: %s", idStr), http.StatusBadRequest)
            return
        }
        eventIDs = append(eventIDs, id)
    }

    var response struct {
        Events []EventDetails `json:"events"`
    }

    for _, eventID := range eventIDs {
        details, err := h.getEventDetails(ctx, eventID, userID)
        if err != nil {
            slog.Error("Error fetching event details", "error", err, "eventID", eventID)
            continue // Skip this event and continue with others
        }
        response.Events = append(response.Events, *details)
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
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
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
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
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
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

func (h *Handlers) handleEventSearch(w http.ResponseWriter, r *http.Request) {
	var searchParams EventSearchParams

	// Parse JSON input
	err := json.NewDecoder(r.Body).Decode(&searchParams)
	if err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	// Start building the query
	query := h.DB.NewSelect().
		Model((*api.Event)(nil)).
		ColumnExpr("DISTINCT e.*").
		Join("LEFT JOIN event_locations AS el ON e.location_id = el.id").
		Join("LEFT JOIN event_categories AS ec ON e.id = ec.event_id").
		Join("LEFT JOIN event_population AS ep ON e.id = ep.event_id").
		Join("LEFT JOIN venues AS v ON e.venue_id = v.id")

	// Apply search filters
	if searchParams.Title != nil {
		query = query.Where("e.title ILIKE ?", "%"+*searchParams.Title+"%")
	}
	if searchParams.Description != nil {
		query = query.Where("e.description ILIKE ?", "%"+*searchParams.Description+"%")
	}
	if searchParams.LocationLongitude != nil {
		query = query.Where("el.longitude = ?", *searchParams.LocationLongitude)
	}
	if searchParams.LocationLatitude != nil {
		query = query.Where("el.latitude = ?", *searchParams.LocationLatitude)
	}
	if searchParams.LocationAddress != nil {
		query = query.Where("el.address ILIKE ?", "%"+*searchParams.LocationAddress+"%")
	}
	if searchParams.Price != nil {
		query = query.Where("e.price = ?", *searchParams.Price)
	}
	if searchParams.IsDeal != nil {
		query = query.Where("e.is_deal = ?", *searchParams.IsDeal)
	}
	if searchParams.StartingTime != nil {
		query = query.Where("e.starting_time >= ?", *searchParams.StartingTime)
	}
	if searchParams.EndingTime != nil {
		query = query.Where("e.ending_time <= ?", *searchParams.EndingTime)
	}
	if searchParams.VenueID != nil {
		query = query.Where("e.venue_id = ?", *searchParams.VenueID)
	}
	if searchParams.VenueTitle != nil {
		query = query.Where("v.title ILIKE ?", *searchParams.VenueTitle)
	}
	if searchParams.OwnerID != nil {
		query = query.Where("e.owner_id = ?", *searchParams.OwnerID)
	}
	if searchParams.Category != nil {
		query = query.Where("ec.category = ?", *searchParams.Category)
	}
	if searchParams.PopulationUserCountMin != nil {
		query = query.Where("ep.user_count >= ?", *searchParams.PopulationUserCountMin)
	}
	if searchParams.PopulationUserCountMax != nil {
		query = query.Where("ep.user_count <= ?", *searchParams.PopulationUserCountMax)
	}

	// Execute the query
	var events []api.Event
	err = query.Scan(r.Context(), &events)
	if err != nil {
		http.Error(w, "Error executing search query", http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := map[string]interface{}{
		"events": events,
		"count":  len(events),
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) handlePostEvent(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		slog.Error(userID.String() + " could not be auth'd")
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Check that the user is either a business or admin
	user_account_type := r.Context().Value(middleware.AccountTypeKey).(string)
	if !(user_account_type == "business" || user_account_type == "admin") {
		slog.Error(userID.String() + " is not a business or admin, so they cannot create a new event")
		http.Error(w, "User does not have permission", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var req EventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error(err.Error())
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check that the user owns the venue for which they're creating an event
	isOwner, err := h.isVenueOwner(r.Context(), uuid.MustParse(req.VenueID), userID)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if !isOwner {
		http.Error(w, "User does not have permission to create an event for given venue", http.StatusUnauthorized)
		return
	}

	// Start a transaction
	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Generate UUIDs
	new_event_id, err := uuid.NewV7()
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Failed to generate UUID", http.StatusInternalServerError)
		return
	}

	new_event_location_id, err := uuid.NewV7()
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Failed to generate UUID", http.StatusInternalServerError)
		return
	}

	// Create event
	startTime, err := time.Parse(time.RFC3339, req.StartingTime)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Failed to parse StartingTime", http.StatusInternalServerError)
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndingTime)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Failed to parse EndingTime", http.StatusInternalServerError)
		return
	}

	venueId, err := uuid.Parse(req.VenueID)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Failed to parse Venue UUID", http.StatusInternalServerError)
		return
	}

	event := &api.Event{
		ID:           new_event_id,
		Title:        req.Title,
		Description:  req.Description,
		Detailed:     req.Detailed,
		Tag:          req.Tag,
		LocationID:   new_event_location_id,
		Sponsor:      req.Sponsor,
		Price:        req.Price,
		IsDeal:       req.IsDeal,
		StartingTime: startTime,
		EndingTime:   endTime,
		OwnerID:      userID,
		VenueID:      venueId,
		CreatedAt:    time.Now(),
	}

	if _, err := tx.NewInsert().Model(event).Exec(r.Context()); err != nil {
		slog.Error(err.Error())
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	// Create event location
	location := &api.EventLocation{
		ID:        new_event_location_id,
		Longitude: req.Location.Longitude,
		Latitude:  req.Location.Latitude,
		Address:   req.Location.Address,
	}

	if _, err := tx.NewInsert().Model(location).Exec(r.Context()); err != nil {
		slog.Error(err.Error())
		http.Error(w, "Failed to create event location", http.StatusInternalServerError)
		return
	}

	// Create event categories
	if len(req.Categories) > 0 {
		categories := make([]api.EventCategory, len(req.Categories))
		for i, category := range req.Categories {
			categories[i] = api.EventCategory{
				EventID:  event.ID,
				Category: category,
			}
		}
		if _, err := tx.NewInsert().Model(&categories).Exec(r.Context()); err != nil {
			slog.Error(err.Error())
			http.Error(w, "Failed to create event categories", http.StatusInternalServerError)
			return
		}
	}

	// Create event assets
	if len(req.Assets) > 0 {
		assets := make([]api.EventAssets, len(req.Assets))
		for i, assetID := range req.Assets {
			assets[i] = api.EventAssets{
				EventID: event.ID,
				AssetID: assetID,
			}
		}
		if _, err := tx.NewInsert().Model(&assets).Exec(r.Context()); err != nil {
			slog.Error(err.Error())
			http.Error(w, "Failed to create event assets", http.StatusInternalServerError)
			return
		}
	}

	// Create event recurrence pattern if provided
	if req.RecurrencePattern.EndDate != "" {
		startDate, _ := time.Parse("2006-01-02", req.RecurrencePattern.StartDate)
		endDate, _ := time.Parse("2006-01-02", req.RecurrencePattern.EndDate)

		pattern := api.EventRecurrencePattern{
			EventID:     event.ID.String(),
			Frequency:   req.RecurrencePattern.Frequency,
			DaysOfWeek:  req.RecurrencePattern.DaysOfWeek,
			WeekOfMonth: req.RecurrencePattern.WeekOfMonth,
			StartDate:   startDate,
			EndDate:     endDate,
		}
		if _, err := tx.NewInsert().Model(&pattern).Exec(r.Context()); err != nil {
			slog.Error(err.Error())
			http.Error(w, "Failed to create event recurrence pattern", http.StatusInternalServerError)
			return
		}
	}

	// Create event exceptions if provided
	if len(req.Exceptions) > 0 {
		exceptions := make([]api.EventException, len(req.Exceptions))
		for i, e := range req.Exceptions {
			exceptionDate, _ := time.Parse("2006-01-02", e.ExceptionDate)
			startTime, _ := time.Parse("15:04", e.AlternateStartTime)
			endTime, _ := time.Parse("15:04", e.AlternateEndTime)

			exceptions[i] = api.EventException{
				EventID:               event.ID.String(),
				ExceptionDate:         exceptionDate,
				IsCancelled:           e.IsCancelled,
				AlternateStartingTime: startTime,
				AlternateEndingTime:   endTime,
			}
		}
		if _, err := tx.NewInsert().Model(&exceptions).Exec(r.Context()); err != nil {
			slog.Error(err.Error())
			http.Error(w, "Failed to create event exceptions", http.StatusInternalServerError)
			return
		}
	}

	// Initialize event population
	population := &api.EventPopulation{
		EventID:     event.ID,
		UserCount:   0,
		LastUpdated: time.Now(),
	}
	if _, err := tx.NewInsert().Model(population).Exec(r.Context()); err != nil {
		slog.Error(err.Error())
		http.Error(w, "Failed to create event population", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		slog.Error(err.Error())
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Return the created event ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id": event.ID.String(),
	})
}
