package successor

import (
	"errors"
	"testing"
)

func TestSuccessorPreservesBusinessIdentityButRenewsExecutionIdentity(t *testing.T) {
	p := Prior{PaymentID: "pi_1", AttemptID: "attempt_1", AttemptNo: 1, OperationID: "op_1", ProviderKey: "key_1", CandidateID: "a", AmountMinor: 12900, Currency: "USD"}
	ids := NewIDs{AttemptID: "attempt_2", OperationID: "op_2", ProviderKey: "key_2"}
	got, err := Build(p, ids, "b")
	if err != nil || got.PaymentID != p.PaymentID || got.AmountMinor != p.AmountMinor || got.Currency != p.Currency || got.AttemptNo != 2 || got.FallbackFromAttemptID != p.AttemptID || got.AttemptID != ids.AttemptID || got.OperationID != ids.OperationID || got.ProviderKey != ids.ProviderKey || got.CandidateID != "b" {
		t.Fatalf("successor=%+v,%v", got, err)
	}
	for _, bad := range []NewIDs{{p.AttemptID, ids.OperationID, ids.ProviderKey}, {ids.AttemptID, p.OperationID, ids.ProviderKey}, {ids.AttemptID, ids.OperationID, p.ProviderKey}, {"", ids.OperationID, ids.ProviderKey}} {
		if _, err := Build(p, bad, "b"); !errors.Is(err, ErrIdentityReuse) {
			t.Fatalf("identity reuse accepted: %+v,%v", bad, err)
		}
	}
	if _, err := Build(p, ids, "a"); !errors.Is(err, ErrIdentityReuse) {
		t.Fatal("reused candidate")
	}
	for _, change := range []func(*Prior){func(p *Prior) { p.AmountMinor = 0 }, func(p *Prior) { p.CandidateID = "" }, func(p *Prior) { p.AttemptNo = int(^uint(0) >> 1) }} {
		bad := p
		change(&bad)
		if _, err := Build(bad, ids, "b"); !errors.Is(err, ErrIdentityReuse) {
			t.Fatalf("invalid prior accepted: %+v,%v", bad, err)
		}
	}
}
