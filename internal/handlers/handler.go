package handlers

import (
	"firebase.google.com/go/auth"
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"

	"github.com/roundtown-app/roundtown-api/db"
	"github.com/roundtown-app/roundtown-api/internal/middleware"
)

type Handlers struct {
	DB                      *db.DB
	RecommendationServerURL string
	AuthClient              *auth.Client
}

func NewHandler(db *db.DB, recServerURL string, client *auth.Client) *Handlers {
	return &Handlers{
		DB:                      db,
		RecommendationServerURL: recServerURL,
		AuthClient:              client,
	}
}

func (h *Handlers) Handler(router *chi.Mux) {
	router.Use(chi_middleware.StripSlashes)

	router.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Use(middleware.Authorization(h.AuthClient, h.DB.DB))

		apiRouter.Route("/events", func(eventRouter chi.Router) {
			eventRouter.Get("/{eventID}", h.handleGetEvent)
			eventRouter.Put("/{eventID}", h.handlePutEvent)
			eventRouter.Delete("/{eventID}", h.handleDeleteEvent)
			eventRouter.Post("/search", h.handleEventSearch)
			eventRouter.Post("/", h.handlePostEvent)
		})

		apiRouter.Route("/venues", func(venueRouter chi.Router) {
			venueRouter.Get("/{venueID}", h.handleGetVenue)
			venueRouter.Put("/{venueID}", h.handlePutVenue)
			venueRouter.Delete("/{venueID}", h.handleDeleteVenue)
			venueRouter.Post("/search", h.handleVenueSearch)
			venueRouter.Post("/", h.handlePostVenue)
		})

		apiRouter.Route("/plans", func(planRouter chi.Router) {
			planRouter.Get("/{planID}", h.handleGetPlan)
			planRouter.Put("/{planID}", h.handlePutPlan)
			planRouter.Delete("/{planID}", h.handleDeletePlan)
			planRouter.Post("/search", h.handlePlanSearch)
			planRouter.Post("/", h.handlePostPlan)
		})

		apiRouter.Route("/users", func(userRouter chi.Router) {
			userRouter.Get("/{userID}", h.handleGetUser)
			userRouter.Put("/", h.handlePutUser)
			userRouter.Put("/updateLocation", h.handlePutUserLocation)
			userRouter.Delete("/", h.handleDeleteUser)
			userRouter.Put("/search", h.handleUserSearch)
			userRouter.Post("/", h.handlePostUser)
		})

		apiRouter.Route("/recommendations", func(recRouter chi.Router) {
			recRouter.Get("/feed-recs/{userID}", h.handleGetFeedRec)
			recRouter.Get("/plan-recs/{userID}", h.handleGetPlanRec)
			recRouter.Get("/event-based-recs/{userID}", h.handleGetEBRec)
		})
	})
}
