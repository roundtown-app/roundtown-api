package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/roundtown-app/roundtown-api/api"
	"github.com/roundtown-app/roundtown-api/internal/middleware"
)

// ----- PUT helper struct -----

type UserUpdate struct {
	Username *string `json:"username"`
}

// ----- Search helper struct -----

type UserSearchCriteria struct {
	FirebaseUID *string `json:"firebase_uid,omitempty"`
	Username    *string `json:"username,omitempty"`
	AccountType *string `json:"account_type,omitempty"`
}

// ----- Request handlers -----

func (h *Handlers) handleGetSelf(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	user := new(api.User)
	err := h.DB.NewSelect().Model(user).Where("user_id = ?", userID).Scan(r.Context())
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *Handlers) handleGetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user := new(api.User)
	err = h.DB.NewSelect().Model(user).Where("user_id = ?", userID).Scan(r.Context())
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *Handlers) handlePutUser(w http.ResponseWriter, r *http.Request) {
	var user UserUpdate
	ctx := r.Context()

	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err = h.DB.NewUpdate().Model(&user).Where("user_id = ?", userID).Exec(ctx)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) handlePutUserLocation(w http.ResponseWriter, r *http.Request) {
	var location api.UserLocations

	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&location)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	location.UserID = userID

	location.LastUpdated = time.Now()

	_, err = h.DB.NewInsert().Model(&location).
		On("CONFLICT (user_id) DO UPDATE").
		Set("longitude = EXCLUDED.longitude").
		Set("latitude = EXCLUDED.latitude").
		Set("last_updated = EXCLUDED.last_updated").
		Exec(r.Context())

	if err != nil {
		http.Error(w, "Failed to update user location", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	var user api.User

	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user.UserID = userID

	_, err = h.DB.NewDelete().Model(&user).Where("user_id = ?", user.UserID).Exec(r.Context())
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) handleUserSearch(w http.ResponseWriter, r *http.Request) {
	// Parse JSON from request body
	var criteria UserSearchCriteria
	err := json.NewDecoder(r.Body).Decode(&criteria)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start building the query
	query := h.DB.NewSelect().Model((*api.User)(nil))

	// Add conditions based on provided parameters
	if criteria.FirebaseUID != nil {
		query = query.Where("firebase_uid = ?", *criteria.FirebaseUID)
	}
	if criteria.Username != nil {
		query = query.Where("username LIKE ?", "%"+*criteria.Username+"%")
	}
	if criteria.AccountType != nil {
		query = query.Where("account_type = ?", *criteria.AccountType)
	}

	// Execute the query
	var users []api.User
	err = query.Limit(10).Scan(r.Context(), &users)
	if err != nil {
		http.Error(w, "Failed to search users", http.StatusInternalServerError)
		return
	}

	// Return the results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// func (h *Handlers) handlePostUser(w http.ResponseWriter, r *http.Request) {
// 	var user api.User
// 	err := json.NewDecoder(r.Body).Decode(&user)
// 	if err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	user.UserID = uuid.New()
// 	user.AccountCreated = time.Now()
// 	user.AccountType = "user"

// 	_, err = h.DB.NewInsert().Model(&user).Exec(r.Context())
// 	if err != nil {
// 		http.Error(w, "Failed to create user", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(user)
// }
