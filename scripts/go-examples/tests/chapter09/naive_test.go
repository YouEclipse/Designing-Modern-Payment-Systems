package naive

import (
	"math"
	"testing"
)

func TestNaiveWeightsCannotRepresentAHardRiskBlock(t *testing.T) {
	blockedButCheap := Score(Features{SuccessRate: 1, Cost: 0, RiskScore: 1})
	allowed := Score(Features{SuccessRate: 0.6, Cost: 0.5, RiskScore: 0})
	if blockedButCheap <= allowed {
		t.Fatal("expected the counterexample: a soft risk penalty can be compensated")
	}
	if !math.IsNaN(Score(Features{SuccessRate: math.NaN()})) {
		t.Fatal("naive formula must demonstrate the unvalidated NaN boundary")
	}
}
