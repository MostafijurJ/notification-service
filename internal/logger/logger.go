package logger

import (
	"context"
	"fmt"
	"github.com/mostafijurj/notification-service/internal/middleware"
	"log"
	"os"
)

var (
	Info  *log.Logger
	Error *log.Logger
)

func Init() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	Info = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lmicroseconds)
	Error = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lmicroseconds)
}

func InfoWithReqID(ctx context.Context, msg string) {
	reqID := middleware.GetRequestID(ctx)
	Info.Println(fmt.Sprintf("[requestId=%s] %s", reqID, msg))
}

func ErrorWithReqID(ctx context.Context, msg string) {
	reqID := middleware.GetRequestID(ctx)
	Error.Println(fmt.Sprintf("[requestId=%s] %s", reqID, msg))
}
