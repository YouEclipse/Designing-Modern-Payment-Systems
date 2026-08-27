package recovery

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func recoveryFixture() (Snapshot, time.Time) {
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	return Snapshot{
		Original: OriginalCommand{OperationID: "op_1", ProviderConnectionID: "conn_eu", ProviderKey: "key_1", RequestHash: "hash_1"},
		Outcome:  OutcomeUnknown, Certainty: EffectPossible, ResubmitProof: ProofNone, QuerySupported: true,
		RecoveryDeadlineAt: now.Add(time.Hour), RecoveryBudgetRemaining: 3,
	}, now
}

func TestRecoveryActionMatrixDoesNotRerouteOrMutateIdentity(t *testing.T) {
	base, now := recoveryFixture()
	for _, tc := range []struct {
		name   string
		change func(*Snapshot)
		action Action
		reason string
		review bool
	}{
		{"unknown", func(*Snapshot) {}, ActionQuery, "side_effect_possible", false},
		{"pending", func(s *Snapshot) { s.Outcome = OutcomePending }, ActionQuery, "side_effect_possible", false},
		{"no query", func(s *Snapshot) { s.QuerySupported = false }, ActionWait, "await_provider_evidence", false},
		{"backoff", func(s *Snapshot) { s.NextActionAt = now.Add(time.Second) }, ActionWait, "backoff_not_elapsed", false},
		{"local proof", func(s *Snapshot) { s.ResubmitProof = ProofLocalNeverDispatched; s.Certainty = EffectNone }, ActionSameKeyResubmit, string(ProofLocalNeverDispatched), false},
		{"provider proof", func(s *Snapshot) { s.ResubmitProof = ProofProviderConclusiveAbsent }, ActionSameKeyResubmit, string(ProofProviderConclusiveAbsent), false},
		{"deadline", func(s *Snapshot) { s.RecoveryDeadlineAt = now }, ActionStop, "recovery_deadline_exhausted", true},
		{"budget", func(s *Snapshot) { s.RecoveryBudgetRemaining = 0 }, ActionStop, "recovery_budget_exhausted", true},
		{"review", func(s *Snapshot) { s.ReviewRequired = true }, ActionStop, "conflicting_evidence", true},
		{"success", func(s *Snapshot) {
			s.Outcome, s.Certainty = OutcomeSuccess, EffectConfirmed
			s.RecoveryBudgetRemaining = 0
		}, ActionStop, "terminal_success", false},
		{"failure", func(s *Snapshot) { s.Outcome, s.Certainty = OutcomeFailure, EffectNone }, ActionStop, "conclusive_failure", false},
		{"confirmed conflict", func(s *Snapshot) { s.Certainty, s.ResubmitProof = EffectConfirmed, ProofProviderConclusiveAbsent }, ActionStop, "conflicting_evidence", true},
		{"pending conflict", func(s *Snapshot) { s.Outcome, s.ResubmitProof = OutcomePending, ProofProviderConclusiveAbsent }, ActionStop, "conflicting_evidence", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := base
			tc.change(&s)
			before := s
			got, err := Plan(now, s)
			if err != nil || got.Action != tc.action || got.Reason != tc.reason || got.RequiresReview != tc.review || !reflect.DeepEqual(before, s) || s.Original != base.Original {
				t.Fatalf("Plan=%+v,%v snapshot=%+v", got, err, s)
			}
		})
	}
}

func TestRecoveryRejectsMissingIdentityAndUnrecognizedEvidence(t *testing.T) {
	base, now := recoveryFixture()
	for _, change := range []func(*Snapshot){
		func(s *Snapshot) { s.Original.OperationID = "" }, func(s *Snapshot) { s.Original.ProviderConnectionID = "" },
		func(s *Snapshot) { s.Original.ProviderKey = "" }, func(s *Snapshot) { s.Original.RequestHash = "" },
		func(s *Snapshot) { s.RecoveryDeadlineAt = time.Time{} }, func(s *Snapshot) { s.Outcome = "HTTP_500" },
		func(s *Snapshot) { s.Certainty = "probably_none" }, func(s *Snapshot) { s.ResubmitProof = "ordinary_404" },
	} {
		bad := base
		change(&bad)
		if _, err := Plan(now, bad); !errors.Is(err, ErrInvalidSnapshot) {
			t.Fatalf("invalid snapshot accepted: %+v,%v", bad, err)
		}
	}
}
