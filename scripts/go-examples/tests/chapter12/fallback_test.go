package fallback

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func fallbackFixture() (Snapshot, time.Time) {
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	return Snapshot{
		Prior: PriorAttempt{ID: "attempt_1", Number: 1, CandidateID: "a", OperationKind: "sale", Terminal: true, SideEffect: SideEffectNone, NormalizedReason: "provider_unavailable_pre_dispatch", TriedCandidates: map[string]struct{}{"a": {}}},
		Plan:  RoutePlan{ID: "plan_1", Version: "plan_v1", Candidates: []Candidate{{ID: "a"}, {ID: "b"}, {ID: "c"}}, MaxAttempts: 3, DeadlineAt: now.Add(time.Minute)},
		Risk:  RiskAllow, RiskExpires: now.Add(time.Minute), Allowed: map[string]struct{}{"provider_unavailable_pre_dispatch": {}},
	}, now
}

func TestFallbackRequiresConclusiveNoEffectAndPolicyPermission(t *testing.T) {
	for _, tc := range []struct {
		name      string
		change    func(*Snapshot)
		action    Action
		reason    string
		candidate string
	}{
		{"allowlisted no effect", func(*Snapshot) {}, ActionCreateAttempt, "allowlisted_no_effect_failure", "b"},
		{"unknown", func(s *Snapshot) { s.Prior.Terminal, s.Prior.SideEffect = false, SideEffectPossible }, ActionRecoverOriginal, "prior_not_terminal", ""},
		{"terminal possible", func(s *Snapshot) { s.Prior.SideEffect = SideEffectPossible }, ActionStop, "side_effect_not_none", ""},
		{"confirmed", func(s *Snapshot) { s.Prior.SideEffect = SideEffectConfirmed }, ActionStop, "side_effect_not_none", ""},
		{"issuer decline", func(s *Snapshot) { s.Prior.NormalizedReason = "issuer_decline" }, ActionStop, "reason_not_allowed", ""},
		{"review", func(s *Snapshot) { s.Prior.ReviewRequired = true }, ActionStop, "review_required", ""},
		{"action", func(s *Snapshot) { s.Prior.RequiresUserAction = true }, ActionStop, "unsupported_or_user_action", ""},
		{"risk block", func(s *Snapshot) { s.Risk = RiskBlock }, ActionStop, "risk_guardrail", ""},
		{"risk 3ds", func(s *Snapshot) { s.Risk = RiskRequire3DS }, ActionStop, "risk_guardrail", ""},
		{"risk review", func(s *Snapshot) { s.Risk = RiskReview }, ActionStop, "risk_guardrail", ""},
		{"expired risk", func(s *Snapshot) { _, at := fallbackFixture(); s.RiskExpires = at }, ActionStop, "risk_guardrail", ""},
		{"deadline", func(s *Snapshot) { _, at := fallbackFixture(); s.Plan.DeadlineAt = at }, ActionStop, "budget_or_deadline", ""},
		{"budget", func(s *Snapshot) { s.Plan.MaxAttempts = 1 }, ActionStop, "budget_or_deadline", ""},
		{"exhausted", func(s *Snapshot) { s.Plan.Candidates = []Candidate{{ID: "a"}} }, ActionStop, "plan_exhausted", ""},
		{"fixed order", func(s *Snapshot) { s.Plan.Candidates = []Candidate{{ID: "a"}, {ID: "c"}, {ID: "b"}} }, ActionCreateAttempt, "allowlisted_no_effect_failure", "c"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, now := fallbackFixture()
			tc.change(&s)
			got, err := Plan(now, s)
			if err != nil || got.Action != tc.action || got.Reason != tc.reason || got.NextCandidateID != tc.candidate {
				t.Fatalf("Plan=%+v,%v", got, err)
			}
		})
	}
}

func TestFollowupOperationsNeverEnterFallback(t *testing.T) {
	for _, kind := range []string{"authorize", "capture", "void", "refund", "reversal"} {
		s, now := fallbackFixture()
		s.Prior.OperationKind = kind
		if got, err := Plan(now, s); err != nil || got.Action != ActionStop || got.NextCandidateID != "" {
			t.Fatalf("%s rerouted: %+v,%v", kind, got, err)
		}
	}
}

func TestMalformedPlanCannotHideAfterFirstUntriedCandidate(t *testing.T) {
	for _, change := range []func(*Snapshot){
		func(s *Snapshot) { s.Prior.ID = "" }, func(s *Snapshot) { s.Prior.Number = 0 },
		func(s *Snapshot) { s.Prior.TriedCandidates = nil },
		func(s *Snapshot) { s.Plan.Candidates = []Candidate{{"a"}, {"b"}, {""}} },
		func(s *Snapshot) { s.Plan.Candidates = []Candidate{{"a"}, {"b"}, {"b"}} },
		func(s *Snapshot) { s.Plan.Candidates = []Candidate{{"b"}} },
		func(s *Snapshot) { s.Prior.TriedCandidates["outside_plan"] = struct{}{} },
	} {
		s, now := fallbackFixture()
		change(&s)
		if _, err := Plan(now, s); !errors.Is(err, ErrInvalidSnapshot) {
			t.Fatalf("malformed snapshot accepted: %+v,%v", s, err)
		}
	}
	s, now := fallbackFixture()
	want, err := Plan(now, s)
	for i := 0; i < 20; i++ {
		got, nextErr := Plan(now, s)
		if err != nil || nextErr != nil || !reflect.DeepEqual(got, want) || s.Prior.Number != 1 {
			t.Fatal("planner changed its durable inputs")
		}
	}
}
