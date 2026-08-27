package naive

import (
	"testing"
	"time"
)

func TestCounterexampleChangesFundsStatusWithoutEvidence(t *testing.T) {
	updated := time.Unix(1_800_000_000, 0)
	if WrongForFunds(updated, updated.Add(24*time.Hour)) != "processing" ||
		WrongForFunds(updated, updated.Add(25*time.Hour)) != "failed" {
		t.Fatal("counterexample no longer demonstrates the elapsed-time-only mutation")
	}
}
