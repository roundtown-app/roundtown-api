package handlers

import (
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"

	"github.com/roundtown-app/roundtown-api/db"
	"github.com/roundtown-app/roundtown-api/internal/middleware"
)

type Handlers struct {
	DB *db.DB
}

func NewHandler(db *db.DB) *Handlers {
	return &Handlers{DB: db}
}

func (h *Handlers) Handler(router *chi.Mux) {
	router.Use(chi_middleware.StripSlashes)

	router.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Use(middleware.Authorization)

		apiRouter.Route("/events", func(eventRouter chi.Router) {
			eventRouter.Get("/{eventID}", h.handleGetEvent)
			eventRouter.Put("/{eventID}", h.handlePutEvent)
			eventRouter.Delete("/{eventID}", h.handleDeleteEvent)
			eventRouter.Post("/", h.handlePostEvent)
		})

		apiRouter.Route("/venues", func(venueRouter chi.Router) {
			venueRouter.Get("/{venueID}", h.handleGetVenue)
			venueRouter.Put("/{venueID}", h.handlePutVenue)
			venueRouter.Delete("/{venueID}", h.handleDeleteVenue)
			venueRouter.Post("/", h.handlePostVenue)
		})

		apiRouter.Route("/plans", func(planRouter chi.Router) {
			planRouter.Get("/{planID}", h.handleGetPlan)
			planRouter.Put("/{planID}", h.handlePutPlan)
			planRouter.Delete("/{planID}", h.handleDeletePlan)
			planRouter.Post("/", h.handlePostPlan)
		})

		apiRouter.Route("/users", func(userRouter chi.Router) {
			userRouter.Get("/{userID}", h.handleGetUser)
			userRouter.Put("/{userID}", h.handlePutUser)
			userRouter.Delete("/{userID}", h.handleDeleteUser)
			userRouter.Post("/", h.handlePostUser)
		})

		apiRouter.Route("/images", func(imageRouter chi.Router) {
			imageRouter.Post("/", h.handlePostImage)
		})

		apiRouter.Route("/recommendations", func(recRouter chi.Router) {
			recRouter.Get("feed-recs/{userID}", h.handleGetFeedRec)
			recRouter.Get("plan-recs/{userID}", h.handleGetPlanRec)
			recRouter.Get("event-based-recs/{userID}", h.handleGetEBRec)
		})
	})
}
