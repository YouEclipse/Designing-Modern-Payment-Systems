package model

import (
	"context"
	"math"
	"reflect"
	"testing"
	"time"
)

func TestUnusableFeatureFallsBackAsOneCompleteProfile(t *testing.T) {
	m, _, _, original, now := fixture(t)
	scorer := &Scorer{Version: m.id, Digest: m.digest, Probability: m.Probability}
	fallback := StaticProfile{Version: "static-v7", Digest: [32]byte{7},
		ExpiresAt: now.Add(time.Hour), Order: []string{"B@3", "A@7"}}
	for name, mutate := range map[string]func(*Vector){
		"unknown schema": func(v *Vector) { v.SchemaID = "unknown-schema" },
		"stale feature":  func(v *Vector) { v.ExpiresAt = now },
		"Inf feature":    func(v *Vector) { v.Values[0] = math.Inf(1) },
		"out of bounds":  func(v *Vector) { v.Values[0] = 2 },
		"unknown value":  func(v *Vector) { v.Known[0] = false },
	} {
		t.Run(name, func(t *testing.T) {
			bad := original
			bad.Values = append([]float64(nil), original.Values...)
			bad.Known = append([]bool(nil), original.Known...)
			mutate(&bad)
			allowed := []Candidate{{ID: "A@7", Vector: original}, {ID: "B@3", Vector: bad}}
			batch, err := ScoreBatch(context.Background(), allowed, scorer, fallback, now, 50*time.Millisecond)
			if err != nil || batch.Mode != "static" || batch.Version != fallback.Version ||
				batch.Digest != fallback.Digest || len(batch.Predictions) != 0 ||
				!reflect.DeepEqual(batch.Order, []string{"B@3", "A@7"}) {
				t.Fatalf("invalid feature preserved a partial model score: %+v, %v", batch, err)
			}
		})
	}
}

func TestFallbackMustBeApprovedUsableAndCoverAllowedSet(t *testing.T) {
	_, _, _, vector, now := fixture(t)
	allowed := []Candidate{{ID: "A@7", Vector: vector}}
	base := StaticProfile{Version: "static-v7", Digest: [32]byte{7},
		ExpiresAt: now.Add(time.Hour), Order: []string{"A@7"}}
	for name, mutate := range map[string]func(*StaticProfile){
		"unknown version": func(p *StaticProfile) { p.Version = "" },
		"no digest":       func(p *StaticProfile) { p.Digest = [32]byte{} },
		"expired":         func(p *StaticProfile) { p.ExpiresAt = now },
		"missing allowed": func(p *StaticProfile) { p.Order = []string{"B@3"} },
		"duplicate rank":  func(p *StaticProfile) { p.Order = []string{"A@7", "A@7"} },
	} {
		t.Run(name, func(t *testing.T) {
			profile := base
			mutate(&profile)
			if _, err := ScoreBatch(context.Background(), allowed, nil, profile, now, 50*time.Millisecond); err == nil {
				t.Fatal("missing model invented or repaired an unusable fallback")
			}
		})
	}
}

func TestExpiredModelUsesCurrentFallbackWithFreshFeatures(t *testing.T) {
	m, _, _, vector, createdAt := fixture(t)
	now := createdAt.Add(2 * time.Hour)
	vector.ExpiresAt = now.Add(time.Minute)
	scorer := &Scorer{Version: m.id, Digest: m.digest, Probability: m.Probability}
	fallback := StaticProfile{Version: "static-v7", Digest: [32]byte{7},
		ExpiresAt: now.Add(time.Hour), Order: []string{"A@7"}}
	batch, err := ScoreBatch(context.Background(), []Candidate{{ID: "A@7", Vector: vector}},
		scorer, fallback, now, 50*time.Millisecond)
	if err != nil || batch.Mode != "static" || batch.Version != fallback.Version || len(batch.Predictions) != 0 {
		t.Fatalf("expired model used fresh features to extend its authority: %+v, %v", batch, err)
	}
}
