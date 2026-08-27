package costrank

import (
	"errors"
	"reflect"
	"testing"
)

func TestRankRequiresExactCompleteEligibleSet(t *testing.T) {
	valid := map[string]Estimate{"a": {CandidateID: "a", QuoteCurrency: "USD", TotalMicros: 10}, "b": {CandidateID: "b", QuoteCurrency: "USD", TotalMicros: 10}}
	for _, ids := range [][]string{{"b", "a"}, {"a", "b"}} {
		got, err := RankEligible(ids, valid)
		if err != nil || len(got) != 2 || got[0].CandidateID != "a" {
			t.Fatalf("rank=%+v, %v", got, err)
		}
	}
	for _, tc := range []struct {
		ids       []string
		estimates map[string]Estimate
	}{
		{nil, nil}, {[]string{"a"}, nil}, {[]string{"a"}, valid},
		{[]string{"a", "a"}, valid},
		{[]string{"a"}, map[string]Estimate{"a": {CandidateID: "b", QuoteCurrency: "USD"}}},
		{[]string{"a"}, map[string]Estimate{"a": {CandidateID: "a", TotalMicros: -1, QuoteCurrency: "USD"}}},
		{[]string{"a"}, map[string]Estimate{"a": {CandidateID: "a"}}},
		{[]string{"a", "b"}, map[string]Estimate{"a": valid["a"], "b": {CandidateID: "b", QuoteCurrency: "EUR"}}},
	} {
		if got, err := RankEligible(tc.ids, tc.estimates); !errors.Is(err, ErrIncompleteComparison) || got != nil {
			t.Fatalf("invalid comparison %+v=%+v,%v", tc, got, err)
		}
	}
	if !reflect.DeepEqual(valid["a"], Estimate{CandidateID: "a", QuoteCurrency: "USD", TotalMicros: 10}) {
		t.Fatal("ranking mutated source evidence")
	}
}
