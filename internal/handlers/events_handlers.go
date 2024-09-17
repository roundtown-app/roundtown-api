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

func handleGetEvent(w http.ResponseWriter, r *http.Request) {
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

	var eventDetails *tools.EventDetails = (*database).GetEvent(chi.URLParam(r, "eventID"))
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

func handlePutEvent(w http.ResponseWriter, r *http.Request) {
	err, shouldReturn := newFunction(r, w)
	if shouldReturn {
		return
	}

	var database *tools.DatabaseInterface
	database, err = tools.NewDatabase()
	if err != nil {
		api.InternalErrorHandler(w)
		return
	}

	var eventDetails *tools.EventDetails = (*database).PutEvents(chi.URLParam(r, "eventID"))
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

func decodeParams[T *interface{}](w http.ResponseWriter, r *http.Request) (T, bool) {
	var params T = &T{}
	var decoder *schema.Decoder = schema.NewDecoder()
	var err error

	err = decoder.Decode(params, r.URL.Query())

	if err != nil {
		slog.Error(err.Error())
		api.InternalErrorHandler(w)
		return nil, true
	}

	return params, false
}
