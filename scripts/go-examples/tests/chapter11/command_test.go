package command

import "testing"

func TestFingerprintUsesCanonicalSemanticCommandAndTarget(t *testing.T) {
	base := ConfirmPayment{SchemaVersion: "confirm.v1", MerchantID: "merchant_1", IntentID: "intent_1", CaptureMethod: "automatic"}
	want, err := FingerprintConfirm(base)
	if err != nil {
		t.Fatal(err)
	}
	alias := base
	alias.MerchantID, alias.IntentID, alias.CaptureMethod = " merchant_1 ", " intent_1 ", " AUTOMATIC "
	if got, err := FingerprintConfirm(alias); err != nil || got != want {
		t.Fatalf("canonical alias changed semantic digest: %x,%v", got, err)
	}
	for _, change := range []func(*ConfirmPayment){
		func(c *ConfirmPayment) { c.SchemaVersion = "confirm.v2" }, func(c *ConfirmPayment) { c.MerchantID = "merchant_2" },
		func(c *ConfirmPayment) { c.IntentID = "intent_2" }, func(c *ConfirmPayment) { c.CaptureMethod = "manual" },
	} {
		other := base
		change(&other)
		if got, err := FingerprintConfirm(other); err != nil || got == want {
			t.Fatalf("changed command reused fingerprint: %+v,%v", other, err)
		}
	}
	for _, change := range []func(*ConfirmPayment){func(c *ConfirmPayment) { c.SchemaVersion = "" }, func(c *ConfirmPayment) { c.MerchantID = " " }, func(c *ConfirmPayment) { c.IntentID = "" }} {
		bad := base
		change(&bad)
		if _, err := FingerprintConfirm(bad); err == nil {
			t.Fatalf("accepted incomplete subject %+v", bad)
		}
	}
}
