package naive

import (
	"context"
	"testing"
)

type lostResponseProvider struct{ sales int }

func (p *lostResponseProvider) Sale(context.Context, Command) (Result, error) {
	p.sales++ // The remote side effect happened before the response was lost.
	if p.sales == 1 {
		return Result{}, context.DeadlineExceeded
	}
	return Result{Status: "captured"}, nil
}

func TestBlindRetryDuplicatesACompletedRemoteSale(t *testing.T) {
	p := &lostResponseProvider{}
	got, err := Confirm(context.Background(), p, Command{AmountMinor: 12900, Currency: "USD"})
	if err != nil || got.Status != "captured" || p.sales != 2 {
		t.Fatalf("counterexample result=%+v,%v side effects=%d", got, err, p.sales)
	}
}
