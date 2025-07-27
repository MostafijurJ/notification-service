package routes

import (
	"github.com/gorilla/mux"
	"github.com/mostafijurj/notification-service/internal/controller"
)

var HomeRoutes = func(router *mux.Router) {

	router.HandleFunc("/", controller.IndexPage).Methods("GET")
	router.HandleFunc("/tex", controller.GenerateRandomText).Methods("GET")
}
