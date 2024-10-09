package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"

	"github.com/roundtown-app/roundtown-api/db"
	"github.com/roundtown-app/roundtown-api/internal/handlers"
)

func main() {
	slog.Info("Starting Roundtown API")

	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file")
	}

	dsn := os.Getenv("DATABASE_SOURCE_NAME")
	recServerURL := os.Getenv("RECOMMENDATION_SERVER_URL")

	slog.Info("Retrieved environment variables")

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
