package capacityplan

import (
	"math"
	"reflect"
	"testing"
)

func caseStudyInput() Input {
	return Input{
		NewIntentsPS: 100000, MerchantRequestsPerIntent: 2.2, FallbackAttemptsPerIntent: 0.03,
		FallbackGuardrailChecksPerIntent: 0.03,
		SameKeyResubmitsPerOperation: 0.01, UnknownFraction: 0.05, QueriesPerUnknown: 3,
		WebhookEventsPerOperation: 1.1, ProviderMeanSeconds: 0.25, DecisionBytes: 2048,
		OutboxOutageSeconds: 600, CatchupSeconds: 3600, GatewayCallsCapacityPerRegion: 85000,
		TargetUtilization: 0.65, Regions: 3, LostRegions: 1,
	}
}

func TestCapacityReproducesEveryCaseStudyQuantity(t *testing.T) {
	input := caseStudyInput()
	before := input
	got, err := Calculate(input)
	if err != nil {
		t.Fatal(err)
	}
	if input != before || got.ModelVersion != "card-sale-capacity-v1" || got.WithinArithmeticBudget {
		t.Fatalf("changed assumptions or approved the under-capacity plan: %+v", got)
	}
	for _, tc := range []struct {
		name      string
		got, want float64
	}{
		{"merchant requests/s", got.MerchantRequestsPS, 220000},
		{"logical operations/s", got.LogicalOperationsPS, 103000},
		{"fallback guardrail checks/s", got.GuardrailRevalidationsPS, 3000},
		{"execute calls/s", got.ExecuteCallsPS, 104030},
		{"query calls/s", got.QueryCallsPS, 15450},
		{"provider calls/s", got.ProviderCallsPS, 119480},
		{"webhook events/s", got.WebhookEventsPS, 113300},
		{"calls/s/surviving region", got.CallsPerSurvivingRegionPS, 59740},
		{"minimum capacity/s/region", got.MinimumCapacityPerRegionPS, 91908},
		{"post-loss utilization", got.PostLossUtilization, 59740.0 / 85000},
		{"mean inflight/region", got.MeanInflightPerSurvivingRegion, 14935},
		{"decision bytes/s", got.DecisionBytesPS, 210944000},
		{"backlog events", got.BacklogEvents, 61800000},
		{"projector events/s", got.RequiredProjectorEventsPS, 120166.66666666667},
	} {
		t.Run(tc.name, func(t *testing.T) { assertNear(t, tc.got, tc.want) })
	}
}

func TestCapacityChecksFaultHeadroomInsteadOfNominalLimit(t *testing.T) {
	input := caseStudyInput()
	withLoss, err := Calculate(input)
	if err != nil {
		t.Fatal(err)
	}
	input.LostRegions = 0
	steady, err := Calculate(input)
	if err != nil {
		t.Fatal(err)
	}
	if !steady.WithinArithmeticBudget || withLoss.WithinArithmeticBudget || withLoss.PostLossUtilization >= 1 {
		t.Fatalf("steady and fault budgets confused: steady=%+v loss=%+v", steady, withLoss)
	}
	input.LostRegions = 1
	input.GatewayCallsCapacityPerRegion = withLoss.MinimumCapacityPerRegionPS
	provisioned, err := Calculate(input)
	if err != nil || !provisioned.WithinArithmeticBudget {
		t.Fatalf("rounded-up required capacity did not satisfy arithmetic: %+v %v", provisioned, err)
	}
	input.LostRegions = 2
	remainingOne, err := Calculate(input)
	if err != nil || remainingOne.WithinArithmeticBudget {
		t.Fatalf("one surviving region was silently treated as two: %+v %v", remainingOne, err)
	}
}

func TestCapacityRetainsRecoveryAndBacklogWork(t *testing.T) {
	input := caseStudyInput()
	baseline, err := Calculate(input)
	if err != nil {
		t.Fatal(err)
	}
	input.UnknownFraction = 0.5
	degraded, err := Calculate(input)
	if err != nil {
		t.Fatal(err)
	}
	if degraded.LogicalOperationsPS != baseline.LogicalOperationsPS || degraded.ExecuteCallsPS != baseline.ExecuteCallsPS ||
		degraded.QueryCallsPS <= baseline.QueryCallsPS || degraded.ProviderCallsPS <= baseline.ProviderCallsPS ||
		degraded.MeanInflightPerSurvivingRegion <= baseline.MeanInflightPerSurvivingRegion {
		t.Fatalf("query amplification lost or turned into new charges: baseline=%+v degraded=%+v", baseline, degraded)
	}
	input = caseStudyInput()
	input.OutboxOutageSeconds = 0
	noBacklog, err := Calculate(input)
	if err != nil || noBacklog.BacklogEvents != 0 || noBacklog.RequiredProjectorEventsPS != noBacklog.LogicalOperationsPS {
		t.Fatalf("zero outage invented catch-up work: %+v %v", noBacklog, err)
	}
	input.FallbackAttemptsPerIntent = 1
	input.FallbackGuardrailChecksPerIntent = 1
	input.SameKeyResubmitsPerOperation, input.UnknownFraction = 0, 0
	oneSuccessor, err := Calculate(input)
	if err != nil || oneSuccessor.LogicalOperationsPS != 200000 || oneSuccessor.ProviderCallsPS != 200000 {
		t.Fatalf("one-successor boundary is inconsistent: %+v %v", oneSuccessor, err)
	}
}

func TestCapacityRejectsNonFiniteOrNegativeInputs(t *testing.T) {
	inputType := reflect.TypeOf(Input{})
	for field := 0; field < inputType.NumField(); field++ {
		if inputType.Field(field).Type.Kind() != reflect.Float64 {
			continue
		}
		for name, value := range map[string]float64{
			"NaN": math.NaN(), "positive infinity": math.Inf(1), "negative infinity": math.Inf(-1), "negative": -1,
		} {
			t.Run(inputType.Field(field).Name+"/"+name, func(t *testing.T) {
				input := caseStudyInput()
				reflect.ValueOf(&input).Elem().Field(field).SetFloat(value)
				if got, err := Calculate(input); err == nil || got != (Plan{}) {
					t.Fatalf("invalid input accepted: %+v %v", got, err)
				}
			})
		}
	}
}

func TestCapacityRejectsInvalidFractionsTopologyAndZeroDenominators(t *testing.T) {
	for name, edit := range map[string]func(*Input){
		"zero demand":        func(i *Input) { i.NewIntentsPS = 0 },
		"too few API calls":  func(i *Input) { i.MerchantRequestsPerIntent = 0.99 },
		"two successors":     func(i *Input) { i.FallbackAttemptsPerIntent = 1.01 },
		"missing guardrail":  func(i *Input) { i.FallbackGuardrailChecksPerIntent = 0.02 },
		"invalid unknown":    func(i *Input) { i.UnknownFraction = 1.01 },
		"zero mean time":     func(i *Input) { i.ProviderMeanSeconds = 0 },
		"zero bytes":         func(i *Input) { i.DecisionBytes = 0 },
		"zero catchup":       func(i *Input) { i.CatchupSeconds = 0 },
		"zero capacity":      func(i *Input) { i.GatewayCallsCapacityPerRegion = 0 },
		"zero utilization":   func(i *Input) { i.TargetUtilization = 0 },
		"no reserve":         func(i *Input) { i.TargetUtilization = 1 },
		"excess utilization": func(i *Input) { i.TargetUtilization = 1.01 },
		"zero regions":       func(i *Input) { i.Regions = 0 },
		"negative regions":   func(i *Input) { i.Regions = -1 },
		"negative losses":    func(i *Input) { i.LostRegions = -1 },
		"all regions lost":   func(i *Input) { i.LostRegions = i.Regions },
		"too many lost":      func(i *Input) { i.LostRegions = i.Regions + 1 },
	} {
		t.Run(name, func(t *testing.T) {
			input := caseStudyInput()
			edit(&input)
			if got, err := Calculate(input); err == nil || got != (Plan{}) {
				t.Fatalf("invalid assumptions accepted: %+v %v", got, err)
			}
		})
	}
}

func TestCapacityRejectsDerivedOverflow(t *testing.T) {
	for name, edit := range map[string]func(*Input){
		"demand":   func(i *Input) { i.NewIntentsPS = math.MaxFloat64 },
		"inflight": func(i *Input) { i.ProviderMeanSeconds = math.MaxFloat64 },
		"bytes":    func(i *Input) { i.DecisionBytes = math.MaxFloat64 },
		"backlog":  func(i *Input) { i.OutboxOutageSeconds = math.MaxFloat64 },
		"catchup":  func(i *Input) { i.CatchupSeconds = math.SmallestNonzeroFloat64 },
		"capacity": func(i *Input) { i.GatewayCallsCapacityPerRegion = math.SmallestNonzeroFloat64 },
		"queries":  func(i *Input) { i.QueriesPerUnknown = math.MaxFloat64 },
		"headroom": func(i *Input) { i.TargetUtilization = math.SmallestNonzeroFloat64 },
	} {
		t.Run(name, func(t *testing.T) {
			input := caseStudyInput()
			edit(&input)
			if got, err := Calculate(input); err == nil || got != (Plan{}) {
				t.Fatalf("overflow became a capacity conclusion: %+v %v", got, err)
			}
		})
	}
}

func assertNear(t *testing.T, got, want float64) {
	t.Helper()
	if math.IsNaN(got) || math.IsInf(got, 0) || math.Abs(got-want) > math.Max(1e-9, math.Abs(want)*1e-12) {
		t.Fatalf("got %.15g, want %.15g", got, want)
	}
}
