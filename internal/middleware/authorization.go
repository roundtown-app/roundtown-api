package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/uptrace/bun"
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

			// Check if the session is active (you may want to implement additional checks)
			if !token.Claims["session_active"].(bool) {
				http.Error(w, "Inactive session", http.StatusUnauthorized)
				return
			}

			// Get the Firebase UID from the token
			firebaseUID := token.UID

			// Query the database to get userID and accountType
			var userID string
			var accountType string
			err = db.QueryRowContext(r.Context(), "SELECT user_id, account_type FROM users WHERE firebase_uid = ?", firebaseUID).Scan(&userID, &accountType)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "User not found", http.StatusUnauthorized)
				} else {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
				}
				return
			}

			// Add userID and accountType to the request context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, AccountTypeKey, accountType)

			// Call the next handler with the updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
