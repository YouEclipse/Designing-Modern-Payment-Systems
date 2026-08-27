package launchgate

import (
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

func completeEvidenceFixture() (Scope, time.Time, []Evidence) {
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	target := Scope{"implementation-sha", "bundle-digest", "live", "three-region-peak"}
	names := []string{
		"code_and_contracts", "postgres_integration", "funds_invariants",
		"configuration_rollback", "region_fencing", "fault_capacity",
		"security_boundary", "provider_live_approval",
	}
	items := make([]Evidence, 0, len(names))
	for _, name := range names {
		items = append(items, Evidence{Name: name, State: "pass", Owner: "release-owner", Artifact: "artifact:" + name, Scope: target, CheckedAt: now})
	}
	return target, now, items
}

func TestGlobalLiveGateRequiresEachExplicitPass(t *testing.T) {
	target, now, complete := completeEvidenceFixture()
	for i, item := range complete {
		t.Run(item.Name, func(t *testing.T) {
			missing := append([]Evidence(nil), complete[:i]...)
			missing = append(missing, complete[i+1:]...)
			got := EvaluateGlobalLive(target, now, time.Hour, missing)
			if got.Go || !slices.Contains(got.Reasons, item.Name+":missing") {
				t.Fatalf("missing evidence silently waived: %+v", got)
			}
			for _, state := range []string{"", "fail", "unknown", "not_run", "not_applicable", "PASS"} {
				items := append([]Evidence(nil), complete...)
				items[i].State = state
				got := EvaluateGlobalLive(target, now, time.Hour, items)
				if got.Go || !slices.Contains(got.Reasons, item.Name+":not_passed") {
					t.Fatalf("unapproved state %q became pass: %+v", state, got)
				}
			}
		})
	}
	if got := EvaluateGlobalLive(target, now, time.Hour, nil); got.Go || len(got.Reasons) != len(complete) {
		t.Fatalf("empty evidence is not a release approval: %+v", got)
	}
}

func TestGlobalLiveGateBindsEvidenceToItsWholeScope(t *testing.T) {
	target, now, complete := completeEvidenceFixture()
	for name, edit := range map[string]func(*Evidence){
		"owner":       func(e *Evidence) { e.Owner = "" },
		"artifact":    func(e *Evidence) { e.Artifact = "" },
		"code":        func(e *Evidence) { e.Scope.CodeSHA = "other" },
		"bundle":      func(e *Evidence) { e.Scope.BundleDigest = "other" },
		"environment": func(e *Evidence) { e.Scope.Environment = "test" },
		"workload":    func(e *Evidence) { e.Scope.WorkloadID = "single-region-low-load" },
	} {
		t.Run(name, func(t *testing.T) {
			items := append([]Evidence(nil), complete...)
			edit(&items[0])
			if got := EvaluateGlobalLive(target, now, time.Hour, items); got.Go || !slices.Contains(got.Reasons, items[0].Name+":unbound_evidence") {
				t.Fatalf("evidence from another release scope accepted: %+v", got)
			}
		})
	}
}

func TestGlobalLiveGateRejectsInvalidTargetsAndTimes(t *testing.T) {
	target, now, complete := completeEvidenceFixture()
	for name, edit := range map[string]func(*Scope){
		"code":        func(s *Scope) { s.CodeSHA = "" },
		"bundle":      func(s *Scope) { s.BundleDigest = "" },
		"environment": func(s *Scope) { s.Environment = "local" },
		"workload":    func(s *Scope) { s.WorkloadID = "" },
	} {
		t.Run(name, func(t *testing.T) {
			invalid := target
			edit(&invalid)
			if got := EvaluateGlobalLive(invalid, now, time.Hour, complete); got.Go {
				t.Fatalf("invalid target passed: %+v", got)
			}
		})
	}
	for _, age := range []time.Duration{0, -time.Hour} {
		if got := EvaluateGlobalLive(target, now, age, complete); got.Go {
			t.Fatalf("invalid age budget passed: %+v", got)
		}
	}
	if got := EvaluateGlobalLive(target, time.Time{}, time.Hour, complete); got.Go {
		t.Fatalf("zero current time accepted: %+v", got)
	}
	for _, tc := range []struct {
		name string
		when time.Time
		ok   bool
	}{
		{"exact freshness boundary", now.Add(-time.Hour), true},
		{"expired by one nanosecond", now.Add(-time.Hour - time.Nanosecond), false},
		{"future by one nanosecond", now.Add(time.Nanosecond), false},
		{"zero timestamp", time.Time{}, false},
		{"same instant in other zone", now.In(time.FixedZone("test", 8*60*60)), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			items := append([]Evidence(nil), complete...)
			items[0].CheckedAt = tc.when
			if got := EvaluateGlobalLive(target, now, time.Hour, items); got.Go != tc.ok {
				t.Fatalf("freshness result=%+v, want Go=%v", got, tc.ok)
			}
		})
	}
}

func TestGlobalLiveGateRetainsAllFailuresAndDoesNotMutateEvidence(t *testing.T) {
	target, now, items := completeEvidenceFixture()
	items[0].State = "fail"
	items[4].Owner = ""
	items[6].CheckedAt = now.Add(-2 * time.Hour)
	items = append(items, items[0])
	items = append(items, Evidence{Name: "unrecognized-marker-secret", State: "pass"})
	before := append([]Evidence(nil), items...)
	got := EvaluateGlobalLive(target, now, time.Hour, items)
	if got.Go || len(got.Reasons) != 5 || !slices.IsSorted(got.Reasons) || !reflect.DeepEqual(items, before) {
		t.Fatalf("failures discarded or evidence modified: %+v", got)
	}
	for _, reason := range got.Reasons {
		if strings.Contains(reason, "marker-secret") {
			t.Fatal("unknown gate label leaked into the stable reason vocabulary")
		}
	}
	slices.Reverse(items)
	reversed := EvaluateGlobalLive(target, now, time.Hour, items)
	if !reflect.DeepEqual(got, reversed) {
		t.Fatalf("reordering equivalent evidence changed the decision: %+v / %+v", got, reversed)
	}
	tooMany := make([]Evidence, 65)
	if got := EvaluateGlobalLive(target, now, time.Hour, tooMany); got.Go || !slices.Contains(got.Reasons, "invalid_target_or_evidence") {
		t.Fatalf("oversized report accepted: %+v", got)
	}
}

func TestGlobalLiveGateCanReadSharedEvidenceConcurrently(t *testing.T) {
	target, now, items := completeEvidenceFixture()
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if got := EvaluateGlobalLive(target, now, time.Hour, items); !got.Go {
				t.Errorf("concurrent evaluation failed: %+v", got)
			}
		}()
	}
	wg.Wait()
}
