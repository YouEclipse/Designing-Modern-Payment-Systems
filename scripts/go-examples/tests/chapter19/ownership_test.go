package regionownership

import (
	"errors"
	"testing"
	"time"
)

func TestOwnershipGateRejectsStaleOrUnusablePermission(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	base := Lease{IntentID: "intent-1", Region: "eu", Epoch: 42,
		ValidUntil: now.Add(time.Minute), Mode: "active"}
	permit := Permit{IntentID: base.IntentID, Region: base.Region, Epoch: base.Epoch}
	if err := Check(base, permit, now, true); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*Lease, *Permit){
		"old owner":      func(_ *Lease, p *Permit) { p.Epoch = 41 },
		"future epoch":   func(_ *Lease, p *Permit) { p.Epoch = 43 },
		"other region":   func(_ *Lease, p *Permit) { p.Region = "us" },
		"other intent":   func(_ *Lease, p *Permit) { p.IntentID = "intent-2" },
		"no epoch":       func(_ *Lease, p *Permit) { p.Epoch = 0 },
		"lease boundary": func(l *Lease, _ *Permit) { l.ValidUntil = now },
		"frozen":         func(l *Lease, _ *Permit) { l.Mode = "frozen" },
		"unknown mode":   func(l *Lease, _ *Permit) { l.Mode = "unknown" },
		"recovery only":  func(l *Lease, _ *Permit) { l.Mode = "recovery_only" },
	} {
		t.Run(name, func(t *testing.T) {
			lease, candidate := base, permit
			mutate(&lease, &candidate)
			if err := Check(lease, candidate, now, true); !errors.Is(err, ErrFenced) {
				t.Fatalf("new funds accepted without current active ownership: %v", err)
			}
		})
	}
	base.Mode = "recovery_only"
	if err := Check(base, permit, now, false); err != nil {
		t.Fatal("current recovery owner must be allowed to handle an existing operation")
	}
	permit.Epoch--
	if err := Check(base, permit, now, false); !errors.Is(err, ErrFenced) {
		t.Fatal("recovery mode must not authorize the old owner")
	}
}
