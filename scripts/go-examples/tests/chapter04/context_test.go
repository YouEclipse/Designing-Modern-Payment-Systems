package routingcontext

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func goldenContext() ContextV1 {
	now := time.Date(2026, 8, 26, 1, 0, 0, 0, time.UTC)
	country, domain := "US", "simulator_a"
	merchantExpiry, metadataExpiry := now.Add(time.Hour), now.Add(5*time.Minute)
	unknown := Observed[string]{State: FieldUnknown, Source: "simulator_metadata", SourceVer: "simulator_metadata.v1", ObservedAt: now}
	return ContextV1{
		SchemaVersion: "routing_context.v1", BuiltAt: now,
		Base: ContextBase{PaymentID: "pi_example", AttemptID: "att_example", MerchantID: "merchant_demo", AmountMinor: 5200,
			Currency: "USD", PaymentMethod: "card", Operation: "sale", CaptureMethod: "automatic"},
		Merchant: MerchantContext{EntityID: "merchant_demo_us", Country: Observed[string]{State: FieldKnown, Value: &country,
			Source: "merchant_config", SourceVer: "merchant_config.v1", ObservedAt: now.Add(-time.Hour), ExpiresAt: &merchantExpiry}},
		PaymentMethod: PaymentMethodContext{Category: "card", Network: unknown, IssuerCountry: unknown,
			CredentialDomain: Observed[string]{State: FieldKnown, Value: &domain, Source: "simulator_metadata",
				SourceVer: "simulator_metadata.v1", ObservedAt: now, ExpiresAt: &metadataExpiry}},
	}
}

func TestDigestMatchesIndependentPaymentAndRouterGolden(t *testing.T) {
	context := goldenContext()
	digest, err := SnapshotDigest(context)
	const expected = "sha256:c82626bdd259b045998a4c1146bb09a33d066f62fa1e0c51cf833a5dd104d274"
	if err != nil || digest != expected {
		t.Fatalf("wire format diverged from code-repository golden: %q %v", digest, err)
	}
	payload, err := json.Marshal(context)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{`"payment_method_token"`, `"pan"`, `"cvv"`, `"value":null`} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("forbidden wire representation: %s", forbidden)
		}
	}
	var restored ContextV1
	if err := json.Unmarshal(payload, &restored); err != nil {
		t.Fatal(err)
	}
	if got, err := SnapshotDigest(restored); err != nil || got != expected {
		t.Fatalf("round trip: %s %v", got, err)
	}
	restored.Base.AmountMinor++
	if got, _ := SnapshotDigest(restored); got == expected {
		t.Fatal("amount mutation did not change digest")
	}
}
