package controllers

import (
	"testing"

	"github.com/google/uuid"
)

func TestGenerateRandomUUID(t *testing.T) {
	first := generateRandomUUID()
	second := generateRandomUUID()
	if _, err := uuid.Parse(first); err != nil {
		t.Fatalf("generated invalid UUID %q: %v", first, err)
	}
	if first == second {
		t.Fatalf("generated duplicate UUID %q", first)
	}
}
