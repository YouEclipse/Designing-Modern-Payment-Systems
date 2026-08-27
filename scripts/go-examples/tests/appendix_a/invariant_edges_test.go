package statespec

import (
	"math"
	"testing"
)

func capturedFixture() (Intent, Attempt, Operation) {
	identity := Identity{"connection_a", "revision_a", "namespace_a", "provider_a"}
	p := Intent{ID: "pi_1", Status: "succeeded", Integrity: "healthy", Currency: "USD", Winner: "attempt_1", Amount: 500, Authorized: 500, Captured: 500}
	a := Attempt{ID: "attempt_1", IntentID: p.ID, Status: "captured", PaymentReference: "payment_ref", Identity: identity}
	o := Operation{ID: "operation_1", IntentID: p.ID, AttemptID: a.ID, Kind: "sale", Status: "succeeded", Certainty: "confirmed", Currency: p.Currency, Reference: "operation_ref", IdempotencyKey: "key_1", Amount: p.Amount, Identity: identity}
	return p, a, o
}

func TestCapturedSaleRejectsIncompleteOrInconsistentSnapshots(t *testing.T) {
	for name, edit := range map[string]func(*Intent, *Attempt, *Operation){
		"missing intent":        func(p *Intent, _ *Attempt, _ *Operation) { p.ID = "" },
		"wrong intent state":    func(p *Intent, _ *Attempt, _ *Operation) { p.Status = "processing" },
		"unknown integrity":     func(p *Intent, _ *Attempt, _ *Operation) { p.Integrity = "" },
		"zero amount":           func(p *Intent, _ *Attempt, _ *Operation) { p.Amount = 0 },
		"negative refund":       func(p *Intent, _ *Attempt, _ *Operation) { p.Refunded = -1 },
		"excess refund":         func(p *Intent, _ *Attempt, _ *Operation) { p.Refunded = p.Captured + 1 },
		"authorization amount":  func(p *Intent, _ *Attempt, _ *Operation) { p.Authorized++ },
		"active after success":  func(p *Intent, _ *Attempt, _ *Operation) { p.Active = "attempt_1" },
		"different winner":      func(p *Intent, _ *Attempt, _ *Operation) { p.Winner = "attempt_other" },
		"no winner":             func(p *Intent, _ *Attempt, _ *Operation) { p.Winner = "" },
		"no attempt":            func(_ *Intent, a *Attempt, _ *Operation) { a.ID = "" },
		"unknown attempt":       func(_ *Intent, a *Attempt, _ *Operation) { a.Status = "unknown" },
		"no payment reference":  func(_ *Intent, a *Attempt, _ *Operation) { a.PaymentReference = "" },
		"no operation":          func(_ *Intent, _ *Attempt, o *Operation) { o.ID = "" },
		"foreign operation":     func(_ *Intent, _ *Attempt, o *Operation) { o.IntentID = "pi_other" },
		"foreign attempt link":  func(_ *Intent, _ *Attempt, o *Operation) { o.AttemptID = "attempt_other" },
		"unconfirmed effect":    func(_ *Intent, _ *Attempt, o *Operation) { o.Certainty = "possible" },
		"operation amount":      func(_ *Intent, _ *Attempt, o *Operation) { o.Amount++ },
		"operation currency":    func(_ *Intent, _ *Attempt, o *Operation) { o.Currency = "EUR" },
		"missing currency":      func(p *Intent, _ *Attempt, o *Operation) { p.Currency, o.Currency = "", "" },
		"missing provider key":  func(_ *Intent, _ *Attempt, o *Operation) { o.IdempotencyKey = "" },
		"connection mismatch":   func(_ *Intent, _ *Attempt, o *Operation) { o.Identity.ConnectionID = "connection_b" },
		"namespace mismatch":    func(_ *Intent, _ *Attempt, o *Operation) { o.Identity.NamespaceID = "namespace_b" },
		"provider mismatch":     func(_ *Intent, _ *Attempt, o *Operation) { o.Identity.Provider = "provider_b" },
		"missing connection":    func(_ *Intent, a *Attempt, o *Operation) { a.Identity.ConnectionID, o.Identity.ConnectionID = "", "" },
		"missing revision":      func(_ *Intent, a *Attempt, o *Operation) { a.Identity.Revision, o.Identity.Revision = "", "" },
		"missing namespace":     func(_ *Intent, a *Attempt, o *Operation) { a.Identity.NamespaceID, o.Identity.NamespaceID = "", "" },
		"missing provider kind": func(_ *Intent, a *Attempt, o *Operation) { a.Identity.Provider, o.Identity.Provider = "", "" },
	} {
		t.Run(name, func(t *testing.T) {
			p, a, o := capturedFixture()
			edit(&p, &a, &o)
			beforeP, beforeA, beforeO := p, a, o
			if err := ValidateCapturedSale(p, a, o); err == nil {
				t.Fatal("invalid captured-sale projection accepted")
			}
			if p != beforeP || a != beforeA || o != beforeO {
				t.Fatal("checker changed financial evidence")
			}
		})
	}
}

func TestCapturedSalePreservesHistoricalCaptureThroughLegalRefundTotals(t *testing.T) {
	// This is the reference invariant's projection contract, not a claim that the
	// current Payment API implements refunds or turns success into failure.
	for _, integrity := range []string{"healthy", "review_required"} {
		for _, total := range []int64{1, 500, math.MaxInt64} {
			for _, refunded := range []int64{0, total / 2, total} {
				p, a, o := capturedFixture()
				p.Amount, p.Authorized, p.Captured, p.Refunded = total, total, total, refunded
				p.Integrity, o.Amount = integrity, total
				if err := ValidateCapturedSale(p, a, o); err != nil {
					t.Fatalf("lost historical capture: integrity=%s total=%d refunded=%d: %v", integrity, total, refunded, err)
				}
			}
		}
	}
}
