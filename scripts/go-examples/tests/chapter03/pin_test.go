package orchestration

import "testing"

func TestPinRouteBindsExactExecutionIdentity(t *testing.T) {
	selected := SelectedCandidate{CandidateID: "conn_eu", Revision: "connection.v1", NamespaceID: "account_eu", Provider: "psp_a", MerchantAccount: "merchant_eu", Connector: "checkout_v2", Region: "eu"}
	attempt, operation, err := PinRoute("route_1", selected, "att_1", "op_1")
	if err != nil {
		t.Fatal(err)
	}
	if attempt.ProviderConnectionID != selected.CandidateID || operation.ProviderConnectionID != selected.CandidateID || attempt.ConnectionRevision != selected.Revision || operation.ConnectionRevision != selected.Revision || attempt.ProviderNamespaceID != selected.NamespaceID || operation.ProviderNamespaceID != selected.NamespaceID || operation.ProviderIdempotencyKey != operation.OperationID {
		t.Fatalf("identity drift: %+v %+v", attempt, operation)
	}
	for _, mutate := range []func(*SelectedCandidate){func(c *SelectedCandidate) { c.CandidateID = "" }, func(c *SelectedCandidate) { c.Revision = "" }, func(c *SelectedCandidate) { c.NamespaceID = "" }, func(c *SelectedCandidate) { c.Region = "" }} {
		invalid := selected
		mutate(&invalid)
		if _, _, err := PinRoute("route_1", invalid, "att_1", "op_1"); err == nil {
			t.Fatal("accepted incomplete connection")
		}
	}
	if _, _, err := PinRoute("route_1", selected, "", "op_1"); err == nil {
		t.Fatal("accepted no Attempt")
	}
}
