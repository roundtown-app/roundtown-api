package handlers

import (
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"

	"github.com/roundtown-app/roundtown-api/internal/middleware"
)

func Handler(router *chi.Mux) {
	router.Use(chi_middleware.StripSlashes)

	router.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Use(middleware.Authorization)

		apiRouter.Route("/events", func(eventRouter chi.Router) {
			eventRouter.Get("/{eventID}", handleGetEvent)
		})
	})
}
