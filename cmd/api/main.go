package main

import (
	"context"
	"log/slog"
	"net/http"

	firebase "firebase.google.com/go"
	"github.com/go-chi/chi/v5"
	"google.golang.org/api/option"

	"github.com/roundtown-app/roundtown-api/db"
	"github.com/roundtown-app/roundtown-api/internal/handlers"
)

func main() {
	slog.Info("Starting Roundtown API")

	opt := option.WithCredentialsFile("credentials.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		slog.Error("Error initializing Firebase app: " + err.Error() + "\n")
	}
    
	client, err := app.Auth(context.Background())
	if err != nil {
		slog.Error("Error creating Firebase client: " + err.Error() + "\n")
	}

	recServerURL := ""

	dsn := "postgres://username:password@localhost:5432/database_name?sslmode=disable"
	database, db_err := db.NewDB(dsn)
	if db_err != nil {
		slog.Error("Failed to initialize database: " + db_err.Error() + "\n")
	}
	defer database.Close()

	var router *chi.Mux = chi.NewRouter()

	h := handlers.NewHandler(database, recServerURL, client)
	h.Handler(router)

	slog.Info("Started Roundtown API")

	err = http.ListenAndServe("localhost:8080", router)

	slog.Error(err.Error())
}
