package routes

//
//import (
//	"github.com/gorilla/mux"
//	"github.com/mostafijurj/notification-service/internal/controller"
//)
//
//func V1Routes(router *mux.Router, dep *controller.Dependencies) {
//	r := router.PathPrefix("/v1").Subrouter()
//	// notifications
//	r.HandleFunc("/notifications", controller.SendNotification(dep)).Methods("POST")
//	// preferences and DND
//	r.HandleFunc("/preferences/{user_id}", controller.UpsertPreference(dep)).Methods("POST")
//	r.HandleFunc("/dnd/{user_id}", controller.UpsertDND(dep)).Methods("PUT")
//	// groups and campaigns
//	r.HandleFunc("/groups", controller.CreateGroup(dep)).Methods("POST")
//	r.HandleFunc("/groups/{group_id}/members/{user_id}", controller.AddGroupMember(dep)).Methods("POST")
//	r.HandleFunc("/campaigns", controller.CreateCampaign(dep)).Methods("POST")
//	// in-app
//	r.HandleFunc("/inapp/{user_id}", controller.ListInApp(dep)).Methods("GET")
//	r.HandleFunc("/inapp/{id}", controller.MarkInApp(dep)).Methods("PATCH")
//}
