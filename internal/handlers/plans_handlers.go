package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/roundtown-app/roundtown-api/api"
	"github.com/roundtown-app/roundtown-api/internal/middleware"
)

// ----- GET helper struct -----

type PlanDetails struct {
	Plan     api.Plan       `json:"plan"`
	Venues	 []api.Venue	`json:"venues"`
	Events	 []api.Event	`json:"events"`
	Users    []api.PlanUser `json:"users"`
	IsOwner  bool           `json:"is_owner"`
	IsMember bool           `json:"is_member"`
}

// ----- PUT helper struct -----

type PlanUpdate struct {
	Title   *string        `json:"title"`
	Date    *time.Time     `json:"date"`
	Private *bool          `json:"private"`
	Items   []api.PlanItem `json:"items"`
	Users   []uuid.UUID    `json:"users"`
}

// ----- POST helper struct -----

type PlanInput struct {
	Plan  api.Plan       `json:"plan"`
	Items []api.PlanItem `json:"items"`
}

// ----- Search helper struct -----

type PlanSearchParams struct {
	Title     string    `json:"title"`
	CreatorID uuid.UUID `json:"creator_id"`
	Date      time.Time `json:"date"`
}

type PlanSearchResult struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Date      time.Time `json:"date"`
	CreatorID uuid.UUID `json:"creator_id"`
}

// ----- GET helper functions -----

func (h *Handlers) getPlanDetails(ctx context.Context, planID uuid.UUID, userID uuid.UUID) (*PlanDetails, error) {
	var details PlanDetails

	// Fetch the main plan
	err := h.DB.NewSelect().
		Model(&details.Plan).
		Where("id = ?", planID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plan: %w", err)
	}

	// Fetch plan items
	var planItems []api.PlanItem
	err = h.DB.NewSelect().
		Model(&planItems).
		Where("plan_id = ?", planID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plan items: %w", err)
	}

	for _, planItem := range planItems {
		var eventItem api.Event
		var venueItem api.Venue
		if planItem.EventID != uuid.Nil {
			err = h.DB.NewSelect().
				Model(&eventItem).
				Where("id = ?", planItem.EventID).
				Scan(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch plan item event: %w", err)
			}
			details.Events = append(details.Events, eventItem)
		} else {
			err = h.DB.NewSelect().
				Model(&venueItem).
				Where("id = ?", planItem.VenueID).
				Scan(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch plan item event: %w", err)
			}
			details.Venues = append(details.Venues, venueItem)
		}
	}

	// Fetch plan users
	err = h.DB.NewSelect().
		Model(&details.Users).
		Where("plan_id = ?", planID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plan users: %w", err)
	}

	// Check if the user is the owner
	details.IsOwner = details.Plan.CreatorID == userID

	// Check if the user is a member
	for _, user := range details.Users {
		if user.UserID == userID {
			details.IsMember = true
			break
		}
	}

	return &details, nil
}

// ----- PUT helper functions -----

func (h *Handlers) isPlanOwner(ctx context.Context, planID, userID uuid.UUID) (bool, error) {
	var plan api.Plan
	err := h.DB.NewSelect().
		Model(&plan).
		Where("id = ? AND creator_id = ?", planID, userID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (h *Handlers) updatePlan(ctx context.Context, tx bun.Tx, planID uuid.UUID, update *PlanUpdate) error {
	// Update main plan details
	_, err := tx.NewUpdate().
		Model(&api.Plan{}).
		Where("id = ?", planID).
		Set("title = COALESCE(?, title)", update.Title).
		Set("date = COALESCE(?, date)", update.Date).
		Set("private = COALESCE(?, private)", update.Private).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}

	// Update plan items
	if update.Items != nil {
		err = h.updatePlanItems(ctx, tx, planID, update.Items)
		if err != nil {
			return fmt.Errorf("failed to update plan items: %w", err)
		}
	}

	// Update plan users
	if update.Users != nil {
		err = h.updatePlanUsers(ctx, tx, planID, update.Users)
		if err != nil {
			return fmt.Errorf("failed to update plan users: %w", err)
		}
	}

	return nil
}

func (h *Handlers) updatePlanItems(ctx context.Context, tx bun.Tx, planID uuid.UUID, newItems []api.PlanItem) error {
	// Delete existing items
	_, err := tx.NewDelete().
		Model((*api.PlanItem)(nil)).
		Where("plan_id = ?", planID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete existing plan items: %w", err)
	}

	// Insert new items
	for _, item := range newItems {
		item.PlanID = planID
		_, err := tx.NewInsert().
			Model(&item).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to insert new plan item: %w", err)
		}
	}

	return nil
}

func (h *Handlers) updatePlanUsers(ctx context.Context, tx bun.Tx, planID uuid.UUID, newUsers []uuid.UUID) error {
	// Delete existing users
	_, err := tx.NewDelete().
		Model((*api.PlanUser)(nil)).
		Where("plan_id = ?", planID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete existing plan users: %w", err)
	}

	// Insert new users
	for _, userID := range newUsers {
		planUser := api.PlanUser{
			PlanID: planID,
			UserID: userID,
		}
		_, err := tx.NewInsert().
			Model(&planUser).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to insert new plan user: %w", err)
		}
	}

	return nil
}

// ----- Search helper function -----

func (h *Handlers) searchPlans(ctx context.Context, params PlanSearchParams) ([]PlanSearchResult, error) {
	var results []PlanSearchResult

	query := h.DB.NewSelect().Model((*api.Plan)(nil))

	if params.Title != "" {
		query = query.Where("title ILIKE ?", "%"+params.Title+"%")
	}

	if params.CreatorID != uuid.Nil {
		query = query.Where("creator_id = ?", params.CreatorID)
	}

	if !params.Date.IsZero() {
		query = query.Where("date::date = ?::date", params.Date)
	}

	err := query.Limit(10).Scan(ctx, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to search plans: %w", err)
	}

	return results, nil
}

// ----- DELETE helper function -----

func (h *Handlers) DeletePlan(ctx context.Context, tx bun.Tx, planID uuid.UUID) error {
	result, err := tx.NewDelete().
		Model((*api.Plan)(nil)).
		Where("id = ?", planID).
		Exec(ctx)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no plan found with the given ID")
	}

	return nil
}

// ----- Request handlers -----

func (h *Handlers) handleGetPlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	planIDStr := chi.URLParam(r, "planID")

	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		http.Error(w, "Invalid Plan ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	details, err := h.getPlanDetails(ctx, planID, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching plan details: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(details)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) handlePutPlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	planIDStr := chi.URLParam(r, "planID")

	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		http.Error(w, "Invalid Plan ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	isOwner, err := h.isPlanOwner(ctx, planID, userID)
	if err != nil {
		http.Error(w, "Error checking plan ownership", http.StatusInternalServerError)
		return
	}
	if !isOwner {
		http.Error(w, "User not authorized to update this plan", http.StatusForbidden)
		return
	}

	var update PlanUpdate
	err = json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	err = h.updatePlan(ctx, tx, planID, &update)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating plan: %v", err), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Plan updated successfully"})
}

func (h *Handlers) handlePlanSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var params PlanSearchParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	results, err := h.searchPlans(ctx, params)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error searching plans: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(results)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) handleDeletePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	planIDStr := chi.URLParam(r, "planID")

	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		http.Error(w, "Invalid Plan ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	isOwner, err := h.isPlanOwner(ctx, planID, userID)
	if err != nil {
		http.Error(w, "Error checking plan ownership", http.StatusInternalServerError)
		return
	}
	if !isOwner {
		http.Error(w, "User not authorized to delete this plan", http.StatusForbidden)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	err = h.DeletePlan(ctx, tx, planID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting plan: %v", err), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Plan deleted successfully"})
}

func (h *Handlers) handlePostPlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var input PlanInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Insert Plan
	input.Plan.ID = uuid.New()
	input.Plan.CreatedAt = time.Now()
	_, err = tx.NewInsert().Model(&input.Plan).Exec(ctx)
	if err != nil {
		http.Error(w, "Failed to insert plan", http.StatusInternalServerError)
		return
	}

	// Insert PlanItems
	for _, item := range input.Items {
		item.PlanID = input.Plan.ID
		_, err = tx.NewInsert().Model(&item).Exec(ctx)
		if err != nil {
			http.Error(w, "Failed to insert plan item", http.StatusInternalServerError)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "Plan created successfully",
		"planID":  input.Plan.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
