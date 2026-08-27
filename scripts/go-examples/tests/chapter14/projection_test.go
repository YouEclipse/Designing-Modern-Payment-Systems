package decisionprojection

import (
	"crypto/sha256"
	"errors"
	"testing"
)

func TestAppendDuplicateAndQuarantinePreserveHistoricalIdentity(t *testing.T) {
	original := []byte(`{"decision_id":"rd_1","version":"v1"}`)
	in := Incoming{DecisionID: "rd_1", CanonicalBytes: original, DeclaredHash: sha256.Sum256(original)}
	if got, err := PlanIngest(nil, in); err != nil || got.Action != ActionInsert {
		t.Fatalf("insert=%+v,%v", got, err)
	}
	existing := Existing{DecisionID: in.DecisionID, Hash: in.DeclaredHash}
	before := existing
	for i := 0; i < 20; i++ {
		if got, err := PlanIngest(&existing, in); err != nil || got.Action != ActionDuplicate {
			t.Fatalf("duplicate=%+v,%v", got, err)
		}
	}
	in.CanonicalBytes = []byte(`{"decision_id":"rd_1","version":"v2"}`)
	if got, err := PlanIngest(&existing, in); err != nil || got.Action != ActionQuarantine || got.Reason != "payload_hash_invalid" {
		t.Fatalf("bad hash=%+v,%v", got, err)
	}
	in.DeclaredHash = sha256.Sum256(in.CanonicalBytes)
	if got, err := PlanIngest(&existing, in); err != nil || got.Action != ActionQuarantine || got.Reason != "same_id_different_hash" {
		t.Fatalf("changed history=%+v,%v", got, err)
	}
	if existing != before {
		t.Fatal("ingest overwrote original evidence")
	}
	in.DecisionID = "rd_other"
	if _, err := PlanIngest(&existing, in); !errors.Is(err, ErrMissingIdentity) {
		t.Fatal(err)
	}
	in.DecisionID = ""
	if _, err := PlanIngest(nil, in); !errors.Is(err, ErrMissingIdentity) {
		t.Fatal(err)
	}
}
