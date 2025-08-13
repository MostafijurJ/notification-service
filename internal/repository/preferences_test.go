package repository

import "testing"

func TestPreferenceUpsertStruct(t *testing.T) {
	_ = PreferenceUpsert{UserID: 1, TypeKey: "x", Channel: "inapp", OptedIn: true}
}