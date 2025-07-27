package controller

import (
	"encoding/json"
	"github.com/mostafijurj/notification-service/internal/logger"
	"github.com/mostafijurj/notification-service/internal/service"
	"net/http"
)

func IndexPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.InfoWithReqID(ctx, "IndexPage")
	message := service.IndexPageRequest(ctx, "Mostafijur")
	setHeaderValues(w)
	response := JSONResponse[string]{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Success",
		Data:            message,
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
	setHeaderValues(w)
}

func GenerateRandomText(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.InfoWithReqID(ctx, "GenerateRandomText")
	// Simulate generating random text
	randomText := service.GenerateRandomText(ctx, 1001) // Assuming this function exists in the service layer
	setHeaderValues(w)
	response := JSONResponse[string]{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Random text generated successfully",
		Data:            randomText,
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
	setHeaderValues(w)

}

// JSONResponse Generic JSON response struct for consistent API responses
type JSONResponse[T any] struct {
	ResponseCode    int    `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	Data            T      `json:"data"`
}

func setHeaderValues(writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}
