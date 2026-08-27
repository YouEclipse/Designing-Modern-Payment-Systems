package selectedonly

import "testing"

func TestSelectedOnlyRecordRetainsCorrelationButNoDecisionEvidence(t *testing.T) {
	got := NewRecord("rd_1", "pi_1", "attempt_1", "psp_a")
	if got != (Record{DecisionID: "rd_1", PaymentIntentID: "pi_1", AttemptID: "attempt_1", SelectedProvider: "psp_a"}) {
		t.Fatal(got)
	}
	// Two hypothetical connections of the same PSP cannot be distinguished by
	// this constructor: it accepts only the brand, no account/region/version.
	otherConnection := NewRecord("rd_1", "pi_1", "attempt_1", "psp_a")
	if got != otherConnection {
		t.Fatal("selected-only counterexample unexpectedly carries execution evidence")
	}
}
