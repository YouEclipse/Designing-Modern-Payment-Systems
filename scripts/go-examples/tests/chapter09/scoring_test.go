package scoring

import (
	"errors"
	"math"
	"testing"
)

func scoreFixture() (Profile, Input) {
	return Profile{Version: "score_v1", SuccessWeightBPS: 7000, CostWeightBPS: 3000, CostFloorMicros: 100, CostCeilingMicros: 1100}, Input{
		CandidateID: "conn_a", SuccessState: FeatureKnown, ConservativePPM: 900_000, PerformanceVer: "perf_1",
		CostState: FeatureKnown, SucceededCostMicros: 600, CostCurrency: "USD", CostArtifactVer: "fee_1",
	}
}

func TestFixedPointsAndContributionsRemainExplainable(t *testing.T) {
	p, in := scoreFixture()
	got, err := Score(p, in)
	if err != nil || got.SuccessPoints != 9000 || got.CostPoints != 5000 || got.SuccessContribution != 6300 || got.CostContribution != 1500 || got.TotalPoints != 7800 || got.ProfileVersion != p.Version || got.PerformanceVersion != in.PerformanceVer || got.CostArtifactVersion != in.CostArtifactVer {
		t.Fatalf("breakdown=%+v,%v", got, err)
	}
	for _, cost := range []int64{0, 100, 600, 1100, math.MaxInt64} {
		for weight := int64(0); weight <= 10000; weight += 100 {
			p.SuccessWeightBPS, p.CostWeightBPS = weight, 10000-weight
			in.SucceededCostMicros = cost
			v, err := Score(p, in)
			if err != nil || v.TotalPoints < 0 || v.TotalPoints > pointScale || v.TotalPoints != v.SuccessContribution+v.CostContribution || v.CostPoints < 0 || v.CostPoints > pointScale {
				t.Fatalf("bounded score=%+v,%v", v, err)
			}
		}
	}
}

func TestMissingInvalidAndOverflowAreNotZeroCost(t *testing.T) {
	p, in := scoreFixture()
	for _, change := range []func(*Input){
		func(v *Input) { v.CandidateID = "" }, func(v *Input) { v.CostState = "missing" }, func(v *Input) { v.SuccessState = "stale" },
		func(v *Input) { v.ConservativePPM = -1 }, func(v *Input) { v.ConservativePPM = probabilityPPM + 1 }, func(v *Input) { v.SucceededCostMicros = -1 },
		func(v *Input) { v.CostCurrency = "" }, func(v *Input) { v.PerformanceVer = "" }, func(v *Input) { v.CostArtifactVer = "" },
	} {
		bad := in
		change(&bad)
		if _, err := Score(p, bad); !errors.Is(err, ErrInvalidScoreInput) {
			t.Fatalf("input accepted: %+v,%v", bad, err)
		}
	}
	for _, change := range []func(*Profile){
		func(v *Profile) { v.SuccessWeightBPS = -1 }, func(v *Profile) { v.CostWeightBPS = 0 },
		func(v *Profile) { v.CostFloorMicros = -1 }, func(v *Profile) { v.CostCeilingMicros = v.CostFloorMicros },
		func(v *Profile) { v.CostCeilingMicros = math.MaxInt64 }, func(v *Profile) { v.Version = "" },
	} {
		bad := p
		change(&bad)
		if _, err := Score(bad, in); !errors.Is(err, ErrInvalidScoreInput) {
			t.Fatalf("profile accepted: %+v,%v", bad, err)
		}
	}
}
