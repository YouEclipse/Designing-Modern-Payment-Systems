package configruntime

import (
	"context"
	"testing"
	"time"
)

func TestRollbackUsesNewGenerationAndPreservesCompleteSnapshot(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	ctx := context.Background()
	journal := &memoryJournal{}
	runtime, err := NewRuntime("prod/eu", journal, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Decide(ctx, Request{}, now); err == nil {
		t.Fatal("cold runtime invented an approved configuration")
	}
	for _, candidate := range []Prepared{prepared(7, "a", now), prepared(8, "b", now)} {
		if err := runtime.Apply(ctx, candidate, now); err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.Apply(ctx, prepared(7, "a", now), now); err == nil {
		t.Fatal("rollback reused an old generation")
	}
	if err := runtime.Apply(ctx, prepared(8, "a", now), now); err == nil {
		t.Fatal("same generation changed its content")
	}
	rollback := prepared(9, "a", now)
	if err := runtime.Apply(ctx, rollback, now); err != nil {
		t.Fatal(err)
	}
	journal.reject = true
	if err := runtime.Apply(ctx, rollback, now); err != nil {
		t.Fatal("an already active exact retry should be idempotent")
	}
	decision, err := runtime.Decide(ctx, Request{}, now)
	if err != nil || decision.Generation != 9 || decision.Candidate != "a" || decision.Digest != rollback.Activation.Digest {
		t.Fatalf("rollback mixed release and content: %+v, %v", decision, err)
	}
}

func TestExpiredOrMistargetedActivationCannotServeNewDecisions(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	ctx := context.Background()
	for _, variant := range []string{"expired", "future", "scope"} {
		t.Run(variant, func(t *testing.T) {
			runtime, err := NewRuntime("prod/eu", &memoryJournal{}, nil)
			if err != nil {
				t.Fatal(err)
			}
			candidate := prepared(1, "a", now)
			switch variant {
			case "expired":
				candidate.Activation.ExpiresAt = now
			case "future":
				candidate.Activation.NotBefore = now.Add(time.Minute)
			case "scope":
				candidate.Activation.Scope = "prod/us"
			}
			if err := runtime.Apply(ctx, candidate, now); err == nil {
				t.Fatal("unusable activation applied")
			}
		})
	}
	expired := prepared(5, "a", now.Add(-2*time.Hour))
	runtime, err := NewRuntime("prod/eu", &memoryJournal{}, &expired)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Decide(ctx, Request{}, now); err == nil {
		t.Fatal("restored but expired last-known-good served a new decision")
	}
}
