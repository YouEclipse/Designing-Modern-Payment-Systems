package capability

import (
	"slices"
	"testing"
	"time"
)

func edgeFixture() (time.Time, Context, Catalog) {
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	return now, Context{AmountMinor: 100, Currency: "USD", TokenDomain: "acct_us"}, Catalog{
		Meta:        SnapshotMeta{CatalogVersion: "v1", Digest: "digest", PublishedAt: now.Add(-time.Hour), EffectiveFrom: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour)},
		Descriptors: []Descriptor{{CandidateID: "a", AccountEnabled: SupportYes, Currencies: []string{"USD"}, AcceptedTokenDomains: []string{"acct_us"}, MinAmountMinor: 1, MaxAmountMinor: 1000}},
	}
}

func TestFreshnessBoundaryAndUnknownNeverBecomeEligible(t *testing.T) {
	for _, name := range []string{"expired", "future", "unknown", "disabled", "empty_currency", "unknown_token", "zero_max"} {
		t.Run(name, func(t *testing.T) {
			now, ctx, catalog := edgeFixture()
			switch name {
			case "expired":
				catalog.Meta.ExpiresAt = now
			case "future":
				catalog.Meta.PublishedAt = now.Add(time.Second)
			case "unknown":
				catalog.Descriptors[0].AccountEnabled = SupportUnknown
			case "disabled":
				catalog.Descriptors[0].AccountEnabled = SupportNo
			case "empty_currency":
				catalog.Descriptors[0].Currencies = nil
			case "unknown_token":
				ctx.TokenDomain = ""
			case "zero_max":
				catalog.Descriptors[0].MaxAmountMinor = 0
			}
			result := FilterSubset(now, ctx, catalog)
			if len(result.Eligible) != 0 || len(result.Exclusions["a"]) == 0 {
				t.Fatalf("failed open: %+v", result)
			}
		})
	}
}

func TestExpiredFactsDoNotAssertCurrentUnsupported(t *testing.T) {
	now, ctx, catalog := edgeFixture()
	catalog.Meta.ExpiresAt = now
	catalog.Descriptors[0].AccountEnabled = SupportNo
	catalog.Descriptors[0].Currencies = []string{"EUR"}
	got := FilterSubset(now, ctx, catalog)
	if len(got.Exclusions["a"]) != 1 || got.Exclusions["a"][0].Code != "capability_snapshot_stale" {
		t.Fatalf("expired facts treated as current: %+v", got)
	}
}

func TestPreferenceDoesNotMutateOrDuplicate(t *testing.T) {
	eligible, preference := []string{"a", "b"}, []string{"blocked", "b", "b"}
	before := slices.Clone(preference)
	if got := ApplyPreference(eligible, preference); !slices.Equal(got, []string{"b", "a"}) {
		t.Fatalf("wrong order: %v", got)
	}
	if !slices.Equal(before, preference) {
		t.Fatal("preference mutated")
	}
}
