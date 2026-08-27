package performance

import (
	"errors"
	"math"
	"testing"
	"time"
)

func snapshotFixture() (Snapshot, Query, time.Time) {
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	return Snapshot{
		SchemaVersion: "performance.v1", SegmentSchemaVersion: "segment.v1", OutcomeDefinition: "authorization_approval.v1", Version: "perf_1",
		WindowStart: now.Add(-25 * time.Hour), WindowEnd: now.Add(-time.Hour), ObservedAt: now.Add(-30 * time.Minute), ProducedAt: now,
		MinSamples: [4]uint64{20, 20, 50, 100}, MinCoveragePPM: 950_000,
		MaxAge: time.Hour, MaxWindowAge: 2 * time.Hour, Buckets: map[SegmentKey]Bucket{},
	}, Query{CandidateID: "conn_a", ConnectionRevision: "conn_a_v1", Country: "DE", Network: "visa", Currency: "EUR"}, now
}

func TestFixedHierarchyNeverChoosesThePrettiestParentOrOtherConnection(t *testing.T) {
	s, q, now := snapshotFixture()
	s.Buckets[SegmentKey{"conn_a", "conn_a_v1", "DE", "visa", "EUR"}] = Bucket{Estimate: 1, SampleSize: 1}
	s.Buckets[SegmentKey{"conn_a", "conn_a_v1", "DE", any, "EUR"}] = Bucket{Estimate: 0.8, SampleSize: 100}
	s.Buckets[SegmentKey{"conn_a", "conn_a_v1", any, any, "EUR"}] = Bucket{Estimate: 0.99, SampleSize: 1000}
	s.Buckets[SegmentKey{"conn_b", "conn_b_v1", "DE", "visa", "EUR"}] = Bucket{Estimate: 1, SampleSize: 10000}
	got, err := s.Lookup(now, q)
	if err != nil || got.FallbackLevel != 1 || got.Estimate != 0.8 || got.CoveragePPM != 1_000_000 || got.SnapshotVersion != "perf_1" || got.SourceSegment != (SegmentKey{"conn_a", "conn_a_v1", "DE", any, "EUR"}) || !got.ObservedAt.Equal(s.ObservedAt) {
		t.Fatalf("fixed hierarchy=%+v, %v", got, err)
	}
	q.ConnectionRevision = "conn_a_v2"
	if _, err := s.Lookup(now, q); !errors.Is(err, ErrNoEvidence) {
		t.Fatalf("another connection revision became evidence: %v", err)
	}
}

func TestPendingCoverageUsesTheDeclaredThreshold(t *testing.T) {
	s, q, now := snapshotFixture()
	s.Buckets[SegmentKey{"conn_a", "conn_a_v1", "DE", "visa", "EUR"}] = Bucket{Estimate: 0.99, SampleSize: 100, PendingCount: 100}
	s.Buckets[SegmentKey{"conn_a", "conn_a_v1", "DE", any, "EUR"}] = Bucket{Estimate: 0.8, SampleSize: 100}
	got, err := s.Lookup(now, q)
	if err != nil || got.FallbackLevel != 1 || got.Estimate != 0.8 || got.CoveragePPM != 1_000_000 {
		t.Fatalf("incomplete exact bucket did not fall back: %+v,%v", got, err)
	}
	if got := coveragePPM(100, 100); got != 500_000 {
		t.Fatalf("coverage=%d", got)
	}
}

func TestLookupFreshnessAndTimeline(t *testing.T) {
	s, q, now := snapshotFixture()
	s.Buckets[SegmentKey{"conn_a", "conn_a_v1", any, any, any}] = Bucket{Estimate: 0.9, SampleSize: 100}
	if got, err := s.Lookup(now.Add(s.MaxAge), q); err != nil || got.FallbackLevel != 3 {
		t.Fatalf("max-age boundary=%+v,%v", got, err)
	}
	if _, err := s.Lookup(now.Add(s.MaxAge+time.Nanosecond), q); !errors.Is(err, ErrSnapshotStale) {
		t.Fatal(err)
	}
	for _, change := range []func(*Snapshot){
		func(v *Snapshot) { v.ProducedAt = now.Add(time.Second) },
		func(v *Snapshot) { v.ObservedAt = v.WindowEnd.Add(-time.Second) },
		func(v *Snapshot) { v.WindowStart = v.WindowEnd },
		func(v *Snapshot) { v.MaxAge = 0 },
		func(v *Snapshot) { v.MaxWindowAge = 0 },
		func(v *Snapshot) { v.MinSamples[2] = 0 },
		func(v *Snapshot) { v.MinCoveragePPM = 0 },
		func(v *Snapshot) { v.Version = "" },
	} {
		bad := s
		change(&bad)
		if _, err := bad.Lookup(now, q); !errors.Is(err, ErrInvalidSnapshot) {
			t.Fatalf("bad timeline/config accepted: %+v,%v", bad, err)
		}
	}
}

func TestShrinkageAndObservedIntervalsAreDifferentEvidence(t *testing.T) {
	got, err := SmoothedEstimate(1, 1, 0.8, 99)
	if err != nil || math.Abs(got.Probability-0.802) > 1e-12 || got.ObservedLower95 >= 0.21 || got.ObservedUpper95 < 0.999999 {
		t.Fatalf("one observation must not claim certainty: %+v,%v", got, err)
	}
	prior, err := SmoothedEstimate(0, 0, 0.9, 20)
	if err != nil || prior.Probability != 0.9 || prior.SampleSize != 0 || prior.ObservedLower95 != 0 || prior.ObservedUpper95 != 1 {
		t.Fatalf("prior=%+v,%v", prior, err)
	}
	for n := uint64(1); n <= 100; n++ {
		for s := uint64(0); s <= n; s++ {
			v, err := SmoothedEstimate(s, n, 0.9, 20)
			if err != nil || v.Probability < 0 || v.Probability > 1 || math.IsNaN(v.Probability) || v.ObservedLower95 < 0 || v.ObservedUpper95 > 1 || v.ObservedLower95 > v.ObservedUpper95 {
				t.Fatalf("invalid estimate %d/%d: %+v,%v", s, n, v, err)
			}
		}
	}
}

func TestNonfiniteAndInvalidEvidenceIsRejected(t *testing.T) {
	for _, values := range [][2]float64{{math.NaN(), 1}, {0.9, math.NaN()}, {math.Inf(1), 1}, {0.9, math.Inf(1)}, {-0.1, 1}, {1.1, 1}, {0.9, -1}} {
		if _, err := SmoothedEstimate(1, 2, values[0], values[1]); err == nil {
			t.Fatalf("accepted %v", values)
		}
	}
	if _, err := SmoothedEstimate(2, 1, 0.9, 1); err == nil {
		t.Fatal("successes exceeded trials")
	}
	if _, err := SmoothedEstimate(0, 0, 0.9, 0); err == nil {
		t.Fatal("missing evidence became a rate")
	}
}
