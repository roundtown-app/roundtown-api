package handlers

import (
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"

	"github.com/roundtown-app/roundtown-api/internal/middleware"
)

func (h *Handlers) Handler(router *chi.Mux) {
	router.Use(chi_middleware.StripSlashes)

	router.Get("/heartbeat", h.handleHeartbeat)

	router.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Use(middleware.Authorization(h.AuthClient, h.DB.DB))

		apiRouter.Get("/heartbeat", h.handleHeartbeat)

		apiRouter.Route("/events", func(eventRouter chi.Router) {
			eventRouter.Get("/{eventID}", h.handleGetEvent)
			eventRouter.Post("/get-events", h.handleBulkGetEvent)
			eventRouter.Put("/{eventID}", h.handlePutEvent)
			eventRouter.Delete("/{eventID}", h.handleDeleteEvent)
			eventRouter.Post("/search", h.handleEventSearch)
			eventRouter.Post("/", h.handlePostEvent)
		})

		apiRouter.Route("/venues", func(venueRouter chi.Router) {
			venueRouter.Get("/{venueID}", h.handleGetVenue)
			venueRouter.Post("/get-venues", h.handleBulkGetVenue)
			venueRouter.Put("/{venueID}", h.handlePutVenue)
			venueRouter.Delete("/{venueID}", h.handleDeleteVenue)
			venueRouter.Post("/search", h.handleVenueSearch)
			venueRouter.Post("/", h.handlePostVenue)
		})

		apiRouter.Route("/plans", func(planRouter chi.Router) {
			planRouter.Get("/{planID}", h.handleGetPlan)
			planRouter.Post("/get-plans", h.handleBulkGetPlan)
			planRouter.Put("/{planID}", h.handlePutPlan)
			planRouter.Delete("/{planID}", h.handleDeletePlan)
			planRouter.Post("/search", h.handlePlanSearch)
			planRouter.Post("/", h.handlePostPlan)
		})

		apiRouter.Route("/users", func(userRouter chi.Router) {
			userRouter.Get("/self", h.handleGetSelf)
			userRouter.Get("/{userID}", h.handleGetUser)
			userRouter.Put("/", h.handlePutUser)
			userRouter.Put("/updateLocation", h.handlePutUserLocation)
			userRouter.Delete("/", h.handleDeleteUser)
			userRouter.Post("/search", h.handleUserSearch)
			// userRouter.Post("/", h.handlePostUser)
		})

		apiRouter.Route("/recommendations", func(recRouter chi.Router) {
			recRouter.Post("/feed-recs/{userID}", h.handleGetFeedRec)
			//recRouter.Get("/plan-recs/{userID}", h.handleGetPlanRec)
			//recRouter.Get("/event-based-recs/{userID}", h.handleGetEBRec)
		})

		apiRouter.Route("/interactions", func(intRouter chi.Router) {
			intRouter.Get("/savedItems", h.handleGetSavedItems)
			intRouter.Put("/saveItem", h.handleSaveItem)
			intRouter.Post("/unsaveItem", h.handleUnsaveItem)

			intRouter.Get("/subscribedItems", h.handleGetSubscribedItems)
			intRouter.Put("/subscribeItem", h.handleSubscribeItem)
			intRouter.Post("/unsubscribeItem", h.handleUnsubscribeItem)

			intRouter.Get("/visitedItems", h.handleGetVisitedItems)
			intRouter.Put("/visitItem", h.handleVisitItem)

			intRouter.Get("/sharedItems", h.handleGetSharedItems)
			intRouter.Put("/shareItem", h.handleShareItem)
		})
	})
}
