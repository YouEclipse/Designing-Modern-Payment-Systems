package cost

import (
	"errors"
	"math"
	"math/big"
	"testing"
	"time"
)

func feeFixture() (FeeSchedule, Input) {
	at := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	return FeeSchedule{
		ScheduleID: "fee_1", ArtifactVersion: "fees_v1", CandidateID: "conn_a",
		QuoteCurrency: "USD", EffectiveFrom: at, EffectiveUntil: at.Add(time.Hour),
		FixedMicros: 300_000, VariableBPS: 260, PercentageRound: RoundHalfUp,
	}, Input{AmountMinor: 1000, PaymentCurrency: "USD", CurrencyExponent: 2, DecisionAt: at}
}

func TestFixedFeeChangesWinnerWithAmountAndPreservesEvidence(t *testing.T) {
	a, in := feeFixture()
	b := a
	b.CandidateID, b.FixedMicros, b.VariableBPS = "conn_b", 100_000, 290
	for _, tc := range []struct{ amount, wantA, wantB int64 }{{1000, 560_000, 390_000}, {10000, 2_900_000, 3_000_000}} {
		in.AmountMinor = tc.amount
		left, errA := EstimateSameCurrency(a, in)
		right, errB := EstimateSameCurrency(b, in)
		if errA != nil || errB != nil || left.TotalMicros != tc.wantA || right.TotalMicros != tc.wantB {
			t.Fatalf("amount %d: %+v/%v, %+v/%v", tc.amount, left, errA, right, errB)
		}
		if left.ScheduleID != a.ScheduleID || left.ArtifactVersion != a.ArtifactVersion || !left.EvaluatedAt.Equal(in.DecisionAt) || len(left.Breakdown) != 2 {
			t.Fatalf("estimate lost evidence: %+v", left)
		}
	}
}

func TestEffectiveIntervalFXAndCrossBorderAreExplicit(t *testing.T) {
	s, in := feeFixture()
	for _, tc := range []struct {
		at   time.Time
		want error
	}{
		{s.EffectiveFrom.Add(-time.Nanosecond), ErrScheduleNotEffective},
		{s.EffectiveFrom, nil}, {s.EffectiveUntil.Add(-time.Nanosecond), nil},
		{s.EffectiveUntil, ErrScheduleNotEffective},
	} {
		in.DecisionAt = tc.at
		if _, err := EstimateSameCurrency(s, in); !errors.Is(err, tc.want) {
			t.Fatalf("at %s: %v, want %v", tc.at, err, tc.want)
		}
	}
	s, in = feeFixture()
	s.CrossBorderBPS = 100
	if _, err := EstimateSameCurrency(s, in); !errors.Is(err, ErrUnknownCrossBorder) {
		t.Fatalf("unknown cross-border: %v", err)
	}
	in.CrossBorder = CrossBorderFact{State: FactKnown, Value: true, RuleVersion: "geo_v2"}
	got, err := EstimateSameCurrency(s, in)
	if err != nil || got.TotalMicros != 660_000 || got.CrossBorderRuleVer != "geo_v2" || len(got.Breakdown) != 3 {
		t.Fatalf("cross-border=%+v, %v", got, err)
	}
	in.PaymentCurrency = "EUR"
	if _, err := EstimateSameCurrency(s, in); !errors.Is(err, ErrFXRequired) {
		t.Fatalf("FX silently invented: %v", err)
	}
}

func TestMicrosArithmeticRoundingAndOverflow(t *testing.T) {
	for _, tc := range []struct {
		value, numerator, denominator int64
		mode                          RoundingMode
		want                          int64
	}{
		{1, 1, 2, RoundDown, 0}, {1, 1, 2, RoundHalfUp, 1}, {1, 1, 3, RoundHalfUp, 0}, {1, 1, 3, RoundUp, 1},
		{0, 1, math.MaxInt64, RoundHalfUp, 0},
		{math.MaxInt64, 1, 10000, RoundDown, math.MaxInt64 / 10000},
	} {
		got, err := mulDiv(tc.value, tc.numerator, tc.denominator, tc.mode)
		if err != nil || got != tc.want {
			t.Fatalf("mulDiv(%+v)=%d, %v", tc, got, err)
		}
	}
	for _, mode := range []RoundingMode{RoundDown, RoundHalfUp, RoundUp} {
		for value := int64(0); value < 101; value++ {
			product := new(big.Int).Mul(big.NewInt(value), big.NewInt(37))
			q, r := new(big.Int), new(big.Int)
			q.QuoRem(product, big.NewInt(13), r)
			if (mode == RoundUp && r.Sign() > 0) || (mode == RoundHalfUp && r.Int64() >= 7) {
				q.Add(q, big.NewInt(1))
			}
			if got, err := mulDiv(value, 37, 13, mode); err != nil || got != q.Int64() {
				t.Fatalf("rounding %s/%d=%d,%v want %s", mode, value, got, err, q)
			}
		}
	}
	if _, err := mulDiv(math.MaxInt64, 2, 1, RoundDown); !errors.Is(err, ErrCostOverflow) {
		t.Fatal(err)
	}
	if _, err := mulDiv(1, 1, 1, "nearest"); !errors.Is(err, ErrInvalidCostInput) {
		t.Fatal(err)
	}
	if _, err := checkedAdd(math.MaxInt64, 1); !errors.Is(err, ErrCostOverflow) {
		t.Fatal(err)
	}
	s, in := feeFixture()
	in.AmountMinor = math.MaxInt64
	if _, err := EstimateSameCurrency(s, in); !errors.Is(err, ErrCostOverflow) {
		t.Fatal(err)
	}
	in.AmountMinor, in.CurrencyExponent = 1000, 7
	if _, err := EstimateSameCurrency(s, in); !errors.Is(err, ErrInvalidCostInput) {
		t.Fatal(err)
	}
}
