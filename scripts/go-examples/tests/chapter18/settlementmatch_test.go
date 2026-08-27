package settlementmatch

import (
	"reflect"
	"testing"
)

func matchingFixture() (Line, Operation) {
	line := Line{
		ConnectionID: "conn_a", ConnectionRevision: "revision_a", NamespaceID: "namespace_a",
		ProviderReference: "provider_reference_1", Kind: "captured_sale", BasisAmountMinor: 10000, BasisCurrency: "USD",
	}
	operation := Operation{
		ID: "op_1", IntentID: "pi_1", AttemptID: "attempt_1", DecisionID: "decision_1",
		ConnectionID: line.ConnectionID, ConnectionRevision: line.ConnectionRevision,
		NamespaceID: line.NamespaceID, ProviderReference: line.ProviderReference,
		AmountMinor: 10000, Currency: "USD", Kind: "sale", Status: "succeeded", Integrity: "healthy",
	}
	return line, operation
}

func TestSettlementMatchUsesNamespaceBeforeAmountAndConnection(t *testing.T) {
	line, original := matchingFixture()
	if got := Match(line, []Operation{original}); got.State != "matched" || got.Operation != original {
		t.Fatalf("complete match: %+v", got)
	}
	foreignNamespace := original
	foreignNamespace.ID, foreignNamespace.NamespaceID = "op_other", "namespace_b"
	if got := Match(line, []Operation{foreignNamespace}); got.State != "awaiting_projection" || got.Reason != "reference_not_found" {
		t.Fatalf("matched equal reference outside its namespace: %+v", got)
	}
	if got := Match(line, []Operation{foreignNamespace, original}); got.State != "matched" || got.Operation != original {
		t.Fatalf("foreign namespace contaminated a unique match: %+v", got)
	}
	for name, edit := range map[string]func(*Operation){
		"amount":     func(o *Operation) { o.AmountMinor++ },
		"currency":   func(o *Operation) { o.Currency = "EUR" },
		"connection": func(o *Operation) { o.ConnectionID = "conn_b" },
		"revision":   func(o *Operation) { o.ConnectionRevision = "revision_b" },
	} {
		t.Run(name, func(t *testing.T) {
			collision := original
			collision.ID = "op_collision"
			edit(&collision)
			for _, snapshots := range [][]Operation{{original, collision}, {collision, original}} {
				before := append([]Operation(nil), snapshots...)
				got := Match(line, snapshots)
				if got.State != "exception" || got.Reason != "ambiguous_reference" {
					t.Fatalf("filtered collision away before uniqueness check: %+v", got)
				}
				if !reflect.DeepEqual(snapshots, before) {
					t.Fatal("matching mutated Payment snapshots")
				}
			}
		})
	}
}

func TestSettlementMatchRejectsMismatchesAndDefersIncompleteLineage(t *testing.T) {
	line, base := matchingFixture()
	cases := []struct {
		name, state, reason string
		edit                func(*Operation)
	}{
		{"connection", "exception", "connection_mismatch", func(o *Operation) { o.ConnectionID = "other" }},
		{"revision", "exception", "connection_mismatch", func(o *Operation) { o.ConnectionRevision = "other" }},
		{"kind", "exception", "unsupported_or_review_required", func(o *Operation) { o.Kind = "refund" }},
		{"review", "exception", "unsupported_or_review_required", func(o *Operation) { o.Integrity = "review_required" }},
		{"currency", "exception", "currency_mismatch", func(o *Operation) { o.Currency = "EUR" }},
		{"amount", "exception", "amount_mismatch", func(o *Operation) { o.AmountMinor = 360 }},
		{"operation lineage", "awaiting_projection", "lineage_incomplete", func(o *Operation) { o.ID = "" }},
		{"intent lineage", "awaiting_projection", "lineage_incomplete", func(o *Operation) { o.IntentID = "" }},
		{"attempt lineage", "awaiting_projection", "lineage_incomplete", func(o *Operation) { o.AttemptID = "" }},
		{"decision lineage", "awaiting_projection", "lineage_incomplete", func(o *Operation) { o.DecisionID = "" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			operation := base
			tc.edit(&operation)
			got := Match(line, []Operation{operation})
			if got.State != tc.state || got.Reason != tc.reason {
				t.Fatalf("match=%+v, want %s/%s", got, tc.state, tc.reason)
			}
		})
	}
}

func TestSettlementFeeAttributionDoesNotGrantCapture(t *testing.T) {
	line, base := matchingFixture()
	for _, status := range []string{"failed", "pending", "unknown", "sending", "created"} {
		t.Run(status, func(t *testing.T) {
			operation := base
			operation.Status = status
			before := operation
			captured := Match(line, []Operation{operation})
			if captured.State != "exception" || captured.Reason != "payment_evidence_requires_reconciliation" {
				t.Fatalf("non-success gained captured principal: %+v", captured)
			}
			feeLine := line
			feeLine.Kind = "fee"
			fee := Match(feeLine, []Operation{operation})
			if fee.State != "matched" || fee.Operation.Status != status || operation != before {
				t.Fatalf("fee attribution rewrote payment state: %+v", fee)
			}
		})
	}
}

func TestSettlementLineContractIsMandatory(t *testing.T) {
	base, operation := matchingFixture()
	for name, edit := range map[string]func(*Line){
		"connection": func(l *Line) { l.ConnectionID = "" },
		"revision":   func(l *Line) { l.ConnectionRevision = "" },
		"namespace":  func(l *Line) { l.NamespaceID = "" },
		"reference":  func(l *Line) { l.ProviderReference = "" },
		"amount":     func(l *Line) { l.BasisAmountMinor = 0 },
		"currency":   func(l *Line) { l.BasisCurrency = "" },
		"kind":       func(l *Line) { l.Kind = "refund" },
	} {
		t.Run(name, func(t *testing.T) {
			line := base
			edit(&line)
			if got := Match(line, []Operation{operation}); got.State != "exception" || got.Reason != "invalid_line_contract" {
				t.Fatalf("invalid line accepted: %+v", got)
			}
		})
	}
}
