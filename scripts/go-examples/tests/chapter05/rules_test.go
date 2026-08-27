package rules

import (
	"reflect"
	"testing"
)

func TestThreeValuedTruth(t *testing.T) {
	for _, state := range []ValueState{ValueMissing, ValueUnknown} {
		if got := Not(Equal(StringValue{State: state}, "US")); got != Unknown {
			t.Fatalf("unknown was negated into truth: %v", got)
		}
	}
	if Equal(StringValue{State: ValueKnown, Value: "US"}, "US") != True || Equal(StringValue{State: ValueKnown, Value: "DE"}, "US") != False {
		t.Fatal("known equality")
	}
	for _, test := range []struct {
		input []Truth
		want  Truth
	}{
		{nil, True}, {[]Truth{True, Unknown}, Unknown}, {[]Truth{Unknown, False}, False}, {[]Truth{False, Unknown}, False},
	} {
		if got := All(test.input...); got != test.want {
			t.Fatalf("All(%v)=%v, want %v", test.input, got, test.want)
		}
	}
}

func TestOnlyHighestPreferenceAndNoResurrection(t *testing.T) {
	hits := []PreferRule{{ID: "lower", Priority: 10, Candidates: []string{"c"}}, {ID: "higher", Priority: 20, Candidates: []string{"blocked", "b", "b"}}}
	before := append([]PreferRule(nil), hits...)
	allowed := map[string]bool{"a": true, "b": true, "c": true, "blocked": false}
	got := StablePreference(hits, allowed, []string{"a", "b", "c"})
	if !reflect.DeepEqual(got, []string{"b", "a", "c"}) {
		t.Fatalf("lower-priority preference or forbidden candidate affected order: %v", got)
	}
	if !reflect.DeepEqual(hits, before) {
		t.Fatal("input rules mutated")
	}
	if got := StablePreference(hits, map[string]bool{}, []string{"a"}); len(got) != 0 {
		t.Fatalf("empty hard set revived: %v", got)
	}
}
