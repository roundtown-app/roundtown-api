package handlers

import (
	"io"
	"net/http"
)

func (h *Handlers) forwardRequest(w http.ResponseWriter, r *http.Request, url string) {
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
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	url := h.RecommendationServerURL + "/feed-recs/" + userID

	h.forwardRequest(w, r, url)
}

func (h *Handlers) handleGetPlanRec(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	url := h.RecommendationServerURL + "/plan-recs/" + userID

	h.forwardRequest(w, r, url)
}

func (h *Handlers) handleGetEBRec(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	url := h.RecommendationServerURL + "/event-based-recs/" + userID

	h.forwardRequest(w, r, url)
}
