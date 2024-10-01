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

type VenueDetails struct {
	Venue        api.Venue           `json:"venue"`
	Categories   []api.VenueCategory `json:"categories"`
	Location     api.VenueLocation   `json:"location"`
	Ratings      []api.VenueRating   `json:"ratings"`
	Assets       []api.VenueAssets   `json:"assets"`
	Population   api.VenuePopulation `json:"population"`
	IsSaved      bool                `json:"is_saved"`
	IsVisited    bool                `json:"is_visited"`
	IsSharedFrom bool                `json:"is_shared_from"`
	IsSharedTo   bool                `json:"is_shared_to"`
}

// ----- PUT helper structs -----

type VenueUpdate struct {
	Title       *string              `json:"title"`
	Description *string              `json:"description"`
	Detailed    *string              `json:"detailed"`
	Price       *int                 `json:"price"`
	Sponsor     *int                 `json:"sponsor"`
	Hidden      *bool                `json:"hidden"`
	Categories  []string             `json:"categories"`
	Location    *VenueLocationUpdate `json:"location"`
	Assets      []string             `json:"assets"`
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
	Category               *string    `json:"category"`
	PopulationUserCountMin *int       `json:"population_user_count_min"`
	PopulationUserCountMax *int       `json:"population_user_count_max"`
}

// ----- POST helper struct -----

type VenueInput struct {
	Venue      api.Venue         `json:"venue"`
	Categories []string          `json:"categories"`
	Location   api.VenueLocation `json:"location"`
	Assets     []api.VenueAssets `json:"assets"`
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

	// Fetch ratings
	err = h.DB.NewSelect().
		Model(&details.Ratings).
		Where("venue_id = ?", venueID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch venue ratings: %w", err)
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
	if ctx.Value("userRole").(string) == "admin" {
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

	userID, err := uuid.Parse(ctx.Value("userID").(string))
	if err != nil {
		userID = uuid.Nil
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

func (h *Handlers) handlePutVenue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	venueIDStr := chi.URLParam(r, "venueID")

	venueID, err := uuid.Parse(venueIDStr)
	if err != nil {
		http.Error(w, "Invalid Venue ID", http.StatusBadRequest)
		return
	}

	// Get userID from context
	userID, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
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
	userID, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
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

	// Parse JSON input
	err := json.NewDecoder(r.Body).Decode(&searchParams)
	if err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	// Start building the query
	query := h.DB.NewSelect().
		Model((*api.Event)(nil)).
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
	if searchParams.Category != nil {
		query = query.Where("vc.category = ?", *searchParams.Category)
	}
	if searchParams.PopulationUserCountMin != nil {
		query = query.Where("vp.user_count >= ?", *searchParams.PopulationUserCountMin)
	}
	if searchParams.PopulationUserCountMax != nil {
		query = query.Where("vp.user_count <= ?", *searchParams.PopulationUserCountMax)
	}

	// Execute the query
	var venues []api.Venue
	err = query.Scan(r.Context(), &venues)
	if err != nil {
		http.Error(w, "Error executing search query", http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := map[string]interface{}{
		"events": venues,
		"count":  len(venues),
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) handlePostVenue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var input VenueInput

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

	// Insert Venue
	input.Venue.ID = uuid.New()
	_, err = tx.NewInsert().Model(&input.Venue).Exec(ctx)
	if err != nil {
		http.Error(w, "Failed to insert venue", http.StatusInternalServerError)
		return
	}

	// Insert VenueLocation
	_, err = tx.NewInsert().Model(&input.Location).Exec(ctx)
	if err != nil {
		http.Error(w, "Failed to insert venue location", http.StatusInternalServerError)
		return
	}

	// Insert VenueCategories
	for _, category := range input.Categories {
		venueCategory := api.VenueCategory{
			VenueID:  input.Venue.ID,
			Category: category,
		}
		_, err = tx.NewInsert().Model(&venueCategory).Exec(ctx)
		if err != nil {
			http.Error(w, "Failed to insert venue category", http.StatusInternalServerError)
			return
		}
	}

	// Insert VenueAssets
	for _, asset := range input.Assets {
		asset.VenueID = input.Venue.ID
		_, err = tx.NewInsert().Model(&asset).Exec(ctx)
		if err != nil {
			http.Error(w, "Failed to insert venue asset", http.StatusInternalServerError)
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
		"message": "Venue created successfully",
		"venueID": input.Venue.ID,
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
