package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"golang.org/x/sync/errgroup"

	"github.com/roundtown-app/roundtown-api/api"
	"github.com/roundtown-app/roundtown-api/internal/middleware"
)

// ----- GET helper struct -----

type VenueDetails struct {
	Venue        api.Venue            `json:"venue"`
	Categories   []api.VenueCategory  `json:"categories"`
	Location     api.VenueLocation    `json:"location"`
	Assets       []api.VenueAssets    `json:"assets"`
	Population   api.VenuePopulation  `json:"population"`
	IsSaved      bool                 `json:"is_saved"`
	IsVisited    bool                 `json:"is_visited"`
	IsSharedFrom bool                 `json:"is_shared_from"`
	IsSharedTo   bool                 `json:"is_shared_to"`
	Exceptions   []api.VenueException `json:"exceptions"`
	Hours        []api.VenueHours     `json:"hours"`
}

// ----- PUT helper structs -----

type VenueUpdate struct {
	Title       *string              `json:"title"`
	Logo        *string              `json:"logo"`
	Description *string              `json:"description"`
	Detailed    *string              `json:"detailed"`
	Price       *int                 `json:"price"`
	Sponsor     *int                 `json:"sponsor"`
	Hidden      *bool                `json:"hidden"`
	Categories  []string             `json:"categories"`
	Location    *VenueLocationUpdate `json:"location"`
	Assets      []string             `json:"assets"`
	Exceptions  []api.VenueException `json:"exceptions"`
	Hours       []api.VenueHours     `json:"hours"`
}

type VenueLocationUpdate struct {
	ID        int      `json:"id"`
	Longitude *float64 `json:"longitude"`
	Latitude  *float64 `json:"latitude"`
	Address   *string  `json:"address"`
}

// ----- Search helper struct -----

type VenueSearchParams struct {
	Title                  *string    `json:"title"`
	Description            *string    `json:"description"`
	LocationLongitude      *float64   `json:"location_longitude"`
	LocationLatitude       *float64   `json:"location_latitude"`
	LocationAddress        *string    `json:"location_address"`
	Price                  *int       `json:"price"`
	OwnerID                *uuid.UUID `json:"owner_id"`
	Categories             *[]string  `json:"categories"`
	PopulationUserCountMin *int       `json:"population_user_count_min"`
	PopulationUserCountMax *int       `json:"population_user_count_max"`
	HasFutureEvents        *bool      `json:"has_future_events"`
	IsSaved                *bool      `json:"is_saved"`
}

// ----- POST helper struct -----

type VenueRequest struct {
	Title       string `json:"title"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
	Detailed    string `json:"detailed"`
	Location    struct {
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
		Address   string  `json:"address"`
	} `json:"location"`
	Categories []string `json:"categories"`
	Price      int      `json:"price"`
	Sponsor    int      `json:"sponsor"`
	Hidden     bool     `json:"hidden"`
	OwnerID    string   `json:"owner_id"`
	Assets     []string `json:"assets"`
	Hours      []struct {
		Type        string `json:"type"`
		Day         int    `json:"day"`
		OpeningTime string `json:"opening_time"`
		ClosingTime string `json:"closing_time"`
		IsClosed    bool   `json:"is_closed"`
	} `json:"hours,omitempty"`
	Exceptions []struct {
		ExceptionDate      string `json:"exception_date"`
		IsClosed           bool   `json:"is_closed"`
		AlternateStartTime string `json:"alternate_start_time,omitempty"`
		AlternateEndTime   string `json:"alternate_end_time,omitempty"`
	} `json:"exceptions,omitempty"`
}

// ----- GET helper functions -----

func (h *Handlers) getVenueDetails(ctx context.Context, venueID uuid.UUID, userID uuid.UUID) (*VenueDetails, error) {
	var details VenueDetails

	// Fetch the main venue
	err := h.DB.NewSelect().
		Model(&details.Venue).
		Where("id = ?", venueID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch venue: %w", err)
	}

	// Fetch categories
	err = h.DB.NewSelect().
		Model(&details.Categories).
		Where("venue_id = ?", venueID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch venue categories: %w", err)
	}

	// Fetch location
	err = h.DB.NewSelect().
		Model(&details.Location).
		Where("id = ?", details.Venue.LocationID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch venue location: %w", err)
	}

	// Fetch assets
	err = h.DB.NewSelect().
		Model(&details.Assets).
		Where("venue_id = ?", venueID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch venue assets: %w", err)
	}

	// Fetch population
	err = h.DB.NewSelect().
		Model(&details.Population).
		Where("venue_id = ?", venueID).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			// If no population record exists, create a default one
			details.Population = api.VenuePopulation{
				VenueID:     venueID,
				UserCount:   0,
				LastUpdated: time.Now(),
			}
		} else {
			return nil, fmt.Errorf("failed to fetch venue population: %w", err)
		}
	}

	// Fetch exceptions
	err = h.DB.NewSelect().
		Model(&details.Exceptions).
		Where("venue_id = ?", venueID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch venue exceptions: %w", err)
	}

	// Fetch hours
	err = h.DB.NewSelect().
		Model(&details.Hours).
		Where("venue_id = ?", venueID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch venue hours: %w", err)
	}

	if userID != uuid.Nil {
		// Check if the venue is saved by the user
		var savedItem api.SavedItem
		err = h.DB.NewSelect().
			Model(&savedItem).
			Where("user_id = ? AND venue_id = ?", userID, venueID).
			Scan(ctx)
		details.IsSaved = err == nil

		// Check if the venue is visited by the user
		var visitedItem api.VisitedItem
		err = h.DB.NewSelect().
			Model(&visitedItem).
			Where("user_id = ? AND venue_id = ?", userID, venueID).
			Scan(ctx)
		details.IsVisited = err == nil

		// Check if the venue is shared by the user
		var sharedItemFrom api.SharedItem
		err = h.DB.NewSelect().
			Model(&sharedItemFrom).
			Where("from_user_id = ? AND venue_id = ?", userID, venueID).
			Scan(ctx)
		details.IsSharedFrom = err == nil

		// Check if the venue is shared to the user
		var sharedItemTo api.SharedItem
		err = h.DB.NewSelect().
			Model(&sharedItemTo).
			Where("from_user_id = ? AND venue_id = ?", userID, venueID).
			Scan(ctx)
		details.IsSharedFrom = err == nil
	}

	return &details, nil
}

// ----- PUT helper functions -----

func (h *Handlers) isVenueOwner(ctx context.Context, venueID, userID uuid.UUID) (bool, error) {
	if ctx.Value(middleware.AccountTypeKey).(string) == "admin" {
		return true, nil
	}

	var venue api.Venue
	err := h.DB.NewSelect().
		Model(&venue).
		Where("id = ? AND owner_id = ?", venueID, userID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (h *Handlers) updateVenue(ctx context.Context, tx bun.Tx, venueID uuid.UUID, update *VenueUpdate) error {
	// Update main venue details
	_, err := tx.NewUpdate().
		Model(&api.Venue{}).
		Where("id = ?", venueID).
		Set("title = COALESCE(?, title)", update.Title).
		Set("logo = COALESCE(?, logo)", update.Logo).
		Set("description = COALESCE(?, description)", update.Description).
		Set("detailed = COALESCE(?, detailed)", update.Detailed).
		Set("price = COALESCE(?, price)", update.Price).
		Set("sponsor = COALESCE(?, sponsor)", update.Sponsor).
		Set("hidden = COALESCE(?, hidden)", update.Hidden).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update venue: %w", err)
	}

	// Update categories
	if update.Categories != nil {
		err = h.updateVenueCategories(ctx, tx, venueID, update.Categories)
		if err != nil {
			return fmt.Errorf("failed to update venue categories: %w", err)
		}
	}

	// Update location
	if update.Location != nil {
		err = h.updateVenueLocation(ctx, tx, venueID, update.Location)
		if err != nil {
			return fmt.Errorf("failed to update venue location: %w", err)
		}
	}

	// Update assets
	if update.Assets != nil {
		err = h.updateVenueAssets(ctx, tx, venueID, update.Assets)
		if err != nil {
			return fmt.Errorf("failed to update venue assets: %w", err)
		}
	}

	// Update exceptions
	if update.Exceptions != nil {
		err = h.updateVenueExceptions(ctx, tx, venueID, update.Exceptions)
		if err != nil {
			return fmt.Errorf("failed to update venue exceptions: %w", err)
		}
	}

	// Update hours
	if update.Hours != nil {
		err = h.updateVenueHours(ctx, tx, venueID, update.Hours)
		if err != nil {
			return fmt.Errorf("failed to update venue hours: %w", err)
		}
	}

	return nil
}

func (h *Handlers) updateVenueLocation(ctx context.Context, tx bun.Tx, venueID uuid.UUID, location *VenueLocationUpdate) error {
	_, err := tx.NewUpdate().
		Model(&api.VenueLocation{}).
		Where("id = ?", location.ID).
		Set("longitude = COALESCE(?, longitude)", location.Longitude).
		Set("latitude = COALESCE(?, latitude)", location.Latitude).
		Set("address = COALESCE(?, address)", location.Address).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Update the location_id in the main venue table
	_, err = tx.NewUpdate().
		Model(&api.Venue{}).
		Where("id = ?", venueID).
		Set("location_id = ?", location.ID).
		Exec(ctx)

	return err
}

func (h *Handlers) updateVenueCategories(ctx context.Context, tx bun.Tx, venueID uuid.UUID, newCategories []string) error {
	// Fetch existing categories
	var existingCategories []api.VenueCategory
	err := tx.NewSelect().
		Model(&existingCategories).
		Where("venue_id = ?", venueID).
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
	var categoriesToAdd []api.VenueCategory
	for _, cat := range newCategories {
		if !existingMap[cat] {
			categoriesToAdd = append(categoriesToAdd, api.VenueCategory{VenueID: venueID, Category: cat})
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
			Model((*api.VenueCategory)(nil)).
			Where("venue_id = ? AND category IN (?)", venueID, bun.In(categoriesToRemove)).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete old categories: %w", err)
		}
	}

	return nil
}

func (h *Handlers) updateVenueAssets(ctx context.Context, tx bun.Tx, venueID uuid.UUID, newAssets []string) error {
	// Fetch existing assets
	var existingAssets []api.VenueAssets
	err := tx.NewSelect().
		Model(&existingAssets).
		Where("venue_id = ?", venueID).
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
	var assetsToAdd []api.VenueAssets
	for _, assetID := range newAssets {
		if !existingMap[assetID] {
			assetsToAdd = append(assetsToAdd, api.VenueAssets{VenueID: venueID, AssetID: assetID})
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
			Model((*api.VenueAssets)(nil)).
			Where("venue_id = ? AND asset_id IN (?)", venueID, bun.In(assetsToRemove)).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete old assets: %w", err)
		}
	}

	return nil
}

func (h *Handlers) updateVenueExceptions(ctx context.Context, tx bun.Tx, venueID uuid.UUID, newExceptions []api.VenueException) error {
	// Delete existing exceptions
	_, err := tx.NewDelete().
		Model((*api.VenueException)(nil)).
		Where("venue_id = ?", venueID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete existing exceptions: %w", err)
	}

	// Insert new exceptions
	for _, exception := range newExceptions {
		exception.VenueID = venueID.String()
		_, err := tx.NewInsert().
			Model(&exception).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to insert new exception: %w", err)
		}
	}

	return nil
}

func (h *Handlers) updateVenueHours(ctx context.Context, tx bun.Tx, venueID uuid.UUID, newHours []api.VenueHours) error {
	// Delete existing hours
	_, err := tx.NewDelete().
		Model((*api.VenueHours)(nil)).
		Where("venue_id = ?", venueID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete existing hours: %w", err)
	}

	// Insert new hours
	for _, hours := range newHours {
		hours.VenueID = venueID
		_, err := tx.NewInsert().
			Model(&hours).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to insert new hours: %w", err)
		}
	}

	return nil
}

// ----- DELETE helper function -----

func (h *Handlers) DeleteVenue(ctx context.Context, tx bun.Tx, venueID uuid.UUID) error {
	// If the user owns the venue, proceed with deletion
	result, err := tx.NewDelete().
		Model((*api.Venue)(nil)).
		Where("id = ?", venueID).
		Exec(ctx)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no venue found with the given ID")
	}

	return nil
}

// ----- Request handlers -----

func (h *Handlers) handleGetVenue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	venueIDStr := chi.URLParam(r, "venueID")

	venueID, err := uuid.Parse(venueIDStr)
	if err != nil {
		http.Error(w, "Invalid Venue ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	details, err := h.getVenueDetails(ctx, venueID, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching venue details: %v", err), http.StatusInternalServerError)
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

func (h *Handlers) handleBulkGetVenue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var request struct {
		Venues []string `json:"venues"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(request.Venues) == 0 {
		http.Error(w, "No venue IDs provided", http.StatusBadRequest)
		return
	}

	venueIDs := make([]uuid.UUID, 0, len(request.Venues))
	for _, idStr := range request.Venues {
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid venue ID: %s", idStr), http.StatusBadRequest)
			return
		}
		venueIDs = append(venueIDs, id)
	}

	var response struct {
		Venues []VenueDetails `json:"venues"`
	}

	for _, venueID := range venueIDs {
		details, err := h.getVenueDetails(ctx, venueID, userID)
		if err != nil {
			slog.Error("Error fetching venue details", "error", err, "venueID", venueID)
			continue // Skip this venue and continue with others
		}
		response.Venues = append(response.Venues, *details)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Error encoding response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) handlePutVenue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	venueIDStr := chi.URLParam(r, "venueID")

	venueID, err := uuid.Parse(venueIDStr)
	if err != nil {
		http.Error(w, "Invalid Venue ID", http.StatusBadRequest)
		return
	}

	// Get userID from context
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Check if the user is the owner of the venue
	isOwner, err := h.isVenueOwner(r.Context(), venueID, userID)
	if err != nil {
		http.Error(w, "Error checking venue ownership", http.StatusInternalServerError)
		return
	}
	if !isOwner {
		http.Error(w, "User not authorized to update this venue", http.StatusForbidden)
		return
	}

	// Parse the request body
	var update VenueUpdate
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

	// Update the venue
	err = h.updateVenue(ctx, tx, venueID, &update)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating venue: %v", err), http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Venue updated successfully"})
}

func (h *Handlers) handleDeleteVenue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	venueIDStr := chi.URLParam(r, "venueID")

	venueID, err := uuid.Parse(venueIDStr)
	if err != nil {
		http.Error(w, "Invalid Venue ID", http.StatusBadRequest)
		return
	}

	// Get userID from context
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Check if the user is the owner of the venue
	isOwner, err := h.isVenueOwner(ctx, venueID, userID)
	if err != nil {
		http.Error(w, "Error checking venue ownership", http.StatusInternalServerError)
		return
	}
	if !isOwner {
		http.Error(w, "User not authorized to update this venue", http.StatusForbidden)
		return
	}

	// Start a transaction
	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Delete the venue
	err = h.DeleteVenue(ctx, tx, venueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting venue: %v", err), http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Venue deleted successfully"})
}

func (h *Handlers) handleVenueSearch(w http.ResponseWriter, r *http.Request) {
	var searchParams VenueSearchParams

	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Parse JSON input
	err := json.NewDecoder(r.Body).Decode(&searchParams)
	if err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	// Start building the query
	query := h.DB.NewSelect().
		Model((*api.Venue)(nil)).
		ColumnExpr("DISTINCT v.*").
		Join("LEFT JOIN venue_locations AS vl ON v.location_id = vl.id").
		Join("LEFT JOIN venue_categories AS vc ON v.id = vc.venue_id").
		Join("LEFT JOIN venue_population AS vp ON v.id = vp.venue_id")

	// Apply search filters
	if searchParams.Title != nil {
		query = query.Where("v.title ILIKE ?", "%"+*searchParams.Title+"%")
	}
	if searchParams.Description != nil {
		query = query.Where("v.description ILIKE ?", "%"+*searchParams.Description+"%")
	}
	if searchParams.LocationLongitude != nil {
		query = query.Where("vl.longitude = ?", *searchParams.LocationLongitude)
	}
	if searchParams.LocationLatitude != nil {
		query = query.Where("vl.latitude = ?", *searchParams.LocationLatitude)
	}
	if searchParams.LocationAddress != nil {
		query = query.Where("vl.address ILIKE ?", "%"+*searchParams.LocationAddress+"%")
	}
	if searchParams.Price != nil {
		query = query.Where("v.price = ?", *searchParams.Price)
	}
	if searchParams.OwnerID != nil {
		query = query.Where("v.owner_id = ?", *searchParams.OwnerID)
	}
	if searchParams.Categories != nil && len(*searchParams.Categories) > 0 {
		query = query.Where("vc.category IN (?)", bun.In(*searchParams.Categories))
	}
	if searchParams.PopulationUserCountMin != nil {
		query = query.Where("vp.user_count >= ?", *searchParams.PopulationUserCountMin)
	}
	if searchParams.PopulationUserCountMax != nil {
		query = query.Where("vp.user_count <= ?", *searchParams.PopulationUserCountMax)
	}
	if searchParams.HasFutureEvents != nil && *searchParams.HasFutureEvents {
		query = query.Where("EXISTS (SELECT 1 FROM events e WHERE e.venue_id = v.id AND e.ending_time > NOW())")
	}
	if searchParams.IsSaved != nil && *searchParams.IsSaved {
		query = query.Join("INNER JOIN saved_items AS si ON v.id = si.venue_id AND si.user_id = ?", userID)
	}

	// Execute the query
	var venues []api.Venue
	err = query.Scan(r.Context(), &venues)
	if err != nil {
		http.Error(w, "Error executing search query", http.StatusInternalServerError)
		return
	}

	// Create a slice to store venue details
	venueDetails := make([]*VenueDetails, 0, len(venues))

	// Create error group for concurrent fetching
	g, ctx := errgroup.WithContext(r.Context())
	detailsMutex := &sync.Mutex{}

	// Fetch details for each venue
	for _, venue := range venues {
		venue := venue // Create new variable for goroutine
		g.Go(func() error {
			details, err := h.getVenueDetails(ctx, venue.ID, userID)
			if err != nil {
				return fmt.Errorf("error fetching details for venue %s: %w", venue.ID, err)
			}

			detailsMutex.Lock()
			venueDetails = append(venueDetails, details)
			detailsMutex.Unlock()

			return nil
		})
	}

	// Wait for all detail fetches to complete
	if err := g.Wait(); err != nil {
		http.Error(w, "Error fetching venue details", http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := map[string]interface{}{
		"venues": venueDetails,
		"count":  len(venueDetails),
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) handlePostVenue(w http.ResponseWriter, r *http.Request) {
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
		slog.Error(userID.String() + " is not a business or admin, so they cannot create a new venue")
		http.Error(w, "User does not have permission", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var req VenueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error(err.Error())
		http.Error(w, "Invalid request body", http.StatusBadRequest)
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
	new_venue_id, err := uuid.NewV7()
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Failed to generate UUID", http.StatusInternalServerError)
		return
	}

	new_venue_location_id, err := uuid.NewV7()
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Failed to generate UUID", http.StatusInternalServerError)
		return
	}

	// Create venue
	venue := &api.Venue{
		ID:          new_venue_id,
		Title:       req.Title,
		Logo:        req.Logo,
		Description: req.Description,
		Detailed:    req.Detailed,
		LocationID:  new_venue_location_id,
		Price:       req.Price,
		Sponsor:     req.Sponsor,
		Hidden:      req.Hidden,
		OwnerID:     userID,
		CreatedAt:   time.Now(),
	}

	if _, err := tx.NewInsert().Model(venue).Exec(r.Context()); err != nil {
		slog.Error(err.Error())
		http.Error(w, "Failed to create venue", http.StatusInternalServerError)
		return
	}

	// Create venue location
	location := &api.VenueLocation{
		ID:        new_venue_location_id,
		Longitude: req.Location.Longitude,
		Latitude:  req.Location.Latitude,
		Address:   req.Location.Address,
	}

	if _, err := tx.NewInsert().Model(location).Exec(r.Context()); err != nil {
		slog.Error(err.Error())
		http.Error(w, "Failed to create venue location", http.StatusInternalServerError)
		return
	}

	// Create venue categories
	if len(req.Categories) > 0 {
		categories := make([]api.VenueCategory, len(req.Categories))
		for i, category := range req.Categories {
			categories[i] = api.VenueCategory{
				VenueID:  venue.ID,
				Category: category,
			}
		}
		if _, err := tx.NewInsert().Model(&categories).Exec(r.Context()); err != nil {
			slog.Error(err.Error())
			http.Error(w, "Failed to create venue categories", http.StatusInternalServerError)
			return
		}
	}

	// Create venue assets
	if len(req.Assets) > 0 {
		assets := make([]api.VenueAssets, len(req.Assets))
		for i, assetID := range req.Assets {
			assets[i] = api.VenueAssets{
				VenueID: venue.ID,
				AssetID: assetID,
			}
		}
		if _, err := tx.NewInsert().Model(&assets).Exec(r.Context()); err != nil {
			slog.Error(err.Error())
			http.Error(w, "Failed to create venue assets", http.StatusInternalServerError)
			return
		}
	}

	// Create venue hours if provided
	if len(req.Hours) > 0 {
		//hours := make([]api.VenueHours, len(req.Hours))
		for _, day_hours := range req.Hours {
			openTime, _ := time.Parse("0000-01-01 15:04:00+00:00", day_hours.OpeningTime)
			closeTime, _ := time.Parse("0000-01-01 15:04:00+00:00", day_hours.ClosingTime)

			hours_instance := api.VenueHours{
				VenueID:     venue.ID,
				Type:        sql.NullString{String: day_hours.Type, Valid: day_hours.Type != ""},
				Day:         day_hours.Day,
				OpeningTime: openTime,
				ClosingTime: closeTime,
				IsClosed:    day_hours.IsClosed,
			}

			if _, err := tx.NewInsert().Model(&hours_instance).Exec(r.Context()); err != nil {
				slog.Error(err.Error())
				http.Error(w, "Failed to create venue hours", http.StatusInternalServerError)
				return
			}
		}
	}

	// Create venue exceptions if provided
	if len(req.Exceptions) > 0 {
		exceptions := make([]api.VenueException, len(req.Exceptions))
		for i, e := range req.Exceptions {
			exceptionDate, _ := time.Parse("2006-01-02", e.ExceptionDate)
			startTime, _ := time.Parse("15:04", e.AlternateStartTime)
			endTime, _ := time.Parse("15:04", e.AlternateEndTime)

			exceptions[i] = api.VenueException{
				VenueID:               venue.ID.String(),
				ExceptionDate:         exceptionDate,
				IsClosed:              e.IsClosed,
				AlternateStartingTime: startTime,
				AlternateEndingTime:   endTime,
			}
		}
		if _, err := tx.NewInsert().Model(&exceptions).Exec(r.Context()); err != nil {
			slog.Error(err.Error())
			http.Error(w, "Failed to create venue exceptions", http.StatusInternalServerError)
			return
		}
	}

	// Initialize venue population
	population := &api.VenuePopulation{
		VenueID:     venue.ID,
		UserCount:   0,
		LastUpdated: time.Now(),
	}
	if _, err := tx.NewInsert().Model(population).Exec(r.Context()); err != nil {
		slog.Error(err.Error())
		http.Error(w, "Failed to create venue population", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		slog.Error(err.Error())
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Return the created venue ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id": venue.ID.String(),
	})
}
