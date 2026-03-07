package utils

import (
	"testing"
)

// --- LookupCopy ---

func TestLookupCopy_Found(t *testing.T) {
	type val struct {
		Name string
	}
	m := map[string]*val{
		"k1": {Name: "hello"},
	}

	got, err := LookupCopy(m, "k1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "hello" {
		t.Errorf("got %q, want %q", got.Name, "hello")
	}

	// Verify it's a copy — modifying returned value shouldn't affect map.
	got.Name = "modified"
	if m["k1"].Name != "hello" {
		t.Error("modifying copy affected original")
	}
}

func TestLookupCopy_NotFound(t *testing.T) {
	m := map[string]*int{}
	_, err := LookupCopy(m, "missing")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestLookupCopy_NilValue(t *testing.T) {
	m := map[string]*int{"k": nil}
	_, err := LookupCopy(m, "k")
	if err == nil {
		t.Fatal("expected error for nil value")
	}
}

// --- MergeSets ---

func TestMergeSets_Basic(t *testing.T) {
	s1 := map[string]struct{}{"a": {}, "b": {}}
	s2 := map[string]struct{}{"b": {}, "c": {}}

	got := MergeSets(s1, s2)
	if len(got) != 3 {
		t.Errorf("got %d items, want 3", len(got))
	}
	for _, k := range []string{"a", "b", "c"} {
		if _, ok := got[k]; !ok {
			t.Errorf("missing key %q", k)
		}
	}
}

func TestMergeSets_Empty(t *testing.T) {
	got := MergeSets[string]()
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestMergeSets_NilSets(t *testing.T) {
	var s1 map[string]struct{}
	got := MergeSets(s1, nil)
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestMergeSets_Single(t *testing.T) {
	s := map[int]struct{}{1: {}, 2: {}}
	got := MergeSets(s)
	if len(got) != 2 {
		t.Errorf("got %d, want 2", len(got))
	}
}

func TestMergeSets_DoesNotModifyInput(t *testing.T) {
	s1 := map[string]struct{}{"a": {}}
	s2 := map[string]struct{}{"b": {}}

	_ = MergeSets(s1, s2)

	if len(s1) != 1 {
		t.Error("s1 was modified")
	}
	if len(s2) != 1 {
		t.Error("s2 was modified")
	}
}
