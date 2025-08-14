package routes

import (
	"github.com/gorilla/mux"
	"github.com/mostafijurj/notification-service/internal/controller"
)

// V1Routes sets up version 1 API routes
func V1Routes(router *mux.Router, deps *controller.Dependencies) {
	// Notification endpoints
	router.HandleFunc("/api/v1/notifications/send", controller.SendNotification(deps)).Methods("POST")
	router.HandleFunc("/api/v1/notifications/{id}", controller.GetNotification(deps)).Methods("GET")
	router.HandleFunc("/api/v1/notifications/{id}/status", controller.UpdateNotificationStatus(deps)).Methods("PUT")

	// Preference endpoints
	router.HandleFunc("/api/v1/preferences", controller.UpsertPreference(deps)).Methods("POST")
	router.HandleFunc("/api/v1/preferences/{userID}", controller.GetUserPreferences(deps)).Methods("GET")

	// DND endpoints
	router.HandleFunc("/api/v1/dnd", controller.UpsertDND(deps)).Methods("POST")
	router.HandleFunc("/api/v1/dnd/{userID}", controller.GetUserDND(deps)).Methods("GET")

	// In-app notification endpoints
	router.HandleFunc("/api/v1/inapp", controller.CreateInAppNotification(deps)).Methods("POST")
	router.HandleFunc("/api/v1/inapp/{userID}", controller.ListInAppNotifications(deps)).Methods("GET")
	router.HandleFunc("/api/v1/inapp/{id}/read", controller.MarkInAppRead(deps)).Methods("PUT")

	// Campaign endpoints
	router.HandleFunc("/api/v1/campaigns", controller.CreateCampaign(deps)).Methods("POST")
	router.HandleFunc("/api/v1/campaigns/{id}", controller.GetCampaign(deps)).Methods("GET")
	router.HandleFunc("/api/v1/campaigns/{id}/send", controller.SendCampaign(deps)).Methods("POST")
}
