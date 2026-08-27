package routing

import "testing"

func TestLegacyEmptyCapabilityIsNotWildcard(t *testing.T) {
	reasons := eligibilityReasons(Request{Currency: "USD", PaymentMethod: "card", Operation: "sale"}, Candidate{})
	if len(reasons) != 4 {
		t.Fatalf("lost hard failures: %v", reasons)
	}
	candidate := Candidate{Enabled: true, SupportedCurrencies: []string{"USD"}, SupportedPaymentMethods: []string{"card"}, SupportedOperations: []string{"sale"}}
	if got := eligibilityReasons(Request{Currency: "USD", PaymentMethod: "card", Operation: "sale"}, candidate); len(got) != 0 {
		t.Fatalf("matching candidate excluded: %v", got)
	}
}
