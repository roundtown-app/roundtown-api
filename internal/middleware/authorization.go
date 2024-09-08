package middleware

import (
	"fmt"
	"net/http"

	"github.com/roundtown-app/roundtown-api/api"
	"github.com/roundtown-app/roundtown-api/internal/tools"
)

func Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var token = r.Header.Get("Authorization")
		var err error

		var database *tools.DatabaseInterface
		database, err = tools.NewDatabase()
		if err != nil {
			api.InternalErrorHandler(w)
			return
		}

		var loginDetails *tools.LoginDetails = (*database).GetUserLoginDetails(token)

		if loginDetails == nil || (token != (*loginDetails).AuthToken) {
			api.RequestErrorHandler(w, fmt.Errorf("invalid token"))
			return
		}

		next.ServeHTTP(w, r)

	})
}
