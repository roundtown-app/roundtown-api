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
			eventRouter.Put("/{eventID}", handlePutEvent)
			eventRouter.Delete("/{eventID}", handlePostEvent)
			eventRouter.Post("/", handlePostEvent)
		})

		apiRouter.Route("/venues", func(venueRouter chi.Router) {
			venueRouter.Get("/{venueID}", handleGetVenue)
			venueRouter.Put("/{venueID}", handlePutVenue)
			venueRouter.Delete("/{venueID}", handlePostVenue)
			venueRouter.Post("/", handlePostVenue)
		})

		apiRouter.Route("/plans", func(planRouter chi.Router) {
			planRouter.Get("/{planID}", handleGetPlan)
			planRouter.Put("/{planID}", handlePutPlan)
			planRouter.Delete("/{planID}", handlePostPlan)
			planRouter.Post("/", handlePostPlan)
		})

		apiRouter.Route("/users", func(userRouter chi.Router) {
			userRouter.Get("/{userID}", handleGetUser)
			userRouter.Put("/{userID}", handlePutUser)
			userRouter.Delete("/{userID}", handlePostUser)
			userRouter.Post("/", handlePostUser)
		})

		apiRouter.Route("/images", func(imageRouter chi.Router) {
			imageRouter.Post("/", handlePostImage)
		})

		apiRouter.Route("/recommendations", func(recRouter chi.Router) {
			recRouter.Get("feed-recs/{userID}", handleGetFeedRec)
			recRouter.Get("plan-recs/{userID}", handleGetPlanRec)
			recRouter.Get("event-based-recs/{userID}", handleGetEBRec)
		})
	})
}
