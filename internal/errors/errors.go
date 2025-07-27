package errors

import "fmt"

type NotificationError struct {
	Code    string
	Message string
}

func (e *NotificationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code, message string) error {
	return &NotificationError{Code: code, Message: message}
}
