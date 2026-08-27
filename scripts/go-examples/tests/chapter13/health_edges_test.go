package health

import (
	"math"
	"testing"
	"time"
)

func healthFixture() (time.Time, Key, Config) {
	return time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC), Key{CandidateID: "cand_eu", ConnectionRevision: "rev7", Operation: "sale", FaultDomain: "provider_api_eu"}, Config{MinSamples: 20, HalfOpenSamples: 3, DegradedRate: 0.2, OpenRate: 0.5, RecoverRate: 0.1, OpenFor: time.Minute, SnapshotTTL: 2 * time.Minute}
}

func TestTechnicalSignalClassificationDoesNotMixBusinessWithFunds(t *testing.T) {
	for _, tc := range []struct {
		class           SignalClass
		sample, failure bool
	}{
		{SignalTechnicalSuccess, true, false}, {SignalTechnicalFailure, true, true}, {SignalBusinessResponse, true, false},
		{SignalIgnoredBusiness, false, false}, {SignalLocalRejected, false, false}, {"unclassified", false, false},
	} {
		sample, failure := TechnicalEffect(tc.class)
		if sample != tc.sample || failure != tc.failure {
			t.Fatalf("%s sample=%v failure=%v", tc.class, sample, failure)
		}
	}
}

func TestMinimumSamplesHysteresisAndSlowPressure(t *testing.T) {
	now, key, cfg := healthFixture()
	for _, tc := range []struct {
		previous                State
		samples, failures, slow uint64
		want                    State
	}{
		{"", 1, 1, 0, StateUnknown}, {"", 20, 0, 0, StateHealthy}, {"", 20, 4, 0, StateDegraded},
		{StateDegraded, 20, 3, 0, StateDegraded}, {StateDegraded, 20, 2, 0, StateHealthy},
		{"", 20, 10, 0, StateUnavailable}, {"", 20, 0, 10, StateUnavailable},
	} {
		got := Next(now, key, Snapshot{Key: key, State: tc.previous}, Aggregate{Samples: tc.samples, Failures: tc.failures, Slow: tc.slow, LatestAt: now}, cfg, "health_v1")
		if got.State != tc.want {
			t.Fatalf("%+v -> %+v", tc, got)
		}
	}
}

func TestOpenExpiresToRecoveringAndFirstFailedProbeImmediatelyReopens(t *testing.T) {
	now, key, cfg := healthFixture()
	open := Snapshot{Key: key, State: StateUnavailable, OpenUntil: now.Add(time.Minute)}
	window := Aggregate{Samples: 20, LatestAt: now}
	if got := Next(now, key, open, window, cfg, "v2"); got.State != StateUnavailable || got.OpenUntil != open.OpenUntil {
		t.Fatalf("open duration ignored: %+v", got)
	}
	recovering := Next(open.OpenUntil, key, open, window, cfg, "v3")
	if recovering.State != StateRecovering {
		t.Fatalf("timer cannot directly make healthy: %+v", recovering)
	}
	for _, window := range []Aggregate{{Samples: 1, Failures: 1, LatestAt: now}, {Samples: 1, Slow: 1, LatestAt: now}} {
		got := Next(now, key, recovering, window, cfg, "v4")
		if got.State != StateUnavailable || !got.OpenUntil.Equal(now.Add(cfg.OpenFor)) {
			t.Fatalf("first failed probe must reopen before minimum success count: %+v", got)
		}
	}
	if got := Next(now, key, recovering, Aggregate{Samples: 2, LatestAt: now}, cfg, "v5"); got.State != StateRecovering {
		t.Fatal(got)
	}
	if got := Next(now, key, recovering, Aggregate{Samples: 3, LatestAt: now}, cfg, "v6"); got.State != StateHealthy {
		t.Fatal(got)
	}
}

func TestNoNewEvidenceCannotRefreshSnapshotAndKeysCannotBleed(t *testing.T) {
	now, key, cfg := healthFixture()
	window := Aggregate{Samples: 20, LatestAt: now}
	healthy := Next(now, key, Snapshot{}, window, cfg, "v1")
	for _, at := range []time.Time{now.Add(cfg.SnapshotTTL), now.Add(24 * time.Hour)} {
		republished := Next(at, key, healthy, window, cfg, "v2")
		if republished.ValidUntil != healthy.ValidUntil || republished.RoutingState(at) != StateUnknown {
			t.Fatalf("empty tick renewed old evidence: %+v", republished)
		}
	}
	for _, change := range []func(*Key){func(k *Key) { k.CandidateID = "cand_us" }, func(k *Key) { k.ConnectionRevision = "rev8" }, func(k *Key) { k.Operation = "query" }, func(k *Key) { k.FaultDomain = "egress_az_a" }} {
		other := key
		change(&other)
		got := Next(now, other, healthy, window, cfg, "v3")
		if got.State != StateUnknown {
			t.Fatalf("previous state from another fault domain was reused: %+v", got)
		}
	}
	bad := cfg
	bad.OpenRate = math.NaN()
	if got := Next(now, key, Snapshot{}, window, bad, "v4"); got.State != StateUnknown {
		t.Fatal(got)
	}
	window.Failures = window.Samples + 1
	if got := Next(now, key, Snapshot{}, window, cfg, "v5"); got.State != StateUnknown {
		t.Fatal(got)
	}
}

func TestInvalidPreviousStateOrUnavailableWindowFailsClosed(t *testing.T) {
	now, key, cfg := healthFixture()
	window := Aggregate{Samples: 20, LatestAt: now}
	for _, previous := range []Snapshot{
		{Key: key, State: "healthy_typo"},
		{Key: key, State: StateUnavailable},
		{Key: key, State: StateUnavailable, ObservedAt: now, OpenUntil: now.Add(-time.Second)},
	} {
		got := Next(now, key, previous, window, cfg, "v_invalid")
		if got.State != StateUnknown || got.Confidence != "low" {
			t.Fatalf("invalid previous snapshot was trusted: previous=%+v got=%+v", previous, got)
		}
	}
}
