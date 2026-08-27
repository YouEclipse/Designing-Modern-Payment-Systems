package envelope

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func envelopeFixture() DecisionEnvelope {
	return DecisionEnvelope{
		SchemaVersion: "aurora-decision-json-v1", DecisionID: "rd_1", CreatedAtUnixMilli: 1787745600000,
		PaymentIntentID: "pi_1", Outcome: OutcomeSelected, ReasonCodes: []string{"highest_score", "within_guardrails"},
		AttemptID: "attempt_1", RoutePlanID: "plan_1", ContextDigest: "context_hash",
		Context:       ContextSummary{SchemaVersion: "routing_context.v1", AmountMinor: 12900, Currency: "USD", PaymentMethod: "card", Operation: "sale", CaptureMode: "automatic"},
		PolicyVersion: "rules_v3", ScorerProfileVersion: "score_v5", RiskReference: "risk_1", RiskVersion: "risk_v1",
		Candidates: []CandidateEvidence{
			{CandidateID: "b", ConnectionRevision: "rev1", Provider: "psp_b", MerchantAccount: "acct_b", Connector: "adapter_b", Region: "us", Eligibility: "eligible", CapabilityVersion: "cap_v1"},
			{CandidateID: "a", ConnectionRevision: "rev1", Provider: "psp_a", MerchantAccount: "acct_a", Connector: "adapter_a", Region: "us", Eligibility: "eligible", CapabilityVersion: "cap_v1"},
			{CandidateID: "excluded", ConnectionRevision: "rev1", Provider: "psp_a", MerchantAccount: "acct_x", Connector: "adapter_a", Region: "eu", Eligibility: "ineligible", CapabilityVersion: "cap_v1"},
		},
		Exclusions:     []Exclusion{{CandidateID: "excluded", Stage: "risk", ReasonCode: "denied", EvidenceRef: "e2"}, {CandidateID: "excluded", Stage: "eligibility", ReasonCode: "denied", EvidenceRef: "e1"}},
		RuleHits:       []RuleHit{{EvaluationOrder: 2, RuleID: "rule_b", Action: "prefer", ReasonCode: "preferred"}, {EvaluationOrder: 1, RuleID: "rule_a", Action: "exclude", ReasonCode: "denied"}},
		Features:       []FeatureEvidence{{CandidateID: "b", Name: "cost", State: "known", Value: 10, Unit: "USD_micros", ArtifactVersion: "cost_v1"}, {CandidateID: "a", Name: "cost", State: "known", Value: 20, Unit: "USD_micros", ArtifactVersion: "cost_v1"}},
		Scores:         []ScoreBreakdown{{CandidateID: "b", TotalPoints: 80, Components: []ScoreComponent{{Name: "success", Contribution: 60}, {Name: "cost", Contribution: 20}}}, {CandidateID: "a", TotalPoints: 70, Components: []ScoreComponent{{Name: "success", Contribution: 60}, {Name: "cost", Contribution: 10}}}},
		CandidateOrder: []string{"b", "a"}, Selected: &Selection{CandidateID: "b", Provider: "psp_b", MerchantAccount: "acct_b", Connector: "adapter_b", Region: "us", TieBreak: "score_desc_policy_rank_id"},
	}
}

func TestCanonicalizationSortsSetsWithoutMutatingTheOriginal(t *testing.T) {
	original := envelopeFixture()
	before, _ := json.Marshal(original)
	wantBytes, wantHash, err := CanonicalBytes(original)
	if err != nil || sha256.Sum256(wantBytes) != wantHash {
		t.Fatalf("canonical=%s,%x,%v", wantBytes, wantHash, err)
	}
	after, _ := json.Marshal(original)
	if !bytes.Equal(before, after) {
		t.Fatal("canonicalization mutated caller slices")
	}
	permuted := envelopeFixture()
	permuted.Candidates[0], permuted.Candidates[1] = permuted.Candidates[1], permuted.Candidates[0]
	permuted.Exclusions[0], permuted.Exclusions[1] = permuted.Exclusions[1], permuted.Exclusions[0]
	permuted.RuleHits[0], permuted.RuleHits[1] = permuted.RuleHits[1], permuted.RuleHits[0]
	permuted.Features[0], permuted.Features[1] = permuted.Features[1], permuted.Features[0]
	permuted.Scores[0], permuted.Scores[1] = permuted.Scores[1], permuted.Scores[0]
	permuted.ReasonCodes[0], permuted.ReasonCodes[1] = permuted.ReasonCodes[1], permuted.ReasonCodes[0]
	for i := range permuted.Scores {
		permuted.Scores[i].Components[0], permuted.Scores[i].Components[1] = permuted.Scores[i].Components[1], permuted.Scores[i].Components[0]
	}
	gotBytes, gotHash, err := CanonicalBytes(permuted)
	if err != nil || gotHash != wantHash || !bytes.Equal(gotBytes, wantBytes) {
		t.Fatalf("unordered evidence changed canonical bytes: %s,%v", gotBytes, err)
	}
	if !reflect.DeepEqual(original.CandidateOrder, []string{"b", "a"}) {
		t.Fatal("semantic order sorted")
	}
	permuted.CandidateOrder = []string{"a", "b"}
	if _, changedHash, err := CanonicalBytes(permuted); err != nil || changedHash == wantHash {
		t.Fatal("changing semantic order must change history hash")
	}
}

func negativeFixture(outcome DecisionOutcome) DecisionEnvelope {
	e := envelopeFixture()
	e.Outcome = outcome
	e.ReasonCodes = []string{"no_executable_route"}
	e.AttemptID = ""
	e.RoutePlanID = ""
	e.CandidateOrder = nil
	e.Selected = nil
	e.Scores = nil
	return e
}

func TestCanonicalEnvelopeAcceptsAuditableNegativeOutcomesWithoutExecutionIdentity(t *testing.T) {
	for _, outcome := range []DecisionOutcome{
		OutcomeNoEligibleCandidate, OutcomeBlocked, OutcomeUnsupported,
		OutcomeRequiresReview, OutcomeRiskUnavailable, OutcomeFeatureUnavailable,
	} {
		e := negativeFixture(outcome)
		if _, _, err := CanonicalBytes(e); err != nil {
			t.Fatalf("negative outcome %q rejected: %v", outcome, err)
		}
	}

	e := negativeFixture(OutcomeNoEligibleCandidate)
	e.Assignment = &AssignmentEvidence{ExperimentID: "exp_1", Variant: "treatment", Version: "v3", Bucket: 412, State: "not_exposed"}
	if _, _, err := CanonicalBytes(e); err != nil {
		t.Fatalf("assignment-only negative outcome rejected: %v", err)
	}
}

func TestCanonicalEnvelopeRejectsOutcomeBranchContradictions(t *testing.T) {
	for _, change := range []func(*DecisionEnvelope){
		func(e *DecisionEnvelope) { e.Outcome = "invented" },
		func(e *DecisionEnvelope) { e.ReasonCodes = nil },
		func(e *DecisionEnvelope) { e.ReasonCodes = []string{"same", "same"} },
		func(e *DecisionEnvelope) { e.AttemptID = "" },
		func(e *DecisionEnvelope) { e.RoutePlanID = "" },
		func(e *DecisionEnvelope) { e.Selected = nil },
		func(e *DecisionEnvelope) { e.Assignment = &AssignmentEvidence{ExperimentID: "exp_1", Variant: "treatment", Version: "v3", Bucket: 1, State: "not_exposed"} },
	} {
		e := envelopeFixture()
		change(&e)
		if _, _, err := CanonicalBytes(e); !errors.Is(err, ErrInvalidEnvelope) {
			t.Fatalf("contradictory selected outcome accepted: %v", err)
		}
	}

	for _, change := range []func(*DecisionEnvelope){
		func(e *DecisionEnvelope) { e.AttemptID = "attempt_illegal" },
		func(e *DecisionEnvelope) { e.RoutePlanID = "plan_illegal" },
		func(e *DecisionEnvelope) { e.Selected = &Selection{CandidateID: "b", Provider: "psp_b"} },
		func(e *DecisionEnvelope) { e.CandidateOrder = []string{"b"} },
		func(e *DecisionEnvelope) { e.Assignment = &AssignmentEvidence{ExperimentID: "exp_1", Variant: "treatment", Version: "v3", Bucket: 1, State: "assigned"} },
	} {
		e := negativeFixture(OutcomeBlocked)
		change(&e)
		if _, _, err := CanonicalBytes(e); !errors.Is(err, ErrInvalidEnvelope) {
			t.Fatalf("negative outcome with execution identity accepted: %v", err)
		}
	}
}

func TestCanonicalEnvelopeRejectsAmbiguousOrIneligibleEvidence(t *testing.T) {
	for _, change := range []func(*DecisionEnvelope){
		func(e *DecisionEnvelope) { e.SchemaVersion = "unknown" }, func(e *DecisionEnvelope) { e.CandidateOrder = []string{"b", "b"} },
		func(e *DecisionEnvelope) { e.CandidateOrder = []string{"b", "not_in_catalog"} }, func(e *DecisionEnvelope) { e.CandidateOrder = []string{"b", "excluded"} },
		func(e *DecisionEnvelope) { e.Candidates[0].Eligibility = "unknown" }, func(e *DecisionEnvelope) { e.Selected.MerchantAccount = "other_account" },
		func(e *DecisionEnvelope) { e.Candidates = append(e.Candidates, e.Candidates[0]) },
		func(e *DecisionEnvelope) { e.Features = append(e.Features, e.Features[0]) },
		func(e *DecisionEnvelope) { e.Scores = append(e.Scores, e.Scores[0]) },
		func(e *DecisionEnvelope) {
			e.Scores[0].Components = append(e.Scores[0].Components, e.Scores[0].Components[0])
		},
		func(e *DecisionEnvelope) { e.RuleHits = append(e.RuleHits, e.RuleHits[0]) },
	} {
		e := envelopeFixture()
		change(&e)
		if _, _, err := CanonicalBytes(e); !errors.Is(err, ErrInvalidEnvelope) {
			t.Fatalf("ambiguous envelope accepted: %v", err)
		}
	}
}

func TestHistoricalHashChangesWhenEvidenceOrVersionsChange(t *testing.T) {
	_, originalHash, _ := CanonicalBytes(envelopeFixture())
	for _, change := range []func(*DecisionEnvelope){
		func(e *DecisionEnvelope) { e.PolicyVersion = "rules_v4" }, func(e *DecisionEnvelope) { e.ScorerProfileVersion = "score_v6" },
		func(e *DecisionEnvelope) { e.Features[0].Value++ }, func(e *DecisionEnvelope) { e.Exclusions[0].EvidenceRef = "new_fact" },
		func(e *DecisionEnvelope) { e.Candidates[0].ConnectionRevision = "rev2" }, func(e *DecisionEnvelope) { e.Context.AmountMinor++ },
	} {
		e := envelopeFixture()
		change(&e)
		if _, got, err := CanonicalBytes(e); err != nil || got == originalHash {
			t.Fatalf("changed evidence reused original hash: %x,%v", got, err)
		}
	}
}
