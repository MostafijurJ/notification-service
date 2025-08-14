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
	UserID      int64          `json:"user_id"`
	TypeKey     string         `json:"type_key"`
	Channels    []string       `json:"channels"`
	Payload     map[string]any `json:"payload"`
	Priority    string         `json:"priority"`
	ScheduledAt *string        `json:"scheduled_at"`
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
		if idem != "" {
			idemPtr = &idem
		}

		ids, err := dep.Svc.Enqueue(ctx, service.NotificationRequest{
			UserID:      req.UserID,
			TypeKey:     req.TypeKey,
			Channels:    req.Channels,
			Payload:     req.Payload,
			Priority:    req.Priority,
			ScheduledAt: req.ScheduledAt,
		}, idemPtr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusInternalServerError, "failed to enqueue", nil})
			return
		}
		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "enqueued", ids})
	}
}

// GetNotification retrieves a notification by ID
func GetNotification(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		setHeaderValues(w)

		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid notification ID", nil})
			return
		}

		notification, err := dep.Repo.GetNotificationByID(ctx, id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusNotFound, "notification not found", nil})
			return
		}

		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "success", notification})
	}
}

// UpdateNotificationStatus updates the status of a notification
func UpdateNotificationStatus(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		setHeaderValues(w)

		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid notification ID", nil})
			return
		}

		var req struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid payload", nil})
			return
		}

		if err := dep.Repo.UpdateNotificationStatus(ctx, id, req.Status); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusInternalServerError, "failed to update status", nil})
			return
		}

		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "status updated", nil})
	}
}

// UpsertPreference creates or updates user channel preferences
func UpsertPreference(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		setHeaderValues(w)

		var req repository.PreferenceUpsert
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid payload", nil})
			return
		}

		if err := dep.Repo.UpsertPreference(ctx, req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusInternalServerError, "failed to update preference", nil})
			return
		}

		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "preference updated", nil})
	}
}

// GetUserPreferences retrieves all preferences for a user
func GetUserPreferences(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		setHeaderValues(w)

		vars := mux.Vars(r)
		userIDStr := vars["userID"]
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid user ID", nil})
			return
		}

		preferences, err := dep.Repo.GetUserPreferences(ctx, userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusInternalServerError, "failed to get preferences", nil})
			return
		}

		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "success", preferences})
	}
}

// UpsertDND creates or updates user DND settings
func UpsertDND(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		setHeaderValues(w)

		var req struct {
			UserID    int64  `json:"user_id"`
			StartTime string `json:"start_time"`
			EndTime   string `json:"end_time"`
			Timezone  string `json:"timezone"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid payload", nil})
			return
		}

		if err := dep.Repo.UpsertDND(ctx, req.UserID, req.StartTime, req.EndTime, req.Timezone); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusInternalServerError, "failed to update DND", nil})
			return
		}

		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "DND updated", nil})
	}
}

// GetUserDND retrieves DND settings for a user
func GetUserDND(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		setHeaderValues(w)

		vars := mux.Vars(r)
		userIDStr := vars["userID"]
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid user ID", nil})
			return
		}

		dndWindow, err := dep.Repo.GetUserDNDWindow(ctx, userID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusNotFound, "DND settings not found", nil})
			return
		}

		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "success", dndWindow})
	}
}

// CreateInAppNotification creates an in-app notification
func CreateInAppNotification(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		setHeaderValues(w)

		var req struct {
			UserID   int64          `json:"user_id"`
			TypeKey  string         `json:"type_key"`
			Title    string         `json:"title"`
			Body     string         `json:"body"`
			Metadata map[string]any `json:"metadata"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid payload", nil})
			return
		}

		metaJSON, _ := json.Marshal(req.Metadata)
		id, err := dep.Repo.CreateInApp(ctx, req.UserID, req.TypeKey, req.Title, req.Body, metaJSON)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusInternalServerError, "failed to create notification", nil})
			return
		}

		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "created", map[string]int64{"id": id}})
	}
}

// ListInAppNotifications retrieves in-app notifications for a user
func ListInAppNotifications(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		setHeaderValues(w)

		vars := mux.Vars(r)
		userIDStr := vars["userID"]
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid user ID", nil})
			return
		}

		onlyUnread := r.URL.Query().Get("unread") == "true"
		limit := 50 // default limit
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		notifications, err := dep.Repo.ListInApp(ctx, userID, onlyUnread, limit)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusInternalServerError, "failed to get notifications", nil})
			return
		}

		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "success", notifications})
	}
}

// MarkInAppRead marks an in-app notification as read
func MarkInAppRead(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		setHeaderValues(w)

		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid notification ID", nil})
			return
		}

		var req struct {
			Read bool `json:"read"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid payload", nil})
			return
		}

		if err := dep.Repo.MarkInAppRead(ctx, id, req.Read); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusInternalServerError, "failed to update read status", nil})
			return
		}

		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "read status updated", nil})
	}
}

// CreateCampaign creates a new campaign
func CreateCampaign(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeaderValues(w)

		var req struct {
			Name           string  `json:"name"`
			TypeKey        string  `json:"type_key"`
			Channel        string  `json:"channel"`
			SegmentGroupID *int64  `json:"segment_group_id"`
			ScheduledAt    *string `json:"scheduled_at"`
			Priority       string  `json:"priority"`
			CreatedBy      string  `json:"created_by"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid payload", nil})
			return
		}

		// TODO: Implement campaign creation logic
		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "campaign created", nil})
	}
}

// GetCampaign retrieves a campaign by ID
func GetCampaign(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeaderValues(w)

		vars := mux.Vars(r)
		idStr := vars["id"]
		_, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid campaign ID", nil})
			return
		}

		// TODO: Implement campaign retrieval logic
		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "success", map[string]int64{"id": 0}})
	}
}

// SendCampaign sends a campaign to all users in the segment
func SendCampaign(dep *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeaderValues(w)

		vars := mux.Vars(r)
		idStr := vars["id"]
		_, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusBadRequest, "invalid campaign ID", nil})
			return
		}

		// TODO: Implement campaign sending logic
		_ = json.NewEncoder(w).Encode(simpleResponse{http.StatusOK, "campaign sent", nil})
	}
}
