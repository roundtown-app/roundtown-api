package middleware

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/roundtown-app/roundtown-api/api"
)

type contextKey string

const (
	UserIDKey      contextKey = "userID"
	AccountTypeKey contextKey = "accountType"
)

func Authorization(firebaseAuth *auth.Client, db *bun.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the token from the Authorization header
			authHeader := r.Header.Get("Authorization")
			idToken := strings.TrimPrefix(authHeader, "Bearer ")
			if idToken == "" {
				http.Error(w, "Missing or invalid authorization token", http.StatusUnauthorized)
				return
			}

			// Verify the Firebase token
			token, err := firebaseAuth.VerifyIDToken(r.Context(), idToken)
			if err != nil {
				http.Error(w, "Invalid authorization token", http.StatusUnauthorized)
				return
			}

			// Check if the session is active
			// if !token.Claims["session_active"].(bool) {
			// 	http.Error(w, "Inactive session", http.StatusUnauthorized)
			// 	return
			// }

			// Get the Firebase UID from the token
			firebaseUID := token.UID

			// Query the database to get userID and accountType
			var user api.User
			err = db.NewSelect().
				Model(&user).
				Where("firebase_uid = ?", firebaseUID).
				Scan(r.Context())

			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					// User doesn't exist, create a new one
					user, err = createNewUser(r.Context(), db, firebaseAuth, firebaseUID)
					if err != nil {
						http.Error(w, "Failed to create user", http.StatusInternalServerError)
						return
					}
				} else {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
			}

			// Add userID and accountType to the request context
			ctx := context.WithValue(r.Context(), UserIDKey, user.UserID)
			ctx = context.WithValue(ctx, AccountTypeKey, user.AccountType)

			// Call the next handler with the updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func createNewUser(ctx context.Context, db *bun.DB, firebaseAuth *auth.Client, firebaseUID string) (api.User, error) {
	// Get user info from Firebase
	firebaseUser, err := firebaseAuth.GetUser(ctx, firebaseUID)
	if err != nil {
		return api.User{}, fmt.Errorf("failed to get Firebase user: %w", err)
	}

	// Generate UUID v7
	userID, err := uuid.NewV7()
	if err != nil {
		return api.User{}, fmt.Errorf("failed to generate UUID: %w", err)
	}

	// Ensure we have a username from Firebase
	// Commenting this code out to enable support for anon users
	// if firebaseUser.DisplayName == "" {
	// 	return api.User{}, fmt.Errorf("firebase user has no display name set")
	// }

	// Create new user
	user := api.User{
		UserID:      userID,
		FirebaseUID: firebaseUID,
		Username:    firebaseUser.DisplayName,
		AccountType: "user", // Default account type
	}

	// Insert the new user into the database
	_, err = db.NewInsert().
		Model(&user).
		Exec(ctx)

	if err != nil {
		return api.User{}, fmt.Errorf("failed to insert user: %w", err)
	}

	return user, nil
}
