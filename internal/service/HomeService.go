package service

import (
	"context"
	"fmt"
	"github.com/mostafijurj/notification-service/internal/logger"
	"math/rand"
)

func IndexPageRequest(ctx context.Context, name string) string {
	logger.InfoWithReqID(ctx, "Processing request in service layer")
	// Simulate work
	return fmt.Sprintf("Hello, %s!", name)
}

func GenerateRandomText(ctx context.Context, i int) string {
	logger.InfoWithReqID(ctx, "Generating random text")
	// Simulate generating random text
	randomText := make([]byte, i)
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for idx := range randomText {
		randomText[idx] = letters[rand.Intn(len(letters))]
	}
	return string(randomText)
}
