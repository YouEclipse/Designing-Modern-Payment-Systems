package naiveregion

import (
	"context"
	"testing"
)

type regionClientFunc func(context.Context, string, string) error

func (f regionClientFunc) Confirm(ctx context.Context, intentID, key string) error {
	return f(ctx, intentID, key)
}

func TestCounterexampleDispatchesAgainAfterAmbiguousLocalResult(t *testing.T) {
	dispatches := 0
	local := regionClientFunc(func(_ context.Context, intentID, key string) error {
		if intentID != "intent-1" || key != "merchant-key" {
			t.Fatal("local identity changed")
		}
		dispatches++
		return context.DeadlineExceeded // The caller cannot infer whether funds moved.
	})
	remote := regionClientFunc(func(_ context.Context, intentID, key string) error {
		if intentID != "intent-1" || key != "merchant-key" {
			t.Fatal("remote identity changed")
		}
		dispatches++
		return nil
	})
	if err := UnsafeFailover(context.Background(), local, remote, "intent-1", "merchant-key"); err != nil || dispatches != 2 {
		t.Fatalf("counterexample no longer demonstrates two region calls: %d, %v", dispatches, err)
	}
}
