package handlers

import (
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/roundtown-app/roundtown-api/internal/middleware"
)

func (h *Handlers) forwardRequest(w http.ResponseWriter, url string) {
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Error fetching recommendations", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *Handlers) handleGetFeedRec(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	url := h.RecommendationServerURL + "/feed-recs/" + userID.String()

	h.forwardRequest(w, url)
}

func (h *Handlers) handleGetPlanRec(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	url := h.RecommendationServerURL + "/plan-recs/" + userID.String()

	h.forwardRequest(w, url)
}

func (h *Handlers) handleGetEBRec(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	url := h.RecommendationServerURL + "/event-based-recs/" + userID.String()

	h.forwardRequest(w, url)
}
