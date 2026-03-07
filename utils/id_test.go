package utils

import (
	"testing"
)

func TestGenerateID_Length(t *testing.T) {
	id, err := GenerateID()
	if err != nil {
		t.Fatalf("GenerateID: %v", err)
	}
	if len(id) != 16 {
		t.Errorf("length: got %d, want 16", len(id))
	}
}

func TestGenerateID_HexChars(t *testing.T) {
	id, err := GenerateID()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("non-hex character: %c", c)
		}
	}
}

func TestGenerateID_Uniqueness(t *testing.T) {
	seen := make(map[string]struct{})
	for range 100 {
		id, err := GenerateID()
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate ID: %s", id)
		}
		seen[id] = struct{}{}
	}
}
