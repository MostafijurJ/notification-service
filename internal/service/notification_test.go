package service

import "testing"

func TestStatusFromSchedule_Basic(t *testing.T) {
	if statusFromSchedule(nil) != "enqueued" {
		t.Fatalf("expected enqueued")
	}
	s := "2024-01-01T00:00:00Z"
	if statusFromSchedule(&s) != "scheduled" {
		t.Fatalf("expected scheduled")
	}
}