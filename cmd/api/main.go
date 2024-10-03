package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"github.com/go-chi/chi/v5"
	"google.golang.org/api/option"

	"github.com/roundtown-app/roundtown-api/db"
	"github.com/roundtown-app/roundtown-api/internal/handlers"
)

func main() {
	slog.Info("Starting Roundtown API")

	// Initialize the app with the service account
	opt := option.WithCredentialsFile("./service_account.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		slog.Error("error initializing firebase app: " + err.Error())
		os.Exit(1)
	}

	slog.Info("Initialized Firebase app")

	client, err := app.Auth(context.Background())
	if err != nil {
		slog.Error("error creating firebase auth client: " + err.Error())
		os.Exit(1)
	}

	slog.Info("Initialized Firebase client")

	recServerURL := ""

	dsn := "postgres://postgres:password1@localhost:5432/postgres?sslmode=disable"
	database, db_err := db.NewDB(dsn)
	if db_err != nil {
		slog.Error("Failed to initialize database: " + db_err.Error() + "\n")
		os.Exit(1)
	}
	defer database.Close()

	slog.Info("Initialized connection to database")

	var router *chi.Mux = chi.NewRouter()

	h := handlers.NewHandler(database, recServerURL, client)
	h.Handler(router)

	slog.Info("Started Roundtown API")

	err = http.ListenAndServe("localhost:8080", router)

	slog.Error(err.Error())
}
