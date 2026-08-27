package settlementfact

import (
	"crypto/sha256"
	"testing"
)

func TestSettlementFactClassifiesIdentityBeforePayload(t *testing.T) {
	base := Fact{Key: Key{"namespace_a", "entry_1", "revision_1"}, PayloadHash: sha256.Sum256([]byte("normalized-fee-v1"))}
	if got, err := Classify(nil, base); err != nil || got != Append {
		t.Fatalf("new fact: action=%q err=%v", got, err)
	}
	existing := base
	if got, err := Classify(&existing, base); err != nil || got != Replay {
		t.Fatalf("exact replay: action=%q err=%v", got, err)
	}
	changed := base
	changed.PayloadHash = sha256.Sum256([]byte("changed-fee-v1"))
	if got, err := Classify(&existing, changed); err != nil || got != Quarantine {
		t.Fatalf("changed payload must quarantine: action=%q err=%v", got, err)
	}
	for name, edit := range map[string]func(*Fact){
		"namespace": func(f *Fact) { f.Key.ProviderNamespaceID = "namespace_b" },
		"entry":     func(f *Fact) { f.Key.EntryID = "entry_2" },
		"revision":  func(f *Fact) { f.Key.Revision = "revision_2" },
	} {
		t.Run(name, func(t *testing.T) {
			incoming := base
			edit(&incoming)
			if _, err := Classify(&existing, incoming); err == nil {
				t.Fatal("lookup from another identity was accepted as replay")
			}
			if action, err := Classify(nil, incoming); err != nil || action != Append {
				t.Fatalf("distinct identity cannot append independently: %q %v", action, err)
			}
		})
	}
	if existing != base {
		t.Fatal("classification overwrote existing immutable evidence")
	}
}

func TestSettlementFactRejectsIncompleteIncomingIdentity(t *testing.T) {
	base := Fact{Key: Key{"namespace_a", "entry_1", "revision_1"}, PayloadHash: sha256.Sum256([]byte("normalized-fee-v1"))}
	for name, edit := range map[string]func(*Fact){
		"namespace": func(f *Fact) { f.Key.ProviderNamespaceID = "" },
		"entry":     func(f *Fact) { f.Key.EntryID = "" },
		"revision":  func(f *Fact) { f.Key.Revision = "" },
		"digest":    func(f *Fact) { f.PayloadHash = [32]byte{} },
	} {
		t.Run(name, func(t *testing.T) {
			incoming := base
			edit(&incoming)
			for _, existing := range []*Fact{nil, &base} {
				if _, err := Classify(existing, incoming); err == nil {
					t.Fatal("incomplete fact accepted")
				}
			}
		})
	}
}
