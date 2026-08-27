package shadow

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type evaluatorFunc func(context.Context, Snapshot) (Decision, error)

func (f evaluatorFunc) Evaluate(ctx context.Context, s Snapshot) (Decision, error) { return f(ctx, s) }

func shadowFixture() Job {
	return Job{ChampionDecisionID: "champion_1", ChampionCandidateID: "a", Snapshot: Snapshot{
		PaymentIntentID: "pi_1", ContextDigest: "context_1", AmountMinor: 12900, Currency: "USD", Operation: "sale", CandidateIDs: []string{"a", "b"},
		GuardrailVersion: "guard_v1", FeatureSnapshotVersion: "feature_v1", ChallengerPolicyVersion: "challenger_v1",
	}}
}

func TestShadowProducesAnalysisOnlyAndDoesNotMutateChampionSnapshot(t *testing.T) {
	job := shadowFixture()
	got := Run(context.Background(), evaluatorFunc(func(_ context.Context, s Snapshot) (Decision, error) {
		s.CandidateIDs[0] = "mutated_by_evaluator"
		return Decision{ID: "shadow_1", SelectedCandidateID: "b", PolicyVersion: "challenger_v1"}, nil
	}), job, time.Second)
	if got.Mode != "shadow" || got.State != "evaluated" || got.Agreement || got.ParentDecisionID != "champion_1" || got.ShadowDecisionID != "shadow_1" || got.SelectedCandidateID != "b" || !reflect.DeepEqual(job.Snapshot.CandidateIDs, []string{"a", "b"}) {
		t.Fatalf("shadow report/ownership=%+v job=%+v", got, job)
	}
}

func TestShadowErrorsAreSanitizedAndNotPaymentOutcomes(t *testing.T) {
	job := shadowFixture()
	for _, tc := range []struct {
		err           error
		state, reason string
	}{
		{context.DeadlineExceeded, "timeout", "shadow_budget_exhausted"}, {context.Canceled, "canceled", "shadow_job_canceled"},
		{errors.New("private diagnostic must not escape"), "error", "shadow_evaluation_error"},
	} {
		got := Run(context.Background(), evaluatorFunc(func(context.Context, Snapshot) (Decision, error) { return Decision{}, tc.err }), job, time.Second)
		if got.State != tc.state || got.Reason != tc.reason || got.SelectedCandidateID != "" {
			t.Fatalf("shadow error=%+v", got)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got := Run(ctx, evaluatorFunc(func(context.Context, Snapshot) (Decision, error) {
		return Decision{ID: "shadow_1", PolicyVersion: "challenger_v1", SelectedCandidateID: "a"}, nil
	}), job, time.Second)
	if got.State != "canceled" {
		t.Fatalf("late successful evaluator bypassed deadline: %+v", got)
	}
}

func TestShadowRejectsVersionIDOrAllowedSetDrift(t *testing.T) {
	job := shadowFixture()
	for _, decision := range []Decision{
		{ID: "shadow_1", PolicyVersion: "other_version", SelectedCandidateID: "a"},
		{ID: "champion_1", PolicyVersion: "challenger_v1", SelectedCandidateID: "a"},
		{ID: "shadow_1", PolicyVersion: "challenger_v1", SelectedCandidateID: "outside_guardrail"},
	} {
		got := Run(context.Background(), evaluatorFunc(func(context.Context, Snapshot) (Decision, error) { return decision, nil }), job, time.Second)
		if got.State != "invalid" || got.SelectedCandidateID != "" {
			t.Fatalf("invalid shadow reported as winner: %+v", got)
		}
	}
	got := Run(context.Background(), evaluatorFunc(func(context.Context, Snapshot) (Decision, error) {
		return Decision{ID: "shadow_1", PolicyVersion: "challenger_v1"}, nil
	}), job, time.Second)
	if got.State != "no_candidate" || got.Reason != "shadow_no_candidate" {
		t.Fatalf("no candidate=%+v", got)
	}
	got = Run(context.Background(), evaluatorFunc(func(context.Context, Snapshot) (Decision, error) {
		t.Fatal("invalid job called evaluator")
		return Decision{}, nil
	}), job, 0)
	if got.State != "invalid" {
		t.Fatal(got)
	}
}
