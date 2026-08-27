package selection

import (
	"errors"
	"reflect"
	"testing"
)

func TestSelectionHasTotalOrderAndDoesNotMutateInput(t *testing.T) {
	values := []Candidate{{"z", 100, 0}, {"b", 200, 1}, {"c", 200, 0}, {"a", 200, 0}}
	var permute func(int)
	permute = func(index int) {
		if index == len(values) {
			before := append([]Candidate(nil), values...)
			got, err := Select(values)
			if err != nil || got.CandidateID != "a" || !reflect.DeepEqual(before, values) {
				t.Fatalf("selection=%+v,%v input=%+v", got, err, values)
			}
			return
		}
		for i := index; i < len(values); i++ {
			values[index], values[i] = values[i], values[index]
			permute(index + 1)
			values[index], values[i] = values[i], values[index]
		}
	}
	permute(0)
}

func TestSelectionRejectsInvalidOrDuplicateCandidates(t *testing.T) {
	for _, values := range [][]Candidate{nil, {{"", 1, 0}}, {{"a", -1, 0}}, {{"a", 1, -1}}, {{"a", 1, 0}, {"a", 2, 0}}} {
		if _, err := Select(values); !errors.Is(err, ErrInvalidRankedCandidates) {
			t.Fatalf("accepted %+v: %v", values, err)
		}
	}
}
