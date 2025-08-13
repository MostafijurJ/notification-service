package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mostafijurj/notification-service/internal/repository"
	"github.com/mostafijurj/notification-service/internal/service"
)

type Dependencies struct {
	Repo *repository.Repository
	Svc  *service.NotificationService
}

type sendNotificationReq struct {
	UserID      int64             `json:"user_id"`
	TypeKey     string            `json:"type_key"`
	Channels    []string          `json:"channels"`
	Payload     map[string]any    `json:"payload"`
	Priority    string            `json:"priority"`
	ScheduledAt *string           `json:"scheduled_at"`
}

type simpleResponse struct {
	ResponseCode    int         `json:"responseCode"`
	ResponseMessage string      `json:"responseMessage"`
	Data            interface{} `json:"data"`
}

func SendNotification(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		setHeaderValues(w)
		var req sendNotificationReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid payload", nil})
			return
		}
		idem := r.Header.Get("Idempotency-Key")
		var idemPtr *string
		if idem != "" { idemPtr = &idem }

		ids, err := dep.Svc.Enqueue(ctx, service.NotificationRequest{
			UserID: req.UserID, TypeKey: req.TypeKey, Channels: req.Channels, Payload: req.Payload, Priority: req.Priority, ScheduledAt: req.ScheduledAt,
		}, idemPtr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusInternalServerError, "failed to enqueue", nil})
			return
		}
		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "enqueued", ids})
	}
}

func UpsertPreference(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		setHeaderValues(w)
		userIDStr := mux.Vars(r)["user_id"]
		userID, _ := strconv.ParseInt(userIDStr, 10, 64)
		var body struct{ TypeKey, Channel string; OptedIn bool }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil { w.WriteHeader(http.StatusBadRequest); _ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid payload", nil}); return }
		if err := dep.Repo.UpsertPreference(ctx, repository.PreferenceUpsert{UserID: userID, TypeKey: body.TypeKey, Channel: body.Channel, OptedIn: body.OptedIn}); err != nil {
			w.WriteHeader(http.StatusInternalServerError); _ = json.NewEncoder(w).Encode(simpleResponse{http.StatusInternalServerError, "failed", nil}); return
		}
		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "ok", nil})
	}
}

func UpsertDND(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		setHeaderValues(w)
		userIDStr := mux.Vars(r)["user_id"]
		userID, _ := strconv.ParseInt(userIDStr, 10, 64)
		var body struct{ StartTime, EndTime, Timezone string }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil { w.WriteHeader(http.StatusBadRequest); _ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid payload", nil}); return }
		if err := dep.Repo.UpsertDND(ctx, userID, body.StartTime, body.EndTime, body.Timezone); err != nil { w.WriteHeader(http.StatusInternalServerError); _ = json.NewEncoder(w).Encode(simpleResponse{http.StatusInternalServerError, "failed", nil}); return }
		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "ok", nil})
	}
}

func CreateGroup(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context(); setHeaderValues(w)
		var body struct{ Name string; Description *string }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil { w.WriteHeader(http.StatusBadRequest); _ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid payload", nil}); return }
		id, err := dep.Repo.CreateGroup(ctx, repository.GroupCreate{Name: body.Name, Description: body.Description})
		if err != nil { w.WriteHeader(http.StatusInternalServerError); _ = json.NewEncoder(w).Encode(simpleResponse{http.StatusInternalServerError, "failed", nil}); return }
		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "created", map[string]int64{"group_id": id}})
	}
}

func AddGroupMember(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context(); setHeaderValues(w)
		groupIDStr := mux.Vars(r)["group_id"]; userIDStr := mux.Vars(r)["user_id"]
		groupID, _ := strconv.ParseInt(groupIDStr, 10, 64); userID, _ := strconv.ParseInt(userIDStr, 10, 64)
		if err := dep.Repo.AddGroupMember(ctx, groupID, userID); err != nil { w.WriteHeader(http.StatusInternalServerError); _ = json.NewEncoder(w).Encode(simpleResponse{http.StatusInternalServerError, "failed", nil}); return }
		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "ok", nil})
	}
}

func CreateCampaign(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context(); setHeaderValues(w)
		var body struct{ Name, TypeKey, Channel string; SegmentGroupID *int64; ScheduledAt *string; Priority string }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil { w.WriteHeader(http.StatusBadRequest); _ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid payload", nil}); return }
		id, err := dep.Repo.CreateCampaign(ctx, repository.CampaignCreate{Name: body.Name, TypeKey: body.TypeKey, Channel: body.Channel, SegmentGroupID: body.SegmentGroupID, ScheduledAt: body.ScheduledAt, Priority: body.Priority})
		if err != nil { w.WriteHeader(http.StatusInternalServerError); _ = json.NewEncoder(w).Encode(simpleResponse{http.StatusInternalServerError, "failed", nil}); return }
		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "created", map[string]int64{"campaign_id": id})
	}
}

func ListInApp(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context(); setHeaderValues(w)
		userIDStr := mux.Vars(r)["user_id"]; userID, _ := strconv.ParseInt(userIDStr, 10, 64)
		items, err := dep.Repo.ListInApp(ctx, userID, false, 50)
		if err != nil { w.WriteHeader(http.StatusInternalServerError); _ = json.NewEncoder(w).Encode(simpleResponse{http.StatusInternalServerError, "failed", nil}); return }
		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "ok", items})
	}
}

func MarkInApp(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context(); setHeaderValues(w)
		idStr := mux.Vars(r)["id"]; readStr := r.URL.Query().Get("read")
		id, _ := strconv.ParseInt(idStr, 10, 64); read := readStr == "true"
		if err := dep.Repo.MarkInAppRead(ctx, id, read); err != nil { w.WriteHeader(http.StatusInternalServerError); _ = json.NewEncoder(w).Encode(simpleResponse{http.StatusInternalServerError, "failed", nil}); return }
		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "ok", nil})
	}
}

func setHeaderValues(writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}