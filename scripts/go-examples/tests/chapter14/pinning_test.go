package pinning

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"reflect"
	"testing"
)

type storeFunc func(context.Context, PrepareCommand) (PreparedOperation, error)

func (f storeFunc) PrepareConfirmation(ctx context.Context, c PrepareCommand) (PreparedOperation, error) {
	return f(ctx, c)
}

type gatewayFunc func(context.Context, PreparedOperation) error

func (f gatewayFunc) Execute(ctx context.Context, op PreparedOperation) error { return f(ctx, op) }

func routedFixture() RoutedDecision {
	payload := []byte(`{"decision_id":"rd_1","attempt_id":"attempt_1","payment_intent_id":"pi_1","selected":{"candidate_id":"conn_eu","provider":"psp_a","merchant_account":"acct_eu","connector":"adapter","region":"eu"}}`)
	return RoutedDecision{DecisionID: "rd_1", AttemptID: "attempt_1", CanonicalBytes: payload, Hash: sha256.Sum256(payload), Selected: CandidateRef{CandidateID: "conn_eu", Provider: "psp_a", MerchantAccount: "acct_eu", Connector: "adapter", Region: "eu"}}
}

func TestGatewayOnlyRunsAfterLocalPrepareAndUsesSealedIdentity(t *testing.T) {
	d := routedFixture()
	original := append([]byte(nil), d.CanonicalBytes...)
	var order []string
	s := Service{
		Store: storeFunc(func(_ context.Context, c PrepareCommand) (PreparedOperation, error) {
			order = append(order, "committed_snapshot_attempt_operation_outbox")
			if c.PaymentIntentID != "pi_1" || c.AttemptID != d.AttemptID || c.DecisionID != d.DecisionID || c.DecisionHash != d.Hash || !bytes.Equal(c.DecisionBytes, d.CanonicalBytes) {
				t.Fatalf("prepare lost binding: %+v", c)
			}
			c.DecisionBytes[0] = 'X'
			return PreparedOperation{OperationID: c.OperationID, ProviderConnectionID: c.ProviderConnectionID, Provider: c.Provider}, nil
		}),
		Gateway: gatewayFunc(func(_ context.Context, op PreparedOperation) error {
			order = append(order, "gateway")
			if op.OperationID != "op_1" || op.ProviderConnectionID != d.Selected.CandidateID {
				t.Fatal(op)
			}
			return nil
		}),
	}
	if err := s.Confirm(context.Background(), "pi_1", "op_1", d); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(order, []string{"committed_snapshot_attempt_operation_outbox", "gateway"}) || !bytes.Equal(d.CanonicalBytes, original) {
		t.Fatalf("ordering/ownership failed: %v", order)
	}
}

func TestHashAndOuterBindingDriftFailBeforePrepare(t *testing.T) {
	for _, change := range []func(*RoutedDecision){
		func(d *RoutedDecision) { d.CanonicalBytes[0] = 'X' }, func(d *RoutedDecision) { d.DecisionID = "rd_other" },
		func(d *RoutedDecision) { d.AttemptID = "attempt_other" }, func(d *RoutedDecision) { d.Selected.CandidateID = "conn_us" },
		func(d *RoutedDecision) { d.Selected.Provider = "other_psp" }, func(d *RoutedDecision) { d.Selected.MerchantAccount = "other_account" },
		func(d *RoutedDecision) { d.Selected.Connector = "other_connector" }, func(d *RoutedDecision) { d.Selected.Region = "other_region" },
	} {
		d := routedFixture()
		change(&d)
		s := Service{Store: storeFunc(func(context.Context, PrepareCommand) (PreparedOperation, error) {
			t.Fatal("invalid decision reached prepare")
			return PreparedOperation{}, nil
		})}
		if err := s.Confirm(context.Background(), "pi_1", "op_1", d); err == nil {
			t.Fatal("drift accepted")
		}
	}
	if err := (Service{}).Confirm(context.Background(), "other_intent", "op_1", routedFixture()); !errors.Is(err, ErrDecisionBindingMismatch) {
		t.Fatalf("intent binding=%v", err)
	}
}

func TestPrepareFailureOrReturnedIdentityDriftNeverCallsGateway(t *testing.T) {
	for _, tc := range []struct {
		op  PreparedOperation
		err error
	}{
		{err: errors.New("local transaction failed")},
		{op: PreparedOperation{OperationID: "op_1", ProviderConnectionID: "other_connection", Provider: "psp_a"}},
		{op: PreparedOperation{OperationID: "other_operation", ProviderConnectionID: "conn_eu", Provider: "psp_a"}},
		{op: PreparedOperation{OperationID: "op_1", ProviderConnectionID: "conn_eu", Provider: "other_provider"}},
	} {
		s := Service{Store: storeFunc(func(context.Context, PrepareCommand) (PreparedOperation, error) { return tc.op, tc.err }), Gateway: gatewayFunc(func(context.Context, PreparedOperation) error {
			t.Fatal("Gateway called without valid local authority")
			return nil
		})}
		if err := s.Confirm(context.Background(), "pi_1", "op_1", routedFixture()); err == nil {
			t.Fatal("unsafe prepared operation accepted")
		}
	}
}
