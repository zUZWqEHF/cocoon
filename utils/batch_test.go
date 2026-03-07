package utils

import (
	"context"
	"fmt"
	"testing"
)

func TestForEach_AllSucceed(t *testing.T) {
	var called []string
	result := ForEach(context.Background(), []string{"a", "b", "c"}, func(_ context.Context, id string) error {
		called = append(called, id)
		return nil
	})

	if len(result.Succeeded) != 3 {
		t.Errorf("succeeded: got %d, want 3", len(result.Succeeded))
	}
	if len(result.Errors) != 0 {
		t.Errorf("errors: got %d, want 0", len(result.Errors))
	}
	if result.Err() != nil {
		t.Errorf("Err(): got %v, want nil", result.Err())
	}
	if len(called) != 3 {
		t.Errorf("called: got %d, want 3", len(called))
	}
}

func TestForEach_AllFail(t *testing.T) {
	result := ForEach(context.Background(), []string{"x", "y"}, func(_ context.Context, id string) error {
		return fmt.Errorf("fail %s", id)
	})

	if len(result.Succeeded) != 0 {
		t.Errorf("succeeded: got %d, want 0", len(result.Succeeded))
	}
	if len(result.Errors) != 2 {
		t.Errorf("errors: got %d, want 2", len(result.Errors))
	}
	if result.Err() == nil {
		t.Error("Err(): expected non-nil")
	}
}

func TestForEach_PartialFailure(t *testing.T) {
	result := ForEach(context.Background(), []string{"ok", "fail", "ok2"}, func(_ context.Context, id string) error {
		if id == "fail" {
			return fmt.Errorf("error on %s", id)
		}
		return nil
	})

	if len(result.Succeeded) != 2 {
		t.Errorf("succeeded: got %v, want 2", result.Succeeded)
	}
	if len(result.Errors) != 1 {
		t.Errorf("errors: got %d, want 1", len(result.Errors))
	}
}

func TestForEach_EmptyIDs(t *testing.T) {
	result := ForEach(context.Background(), nil, func(_ context.Context, _ string) error {
		t.Fatal("should not be called")
		return nil
	})

	if len(result.Succeeded) != 0 || len(result.Errors) != 0 {
		t.Errorf("expected empty result, got %+v", result)
	}
	if result.Err() != nil {
		t.Errorf("Err(): got %v, want nil", result.Err())
	}
}

func TestBatchResult_Err_NilForNoErrors(t *testing.T) {
	r := BatchResult{Succeeded: []string{"a"}}
	if r.Err() != nil {
		t.Errorf("expected nil, got %v", r.Err())
	}
}
