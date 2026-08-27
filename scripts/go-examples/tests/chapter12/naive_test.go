package naivefallback

import (
	"context"
	"testing"
)

type capturedProvider struct {
	captured      int
	responseError error
}

func (p *capturedProvider) Sale(context.Context, Command) error { p.captured++; return p.responseError }

func TestNaiveFallbackCreatesTwoCapturesAfterLostResponse(t *testing.T) {
	a := &capturedProvider{responseError: context.DeadlineExceeded}
	b := &capturedProvider{}
	if err := Charge(context.Background(), []Provider{a, b}, Command{AmountMinor: 12900, Currency: "USD"}); err != nil {
		t.Fatal(err)
	}
	if a.captured != 1 || b.captured != 1 {
		t.Fatalf("counterexample must reveal two independent effects: A=%d B=%d", a.captured, b.captured)
	}
}
