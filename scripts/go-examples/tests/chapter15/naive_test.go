package naive

import (
	"context"
	"testing"
)

func TestCounterexampleCannotDetermineFundsOutcome(t *testing.T) {
	// A transport timeout can conceal a completed payment. The counterexample
	// reports failure solely from an error, demonstrating the lost distinction.
	if got := MetricResult(context.DeadlineExceeded); got != "failure" {
		t.Fatalf("counterexample changed: %q", got)
	}
	// A parsed issuer decline can have no Go transport error at all.
	if got := MetricResult(nil); got != "success" {
		t.Fatalf("counterexample changed: %q", got)
	}
}
