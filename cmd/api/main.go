package main

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/roundtown-app/roundtown-api/db"
	"github.com/roundtown-app/roundtown-api/internal/handlers"
)

func main() {
	slog.Info("Starting Roundtown API")

	recServerURL := ""

	dsn := "postgres://username:password@localhost:5432/database_name?sslmode=disable"
	database, db_err := db.NewDB(dsn)
	if db_err != nil {
		slog.Error("Failed to initialize database: %v", db_err)
	}
	defer database.Close()

	var router *chi.Mux = chi.NewRouter()

	h := handlers.NewHandler(database, recServerURL)
	h.Handler(router)

	slog.Info("Started Roundtown API")

	err := http.ListenAndServe("localhost:8080", router)

	slog.Error(err.Error())
}
