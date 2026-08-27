package riskcontract

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func riskFixture() (time.Time, Decision) {
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	return now, Decision{DecisionID: "risk_1", SubjectDigest: "context_1", Action: Allow, ReasonCodes: []string{"policy_checked"}, Version: "risk_v1", EvaluatedAt: now, ExpiresAt: now.Add(time.Minute)}
}

func TestRiskActionsAreGatesNotScores(t *testing.T) {
	now, d := riskFixture()
	for _, tc := range []struct {
		action     Action
		continuing bool
		outcome    string
	}{
		{Allow, true, "allowed"}, {Block, false, "blocked"}, {ManualReview, false, "requires_review"}, {Require3DS, false, "authentication_unsupported"},
	} {
		d.Action = tc.action
		got, err := Evaluate(now, d.SubjectDigest, d)
		if err != nil || got.Continue != tc.continuing || got.Outcome != tc.outcome {
			t.Fatalf("%s: %+v,%v", tc.action, got, err)
		}
		if tc.action == Require3DS && !reflect.DeepEqual(got.RequiredCapabilities, []string{"3ds_required"}) {
			t.Fatalf("3DS lost hard capability: %+v", got)
		}
		got.ReasonCodes[0] = "changed"
		if d.ReasonCodes[0] != "policy_checked" {
			t.Fatal("gate aliases input evidence")
		}
	}
}

func TestRiskBindingExpiryUnknownActionAndReasonValidation(t *testing.T) {
	now, valid := riskFixture()
	for _, change := range []func(*Decision){
		func(d *Decision) { d.SubjectDigest = "other" }, func(d *Decision) { d.Version = "" },
		func(d *Decision) { d.EvaluatedAt = now.Add(time.Second) }, func(d *Decision) { d.ExpiresAt = now },
		func(d *Decision) { d.Action = "unknown" }, func(d *Decision) { d.ReasonCodes = []string{"z", "a"} },
		func(d *Decision) { d.ReasonCodes = []string{"a", "a"} }, func(d *Decision) { d.ReasonCodes = []string{""} },
		func(d *Decision) { d.Action, d.ReasonCodes = Block, nil },
	} {
		bad := valid
		change(&bad)
		if got, err := Evaluate(now, "context_1", bad); !errors.Is(err, ErrInvalidRiskDecision) || got.Continue {
			t.Fatalf("invalid risk allowed: %+v,%v", got, err)
		}
	}
	valid.SubjectDigest = ""
	if _, err := Evaluate(now, "", valid); !errors.Is(err, ErrInvalidRiskDecision) {
		t.Fatal("two empty subjects are not a binding")
	}
}
