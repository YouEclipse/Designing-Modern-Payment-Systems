package telemetry

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestRegistryRejectsUnboundedOrInvalidDimensions(t *testing.T) {
	key := ProviderDimensions{"cand_us_a", "sale", "provider_api_us"}
	for name, keys := range map[string][]ProviderDimensions{
		"empty":      nil,
		"duplicate":  {key, key},
		"request id": {{"pi-sensitive", "sale", "provider_api_us"}},
		"operation":  {{"cand_us_a", "refund", "provider_api_us"}},
		"too many":   make([]ProviderDimensions, 65),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewMetrics(keys); err == nil {
				t.Fatal("invalid metric registry accepted")
			}
		})
	}
}

func TestConcurrentObservationsKeepFiniteSeries(t *testing.T) {
	key := ProviderDimensions{"cand_us_a", "query", "provider_api_us"}
	m, err := NewMetrics([]ProviderDimensions{key})
	if err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for observation := 0; observation < 100; observation++ {
				m.Observe(key, Response)
				m.Observe(key, TechnicalOutcome("issuer_sensitive_decline"))
				m.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/metrics", nil))
			}
		}()
	}
	workers.Wait()
	response := httptest.NewRecorder()
	m.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := response.Body.String()
	if !strings.Contains(body, `outcome="response"} 800`) ||
		!strings.Contains(body, "payment_telemetry_rejected_observations_total 800") ||
		strings.Contains(body, "issuer_sensitive_decline") ||
		strings.Count(body, "\npayment_provider_api_calls_total{") != len(outcomes) {
		t.Fatalf("metric series or totals are not bounded: %s", body)
	}
	post := httptest.NewRecorder()
	m.ServeHTTP(post, httptest.NewRequest(http.MethodPost, "/metrics", nil))
	if post.Code != http.StatusMethodNotAllowed || post.Header().Get("Allow") != http.MethodGet {
		t.Fatal("scrape handler accepted a mutating method")
	}
}
