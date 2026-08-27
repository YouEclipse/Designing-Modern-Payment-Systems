package routing

import (
	"reflect"
	"testing"
)

func TestHardcodedExampleShowsItsLimitedInputs(t *testing.T) {
	for _, tc := range []struct{ currency, issuer, want string }{
		{"BRL", "US", "psp_b"}, {"USD", "BR", "psp_b"}, {"EUR", "DE", "psp_a"},
	} {
		if got := selectProvider(tc.currency, tc.issuer); got != tc.want {
			t.Fatalf("selectProvider(%q,%q)=%q, want %q", tc.currency, tc.issuer, got, tc.want)
		}
	}
	// This branch has no eligibility input: even an unsupported currency goes
	// to a default. The following tests prove the stronger Route boundary.
	if selectProvider("UNSUPPORTED", "") != "psp_a" {
		t.Fatal("the intentionally limited example no longer demonstrates implicit fallback")
	}
}

func TestEligibilityPrecedesScoreAndNoDefaultIsInvented(t *testing.T) {
	values := []Candidate{
		{Name: "forbidden", Eligible: false, SuccessRate: 1, NormalizedCost: 0},
		{Name: "allowed", Eligible: true, SuccessRate: 0.8, NormalizedCost: 0.5},
	}
	got, err := Route(values)
	if err != nil || got.Provider != "allowed" || !reflect.DeepEqual(got.Reasons, []string{"ELIGIBLE", "BEST_COMPOSITE_SCORE"}) {
		t.Fatalf("Route=%+v, %v", got, err)
	}
	for _, empty := range [][]Candidate{nil, values[:1]} {
		if got, err := Route(empty); err == nil || got.Provider != "" {
			t.Fatalf("no candidate must be explicit: %+v, %v", got, err)
		}
	}
}

func TestStableTieAndHealthPenalty(t *testing.T) {
	values := []Candidate{{Name: "b", Eligible: true, SuccessRate: 0.9}, {Name: "a", Eligible: true, SuccessRate: 0.9}}
	for i := 0; i < 2; i++ {
		got, err := Route(values)
		if err != nil || got.Provider != "a" {
			t.Fatalf("tie changed with input order: %+v, %v", got, err)
		}
		values[0], values[1] = values[1], values[0]
	}
	values[1].Degraded = true // a is in position 1 again.
	if got, err := Route(values); err != nil || got.Provider != "b" {
		t.Fatalf("health penalty not applied: %+v, %v", got, err)
	}
}
