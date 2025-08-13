package kafka

import "testing"

func TestTopicsConstants(t *testing.T) {
	if TopicEnqueued == "" || TopicReadyEmailHigh == "" || TopicReadyInAppLow == "" {
		t.Fatal("topics should not be empty")
	}
}