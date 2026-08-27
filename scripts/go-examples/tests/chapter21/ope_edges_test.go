package ope

import (
	"context"
	"errors"
	"math"
	"testing"
)

type contextualReplay map[string]PolicyPair

func (r contextualReplay) FullDistributions(_ context.Context, row Row, _ string) (PolicyPair, error) {
	if pair, ok := r[row.ContextID]; ok {
		return pair, nil
	}
	return PolicyPair{}, errors.New("complete policy witness unavailable")
}

func TestEveryContextNeedsFullTargetSupport(t *testing.T) {
	rows := []Row{
		{UnitID: "intent-1", ContextID: "context-1", LoggingPolicy: "old", Action: "A@7",
			LabelID: "capture-h-v1", Propensity: 1, Reward: 1, Mature: true},
		{UnitID: "intent-2", ContextID: "context-2", LoggingPolicy: "old", Action: "A@7",
			LabelID: "capture-h-v1", Propensity: 1, Reward: 1, Mature: true},
	}
	replay := contextualReplay{
		"context-1": {ContextID: "context-1", LoggingPolicy: "old", TargetPolicy: "new",
			Logging: map[string]float64{"A@7": 1, "B@3": 0}, Target: map[string]float64{"A@7": 1, "B@3": 0}},
		"context-2": {ContextID: "context-2", LoggingPolicy: "old", TargetPolicy: "new",
			Logging: map[string]float64{"A@7": 1, "B@3": 0}, Target: map[string]float64{"A@7": 0, "B@3": 1}},
	}
	plan := Plan{TargetPolicy: "new", LabelID: "capture-h-v1", MinPropensity: 0.01, MaxWeight: 100, MinimumESS: 1}
	if report, err := Estimate(context.Background(), replay, rows, plan); !errors.Is(err, ErrInsufficientEvidence) || report.Status != "" {
		t.Fatalf("one supported context masked a second unsupported context: %+v, %v", report, err)
	}
}

func TestIncompleteOrUnboundWitnessCannotProduceValue(t *testing.T) {
	row := Row{UnitID: "intent-1", ContextID: "context-1", LoggingPolicy: "old", Action: "A@7",
		LabelID: "capture-h-v1", Propensity: 1, Reward: 1, Mature: true}
	base := PolicyPair{ContextID: row.ContextID, LoggingPolicy: row.LoggingPolicy, TargetPolicy: "new",
		Logging: map[string]float64{"A@7": 1}, Target: map[string]float64{"A@7": 1}}
	plan := Plan{TargetPolicy: "new", LabelID: row.LabelID, MinPropensity: 0.01, MaxWeight: 100, MinimumESS: 1}
	for name, mutate := range map[string]func(*PolicyPair){
		"wrong context":  func(p *PolicyPair) { p.ContextID = "context-other" },
		"wrong logging":  func(p *PolicyPair) { p.LoggingPolicy = "old-but-different" },
		"wrong target":   func(p *PolicyPair) { p.TargetPolicy = "different-proposal" },
		"missing policy": func(p *PolicyPair) { p.Logging = nil },
		"NaN probability": func(p *PolicyPair) {
			p.Target = map[string]float64{"A@7": math.NaN()}
		},
		"not normalized": func(p *PolicyPair) { p.Target = map[string]float64{"A@7": 0.4} },
	} {
		t.Run(name, func(t *testing.T) {
			pair := base
			mutate(&pair)
			report, err := Estimate(context.Background(), contextualReplay{row.ContextID: pair}, []Row{row}, plan)
			if err == nil || report.Status != "" {
				t.Fatalf("untrusted witness produced a report: %+v, %v", report, err)
			}
		})
	}
	if _, err := Estimate(context.Background(), contextualReplay{}, []Row{row}, plan); err == nil {
		t.Fatal("chosen propensity alone replaced a missing complete witness")
	}
}

func TestRareSupportedActionStillNeedsWeightBudget(t *testing.T) {
	row := Row{UnitID: "intent-1", ContextID: "context-1", LoggingPolicy: "old", Action: "A@7",
		LabelID: "capture-h-v1", Propensity: 0.999, Reward: 1, Mature: true}
	pair := PolicyPair{ContextID: row.ContextID, LoggingPolicy: row.LoggingPolicy, TargetPolicy: "new",
		Logging: map[string]float64{"A@7": 0.999, "B@3": 0.001},
		Target:  map[string]float64{"A@7": 0, "B@3": 1}}
	plan := Plan{TargetPolicy: "new", LabelID: row.LabelID, MinPropensity: 0.001, MaxWeight: 100, MinimumESS: 1}
	if _, err := Estimate(context.Background(), contextualReplay{row.ContextID: pair}, []Row{row}, plan); err == nil {
		t.Fatal("unobserved high-weight target action bypassed the full-distribution weight gate")
	}
}
