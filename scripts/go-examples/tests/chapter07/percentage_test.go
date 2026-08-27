package simplecost

import (
	"errors"
	"math"
	"testing"
)

func TestPercentageOnlyUsesEligibleIDsAndStableTies(t *testing.T) {
	rates := map[string]int64{"b": 200, "a": 200, "not_eligible": 0}
	for _, ids := range [][]string{{"b", "a"}, {"a", "b"}} {
		got, err := CheapestPercentage(10_001, ids, rates)
		if err != nil || got.CandidateID != "a" || got.FeeMinor != 200 {
			t.Fatalf("quote=%+v, %v", got, err)
		}
	}
}

func TestPercentageRejectsMissingNegativeAndOverflow(t *testing.T) {
	for _, tc := range []struct {
		amount int64
		ids    []string
		rates  map[string]int64
		want   error
	}{
		{0, []string{"a"}, map[string]int64{"a": 1}, ErrInvalidAmount},
		{100, nil, nil, ErrMissingRate},
		{100, []string{"a"}, nil, ErrMissingRate},
		{100, []string{"a"}, map[string]int64{"a": -1}, ErrMissingRate},
		{math.MaxInt64, []string{"a"}, map[string]int64{"a": 2}, ErrOverflow},
	} {
		if got, err := CheapestPercentage(tc.amount, tc.ids, tc.rates); !errors.Is(err, tc.want) || got != (Quote{}) {
			t.Fatalf("quote=%+v, err=%v, want %v", got, err, tc.want)
		}
	}
	if got, err := CheapestPercentage(math.MaxInt64, []string{"a"}, map[string]int64{"a": 0}); err != nil || got.FeeMinor != 0 {
		t.Fatalf("explicit zero rate is not missing: %+v, %v", got, err)
	}
}
