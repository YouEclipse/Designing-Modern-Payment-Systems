package auditboundary

import (
	"bytes"
	"errors"
	"strings"
	"sync"
	"testing"
)

func validAuditProjection() string {
	return `{"decision_id":"decision_1","payment_intent_id":"pi_1",` +
		`"policy_version":"policy_1","outcome":"captured","execution":{` +
		`"connection_id":"connection_a","revision":"revision_1",` +
		`"namespace_id":"namespace_a","provider":"provider_a"}}`
}

func TestAuditBoundaryRejectsEscapedNamesAndNestedDuplicates(t *testing.T) {
	valid := validAuditProjection()
	cases := map[string]string{
		"escaped restricted key": strings.Replace(valid, `"revision":`, `"\u0043VV":"marker_secret","revision":`, 1),
		"escaped duplicate":      strings.Replace(valid, `"revision":`, `"revisio\u006e":"old","revision":`, 1),
		"nested duplicate":       strings.Replace(valid, `"namespace_id":`, `"namespace_id":"other","namespace_id":`, 1),
		"case equivalent name":   strings.Replace(valid, `"connection_id":`, `"CONNECTION_ID":`, 1),
		"unknown nested object":  strings.Replace(valid, `"revision":`, `"debug":{"payment_method_token":"marker_secret"},"revision":`, 1),
		"known leaf object":      strings.Replace(valid, `"revision_1"`, `{"PAN":"marker_secret"}`, 1),
		"missing root field":     strings.Replace(valid, `"outcome":"captured",`, ``, 1),
		"missing nested field":   strings.Replace(valid, `"revision":"revision_1",`, ``, 1),
		"empty leaf":             strings.Replace(valid, `"revision_1"`, `""`, 1),
		"non-string leaf":        strings.Replace(valid, `"revision_1"`, `true`, 1),
		"numeric leaf":           strings.Replace(valid, `"revision_1"`, `123`, 1),
		"root array":             `[` + valid + `]`,
		"root string":            `"marker_secret"`,
		"root null":              `null`,
		"empty object":           `{}`,
		"trailing scalar":        valid + ` true`,
		"empty input":            ``,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			err := ValidateAuditJSON([]byte(raw))
			if !errors.Is(err, ErrBoundary) || strings.Contains(err.Error(), "marker_secret") {
				t.Fatalf("boundary acceptance or input echo: %v", err)
			}
		})
	}
}

func TestAuditBoundaryEnforcesByteLimitsWithoutMutatingInput(t *testing.T) {
	valid := validAuditProjection()
	atLimit := []byte(valid + strings.Repeat(" ", 4096-len(valid)))
	before := bytes.Clone(atLimit)
	if err := ValidateAuditJSON(atLimit); err != nil {
		t.Fatalf("4096-byte valid object: %v", err)
	}
	if !bytes.Equal(atLimit, before) {
		t.Fatal("validator rewrote original evidence")
	}
	if err := ValidateAuditJSON(append(bytes.Clone(atLimit), ' ')); !errors.Is(err, ErrBoundary) {
		t.Fatalf("4097-byte payload accepted: %v", err)
	}
	for _, length := range []int{128, 129} {
		raw := strings.Replace(valid, `"revision_1"`, `"`+strings.Repeat("r", length)+`"`, 1)
		got := ValidateAuditJSON([]byte(raw))
		if (got == nil) != (length == 128) {
			t.Fatalf("leaf byte limit: length=%d err=%v", length, got)
		}
	}
}

func TestAuditShapeValidationDoesNotClaimValueProvenance(t *testing.T) {
	// The chapter explicitly says an allowlisted slot cannot prove that its
	// value came from the authoritative owner. Keep that limitation observable.
	raw := strings.Replace(validAuditProjection(), `"decision_1"`, `"token_wrongly_placed_in_id"`, 1)
	if err := ValidateAuditJSON([]byte(raw)); err != nil {
		t.Fatalf("shape checker unexpectedly claims business provenance validation: %v", err)
	}
}

func TestAuditBoundaryHasNoMutableSharedDecoderState(t *testing.T) {
	valid := []byte(validAuditProjection())
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ValidateAuditJSON(valid); err != nil {
				t.Errorf("concurrent validation: %v", err)
			}
		}()
	}
	wg.Wait()
}
