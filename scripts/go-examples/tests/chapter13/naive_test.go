package naivehealth

import (
	"errors"
	"testing"
)

func TestNaiveGlobalBreakerMisclassifiesBusinessDeclines(t *testing.T) {
	b := &Breaker{}
	for i := 0; i < 3; i++ {
		b.Record(errors.New("issuer insufficient_funds"))
	}
	if !b.Open {
		t.Fatal("counterexample should expose generic-error classification")
	}
	b.Record(nil)
	if b.Open || b.Failures != 0 {
		t.Fatal("naive example should show immediate full recovery after one success")
	}
}
