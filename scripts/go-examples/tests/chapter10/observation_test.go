package observation

import "testing"

func TestObservationCertaintyIsNotAnHTTPErrorGuess(t *testing.T) {
	for _, tc := range []struct {
		observation ProviderObservation
		class       OutcomeClass
		certainty   SideEffectCertainty
	}{
		{ProviderObservation{Verdict: VerdictCaptured, ProviderReference: "ref_1"}, OutcomeSuccess, EffectConfirmed},
		{ProviderObservation{Verdict: VerdictCaptured}, OutcomeUnknown, EffectPossible},
		{ProviderObservation{Verdict: VerdictCaptured, DispatchProof: DispatchNever}, OutcomeUnknown, EffectPossible},
		{ProviderObservation{Verdict: VerdictDeclinedNoEffect, NormalizedCode: "issuer_decline"}, OutcomeFailure, EffectNone},
		{ProviderObservation{Verdict: VerdictPending}, OutcomePending, EffectPossible},
		{ProviderObservation{DispatchProof: DispatchNever}, OutcomeFailure, EffectNone},
		{ProviderObservation{DispatchProof: DispatchPossible, NormalizedCode: "HTTP_500"}, OutcomeUnknown, EffectPossible},
		{ProviderObservation{Verdict: "authorized", ProviderReference: "ref_1"}, OutcomeUnknown, EffectPossible},
		{ProviderObservation{Verdict: "timeout"}, OutcomeUnknown, EffectPossible},
		{ProviderObservation{}, OutcomeUnknown, EffectPossible},
	} {
		got := Normalize(tc.observation)
		if got.Class != tc.class || got.Certainty != tc.certainty {
			t.Fatalf("Normalize(%+v)=%+v", tc.observation, got)
		}
	}
}
