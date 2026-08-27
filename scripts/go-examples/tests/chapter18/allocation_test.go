package costallocation

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"testing"
)

func TestSplitMinorUsesStableLargestRemainderAndSignedAdjustments(t *testing.T) {
	for _, tc := range []struct {
		name    string
		total   int64
		weights []Weight
		want    []Share
	}{
		{"positive tie", 5, []Weight{{"b", 1}, {"a", 1}}, []Share{{"a", "USD", 3}, {"b", "USD", 2}}},
		{"negative tie", -5, []Weight{{"b", 1}, {"a", 1}}, []Share{{"a", "USD", -3}, {"b", "USD", -2}}},
		{"three way tie", 2, []Weight{{"c", 1}, {"b", 1}, {"a", 1}}, []Share{{"a", "USD", 1}, {"b", "USD", 1}, {"c", "USD", 0}}},
		{"unequal", 7, []Weight{{"a", 1}, {"b", 2}, {"c", 3}}, []Share{{"a", "USD", 1}, {"b", "USD", 2}, {"c", "USD", 4}}},
		{"zero", 0, []Weight{{"b", 1}, {"a", 2}}, []Share{{"a", "USD", 0}, {"b", "USD", 0}}},
		{"full signed minimum", math.MinInt64, []Weight{{"a", 1}}, []Share{{"a", "USD", math.MinInt64}}},
		{"full signed maximum", math.MaxInt64, []Weight{{"a", 1}}, []Share{{"a", "USD", math.MaxInt64}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := append([]Weight(nil), tc.weights...)
			got, err := SplitMinor(tc.total, "USD", tc.weights)
			if err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("shares=%+v err=%v, want %+v", got, err, tc.want)
			}
			if !reflect.DeepEqual(tc.weights, before) {
				t.Fatal("allocation rewrote weight snapshot")
			}
			assertConserved(t, tc.total, "USD", got)
		})
	}
}

func TestSplitMinorSurvivesLargeWeightsAndAmounts(t *testing.T) {
	weights := []Weight{{"a", math.MaxUint64}, {"b", math.MaxUint64 - 1}, {"c", 1}}
	for _, total := range []int64{math.MinInt64, math.MaxInt64, -1, 0, 1} {
		got, err := SplitMinor(total, "JPY", weights)
		if err != nil {
			t.Fatalf("total=%d: %v", total, err)
		}
		assertConserved(t, total, "JPY", got)
		reordered, err := SplitMinor(total, "JPY", []Weight{weights[2], weights[0], weights[1]})
		if err != nil || !reflect.DeepEqual(got, reordered) {
			t.Fatalf("allocation depends on iteration order: %+v / %+v, %v", got, reordered, err)
		}
	}
}

func TestSplitMinorRejectsInvalidAllocationContract(t *testing.T) {
	tooMany := make([]Weight, 10001)
	for i := range tooMany {
		tooMany[i] = Weight{TargetID: fmt.Sprintf("target_%05d", i), Units: 1}
	}
	for _, tc := range []struct {
		name, currency string
		weights        []Weight
	}{
		{"empty currency", "", []Weight{{"a", 1}}},
		{"empty targets", "USD", nil},
		{"empty target ID", "USD", []Weight{{"", 1}}},
		{"duplicate target", "USD", []Weight{{"a", 1}, {"a", 2}}},
		{"zero weight", "USD", []Weight{{"a", 0}}},
		{"too many targets", "USD", tooMany},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got, err := SplitMinor(5, tc.currency, tc.weights); err == nil || got != nil {
				t.Fatalf("invalid allocation accepted: %+v, %v", got, err)
			}
		})
	}
}

func FuzzSplitMinorConservesAndIsOrderIndependent(f *testing.F) {
	f.Add(int64(5), uint64(1), uint64(1), uint64(1))
	f.Add(int64(-5), uint64(1), uint64(2), uint64(3))
	f.Add(int64(math.MinInt64), uint64(math.MaxUint64), uint64(2), uint64(1))
	f.Add(int64(math.MaxInt64), uint64(1), uint64(math.MaxUint64), uint64(1))
	f.Fuzz(func(t *testing.T, total int64, a, b, c uint64) {
		if a == 0 || b == 0 || c == 0 {
			t.Skip("positive weight contract")
		}
		weights := []Weight{{"a", a}, {"b", b}, {"c", c}}
		got, err := SplitMinor(total, "EUR", weights)
		if err != nil {
			t.Fatal(err)
		}
		assertConserved(t, total, "EUR", got)
		permuted, err := SplitMinor(total, "EUR", []Weight{weights[2], weights[1], weights[0]})
		if err != nil || !reflect.DeepEqual(got, permuted) {
			t.Fatalf("non-deterministic allocation: %+v / %+v, %v", got, permuted, err)
		}
	})
}

func assertConserved(t *testing.T, total int64, currency string, shares []Share) {
	t.Helper()
	sum := new(big.Int)
	for i, share := range shares {
		if share.Currency != currency || (i > 0 && shares[i-1].TargetID >= share.TargetID) {
			t.Fatalf("currency or deterministic order lost: %+v", shares)
		}
		if (total >= 0 && share.AmountMinor < 0) || (total <= 0 && share.AmountMinor > 0) {
			t.Fatalf("share sign conflicts with source total: %+v", shares)
		}
		sum.Add(sum, big.NewInt(share.AmountMinor))
	}
	if sum.Cmp(big.NewInt(total)) != 0 {
		t.Fatalf("allocated %s minor units, source total=%d", sum, total)
	}
}
