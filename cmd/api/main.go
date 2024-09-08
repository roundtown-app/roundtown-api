package main

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/roundtown-app/roundtown-api/internal/handlers"
)

func main() {
	var router *chi.Mux = chi.NewRouter()

	handlers.Handler(router)

	slog.Info("Starting Roundtown API")

	err := http.ListenAndServe("localhost:8080", router)

	slog.Error(err.Error())
}
