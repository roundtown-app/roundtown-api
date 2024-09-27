package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"

	"github.com/roundtown-app/roundtown-api/api"
	"github.com/roundtown-app/roundtown-api/internal/tools"
)

func (h *Handlers) handleGetUser(w http.ResponseWriter, r *http.Request) {
	var params = api.EventParams{}
	var decoder *schema.Decoder = schema.NewDecoder()
	var err error

	err = decoder.Decode(&params, r.URL.Query())

	if err != nil {
		slog.Error(err.Error())
		api.InternalErrorHandler(w)
		return
	}

	var database *tools.DatabaseInterface
	database, err = tools.NewDatabase()
	if err != nil {
		api.InternalErrorHandler(w)
		return
	}

	var eventDetails *tools.EventDetails = (*database).GetUser(chi.URLParam(r, "userID"))
	if eventDetails == nil {
		slog.Error("Could not retrieve event")
		api.InternalErrorHandler(w)
		return
	}

	var response = api.EventResponse{
		Name: (*eventDetails).Name,
		Code: http.StatusOK,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.Error(err.Error())
		api.InternalErrorHandler(w)
		return
	}
}

func (h *Handlers) handlePutUser(w http.ResponseWriter, r *http.Request) {
	var params = api.EventParams{}
	var decoder *schema.Decoder = schema.NewDecoder()
	var err error

	err = decoder.Decode(&params, r.URL.Query())

	if err != nil {
		slog.Error(err.Error())
		api.InternalErrorHandler(w)
		return
	}

	var database *tools.DatabaseInterface
	database, err = tools.NewDatabase()
	if err != nil {
		api.InternalErrorHandler(w)
		return
	}

	var eventDetails *tools.EventDetails = (*database).PutUser(chi.URLParam(r, "userID"))
	if eventDetails == nil {
		slog.Error("Could not update event")
		api.InternalErrorHandler(w)
		return
	}

	var response = api.EventResponse{
		Code: http.StatusOK,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.Error(err.Error())
		api.InternalErrorHandler(w)
		return
	}
}

func (h *Handlers) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	var params = api.EventParams{}
	var decoder *schema.Decoder = schema.NewDecoder()
	var err error

	err = decoder.Decode(&params, r.URL.Query())

	if err != nil {
		slog.Error(err.Error())
		api.InternalErrorHandler(w)
		return
	}

	var database *tools.DatabaseInterface
	database, err = tools.NewDatabase()
	if err != nil {
		api.InternalErrorHandler(w)
		return
	}

	var eventDetails *tools.EventDetails = (*database).DeleteUser(chi.URLParam(r, "userID"))
	if eventDetails == nil {
		slog.Error("Could not update event")
		api.InternalErrorHandler(w)
		return
	}

	var response = api.EventResponse{
		Code: http.StatusOK,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.Error(err.Error())
		api.InternalErrorHandler(w)
		return
	}
}

func (h *Handlers) handlePostUser(w http.ResponseWriter, r *http.Request) {
	var params = api.EventParams{}
	var decoder *schema.Decoder = schema.NewDecoder()
	var err error

	err = decoder.Decode(&params, r.URL.Query())

	if err != nil {
		slog.Error(err.Error())
		api.InternalErrorHandler(w)
		return
	}

	var database *tools.DatabaseInterface
	database, err = tools.NewDatabase()
	if err != nil {
		api.InternalErrorHandler(w)
		return
	}

	var eventDetails *tools.EventDetails = (*database).PostUser()
	if eventDetails == nil {
		slog.Error("Could not update event")
		api.InternalErrorHandler(w)
		return
	}

	var response = api.EventResponse{
		Code: http.StatusOK,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.Error(err.Error())
		api.InternalErrorHandler(w)
		return
	}
}
